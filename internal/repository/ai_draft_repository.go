package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"requirement-splitting/internal/domain"
)

type AIDraftRepository interface {
	CreateAIDraft(ctx context.Context, draft domain.AIDraft) (domain.AIDraft, error)
	ListAIDrafts(ctx context.Context, projectID string) ([]domain.AIDraft, error)
}

type PGAIDraftRepository struct {
	pool *pgxpool.Pool
}

func NewPGAIDraftRepository(pool *pgxpool.Pool) *PGAIDraftRepository {
	return &PGAIDraftRepository{pool: pool}
}

func (r *PGAIDraftRepository) CreateAIDraft(ctx context.Context, draft domain.AIDraft) (domain.AIDraft, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO ai_drafts (project_id, requirement_id, task_type, provider, model, input, output, validation_errors, status, created_by)
		VALUES ($1, NULLIF($2, '')::uuid, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, '')::uuid)
		RETURNING id::text, project_id::text, COALESCE(requirement_id::text, ''), task_type, provider, model, input::text, output::text, validation_errors::text, status, COALESCE(created_by::text, ''), published_at, created_at, updated_at
	`, draft.ProjectID, draft.RequirementID, draft.TaskType, draft.Provider, draft.Model, draft.InputJSON, draft.OutputJSON, draft.ValidationErrors, draft.Status, draft.CreatedBy)

	var created domain.AIDraft
	var inputJSON string
	var outputJSON string
	var validationErrors string
	if err := row.Scan(&created.ID, &created.ProjectID, &created.RequirementID, &created.TaskType, &created.Provider, &created.Model, &inputJSON, &outputJSON, &validationErrors, &created.Status, &created.CreatedBy, &created.PublishedAt, &created.CreatedAt, &created.UpdatedAt); err != nil {
		if isForeignKeyViolation(err) {
			return domain.AIDraft{}, ErrNotFound
		}
		return domain.AIDraft{}, err
	}
	created.InputJSON = []byte(inputJSON)
	created.OutputJSON = []byte(outputJSON)
	created.ValidationErrors = []byte(validationErrors)
	return created, nil
}

func (r *PGAIDraftRepository) ListAIDrafts(ctx context.Context, projectID string) ([]domain.AIDraft, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, project_id::text, COALESCE(requirement_id::text, ''), task_type, provider, model, input::text, output::text, validation_errors::text, status, COALESCE(created_by::text, ''), published_at, created_at, updated_at
		FROM ai_drafts
		WHERE project_id = $1
		ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drafts []domain.AIDraft
	for rows.Next() {
		var draft domain.AIDraft
		var inputJSON string
		var outputJSON string
		var validationErrors string
		if err := rows.Scan(&draft.ID, &draft.ProjectID, &draft.RequirementID, &draft.TaskType, &draft.Provider, &draft.Model, &inputJSON, &outputJSON, &validationErrors, &draft.Status, &draft.CreatedBy, &draft.PublishedAt, &draft.CreatedAt, &draft.UpdatedAt); err != nil {
			return nil, err
		}
		draft.InputJSON = []byte(inputJSON)
		draft.OutputJSON = []byte(outputJSON)
		draft.ValidationErrors = []byte(validationErrors)
		drafts = append(drafts, draft)
	}
	return drafts, rows.Err()
}
