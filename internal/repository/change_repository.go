package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"requirement-splitting/internal/domain"
)

type ChangeRepository interface {
	CreateChangeRequest(ctx context.Context, cr domain.ChangeRequest) (domain.ChangeRequest, error)
	GetChangeRequest(ctx context.Context, id string) (domain.ChangeRequest, error)
	ListChangeRequests(ctx context.Context, projectID string) ([]domain.ChangeRequest, error)
	UpdateChangeAnalysis(ctx context.Context, projectID, id string, analysis []byte, status domain.ChangeRequestStatus) error
	UpdateChangeStatus(ctx context.Context, projectID, id string, status domain.ChangeRequestStatus) error
}

type pgChangeRepository struct {
	pool *pgxpool.Pool
}

func NewPGChangeRepository(pool *pgxpool.Pool) ChangeRepository {
	return &pgChangeRepository{pool: pool}
}

func (r *pgChangeRepository) CreateChangeRequest(ctx context.Context, cr domain.ChangeRequest) (domain.ChangeRequest, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO change_requests (id, project_id, title, content, status, impact_analysis, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, '')::uuid, $8, $9)
		RETURNING id::text, project_id::text, title, content, status, impact_analysis, created_by::text, created_at, updated_at
	`, cr.ID, cr.ProjectID, cr.Title, cr.Content, cr.Status, cr.ImpactAnalysis, cr.CreatedBy, cr.CreatedAt, cr.UpdatedAt)

	var created domain.ChangeRequest
	var createdBy *string
	var analysisBytes []byte

	if err := row.Scan(&created.ID, &created.ProjectID, &created.Title, &created.Content, &created.Status, &analysisBytes, &createdBy, &created.CreatedAt, &created.UpdatedAt); err != nil {
		return domain.ChangeRequest{}, err
	}

	created.ImpactAnalysis = analysisBytes
	if createdBy != nil {
		created.CreatedBy = *createdBy
	}
	return created, nil
}

func (r *pgChangeRepository) GetChangeRequest(ctx context.Context, id string) (domain.ChangeRequest, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id::text, project_id::text, title, content, status, impact_analysis, created_by::text, created_at, updated_at
		FROM change_requests WHERE id = $1
	`, id)

	var cr domain.ChangeRequest
	var createdBy *string
	var analysisBytes []byte

	if err := row.Scan(&cr.ID, &cr.ProjectID, &cr.Title, &cr.Content, &cr.Status, &analysisBytes, &createdBy, &cr.CreatedAt, &cr.UpdatedAt); err != nil {
		return domain.ChangeRequest{}, err
	}
	cr.ImpactAnalysis = analysisBytes
	if createdBy != nil {
		cr.CreatedBy = *createdBy
	}
	return cr, nil
}

func (r *pgChangeRepository) UpdateChangeAnalysis(ctx context.Context, projectID, id string, analysis []byte, status domain.ChangeRequestStatus) error {
	cmdTag, err := r.pool.Exec(ctx, "UPDATE change_requests SET impact_analysis = $1, status = $2, updated_at = NOW() WHERE id = $3 AND project_id = $4", analysis, string(status), id, projectID)
	if err == nil && cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

func (r *pgChangeRepository) ListChangeRequests(ctx context.Context, projectID string) ([]domain.ChangeRequest, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, project_id::text, title, content, status, impact_analysis, created_by::text, created_at, updated_at
		FROM change_requests
		WHERE project_id = $1
		ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	changes := make([]domain.ChangeRequest, 0)
	for rows.Next() {
		var cr domain.ChangeRequest
		var createdBy *string
		var analysisBytes []byte
		if err := rows.Scan(&cr.ID, &cr.ProjectID, &cr.Title, &cr.Content, &cr.Status, &analysisBytes, &createdBy, &cr.CreatedAt, &cr.UpdatedAt); err != nil {
			return nil, err
		}
		cr.ImpactAnalysis = analysisBytes
		if createdBy != nil {
			cr.CreatedBy = *createdBy
		}
		changes = append(changes, cr)
	}
	return changes, rows.Err()
}

func (r *pgChangeRepository) UpdateChangeStatus(ctx context.Context, projectID, id string, status domain.ChangeRequestStatus) error {
	cmdTag, err := r.pool.Exec(ctx, "UPDATE change_requests SET status = $1, updated_at = NOW() WHERE id = $2 AND project_id = $3", string(status), id, projectID)
	if err == nil && cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}
