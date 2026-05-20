package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"requirement-splitting/internal/domain"
)

type DefectRepository interface {
	CreateDefect(ctx context.Context, defect domain.Defect) (domain.Defect, error)
	UpdateDefectStatus(ctx context.Context, projectID, defectID string, status domain.DefectStatus) error
	GetDefect(ctx context.Context, defectID string) (domain.Defect, error)
}

type pgDefectRepository struct {
	pool *pgxpool.Pool
}

func NewPGDefectRepository(pool *pgxpool.Pool) DefectRepository {
	return &pgDefectRepository{pool: pool}
}

func (r *pgDefectRepository) CreateDefect(ctx context.Context, defect domain.Defect) (domain.Defect, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO defects (id, project_id, title, description, status, test_run_id, created_by, assigned_to, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, '')::uuid, NULLIF($8, '')::uuid, $9, $10)
		RETURNING id::text, project_id::text, title, description, status, test_run_id::text, created_by::text, assigned_to::text, created_at, updated_at
	`, defect.ID, defect.ProjectID, defect.Title, defect.Description, defect.Status, defect.TestRunID, defect.CreatedBy, defect.AssignedTo, defect.CreatedAt, defect.UpdatedAt)

	var created domain.Defect
	var testRunID, createdBy, assignedTo *string // 处理可能返回 NULL 的字段

	if err := row.Scan(&created.ID, &created.ProjectID, &created.Title, &created.Description, &created.Status, &testRunID, &createdBy, &assignedTo, &created.CreatedAt, &created.UpdatedAt); err != nil {
		return domain.Defect{}, err
	}

	created.TestRunID = testRunID
	if createdBy != nil {
		created.CreatedBy = *createdBy
	}
	if assignedTo != nil {
		created.AssignedTo = *assignedTo
	}
	return created, nil
}

func (r *pgDefectRepository) UpdateDefectStatus(ctx context.Context, projectID, defectID string, status domain.DefectStatus) error {
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE defects 
		SET status = $1, updated_at = NOW() 
		WHERE id = $2 AND project_id = $3
	`, string(status), defectID, projectID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgDefectRepository) GetDefect(ctx context.Context, defectID string) (domain.Defect, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id::text, project_id::text, title, description, status, test_run_id::text, created_by::text, assigned_to::text, created_at, updated_at
		FROM defects
		WHERE id = $1
	`, defectID)

	var defect domain.Defect
	var testRunID, createdBy, assignedTo *string

	if err := row.Scan(&defect.ID, &defect.ProjectID, &defect.Title, &defect.Description, &defect.Status, &testRunID, &createdBy, &assignedTo, &defect.CreatedAt, &defect.UpdatedAt); err != nil {
		return domain.Defect{}, err
	}

	defect.TestRunID = testRunID
	if createdBy != nil {
		defect.CreatedBy = *createdBy
	}
	if assignedTo != nil {
		defect.AssignedTo = *assignedTo
	}
	return defect, nil
}
