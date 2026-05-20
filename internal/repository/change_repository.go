package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"requirement-splitting/internal/domain"
)

type ChangeRepository interface {
	CreateChangeRequest(ctx context.Context, cr domain.ChangeRequest) (domain.ChangeRequest, error)
	GetChangeRequest(ctx context.Context, id string) (domain.ChangeRequest, error)
	UpdateChangeAnalysis(ctx context.Context, projectID, id string, analysis []byte, status domain.ChangeRequestStatus) error
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
