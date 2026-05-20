package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"requirement-splitting/internal/domain"
)

type TestRepository interface {
	CreateTestRun(ctx context.Context, run domain.TestRun) (domain.TestRun, error)
	ConfirmTestCase(ctx context.Context, projectID, testCaseID string) error
	ReviewTestRun(ctx context.Context, projectID, testRunID string, status domain.TestRunReviewStatus) error
	GetTestRun(ctx context.Context, testRunID string) (domain.TestRun, error)
}

type pgTestRepository struct {
	pool *pgxpool.Pool
}

func NewPGTestRepository(pool *pgxpool.Pool) TestRepository {
	return &pgTestRepository{pool: pool}
}

func (r *pgTestRepository) CreateTestRun(ctx context.Context, run domain.TestRun) (domain.TestRun, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO test_runs (id, project_id, test_case_id, executed_by, execution_type, actual_result, evidence, is_defect_suggested, review_status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id::text, project_id::text, test_case_id::text, executed_by, execution_type, actual_result, evidence, is_defect_suggested, review_status, created_at
	`, run.ID, run.ProjectID, run.TestCaseID, run.ExecutedBy, run.ExecutionType, run.ActualResult, run.Evidence, run.IsDefectSuggested, run.ReviewStatus, run.CreatedAt)

	var created domain.TestRun
	var evidenceBytes []byte
	if err := row.Scan(&created.ID, &created.ProjectID, &created.TestCaseID, &created.ExecutedBy, &created.ExecutionType, &created.ActualResult, &evidenceBytes, &created.IsDefectSuggested, &created.ReviewStatus, &created.CreatedAt); err != nil {
		return domain.TestRun{}, err
	}

	created.Evidence = evidenceBytes
	return created, nil
}

func (r *pgTestRepository) ConfirmTestCase(ctx context.Context, projectID, testCaseID string) error {
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE test_cases 
		SET confirmation_status = $1 
		WHERE id = $2 AND project_id = $3
	`, string(domain.TestCaseConfirmationConfirmed), testCaseID, projectID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgTestRepository) GetTestRun(ctx context.Context, testRunID string) (domain.TestRun, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, project_id, test_case_id, executed_by, execution_type, actual_result, evidence, is_defect_suggested, review_status, created_at
		FROM test_runs WHERE id = $1
	`, testRunID)

	var run domain.TestRun
	var evidenceBytes []byte
	if err := row.Scan(&run.ID, &run.ProjectID, &run.TestCaseID, &run.ExecutedBy, &run.ExecutionType, &run.ActualResult, &evidenceBytes, &run.IsDefectSuggested, &run.ReviewStatus, &run.CreatedAt); err != nil {
		return domain.TestRun{}, err
	}
	run.Evidence = evidenceBytes
	return run, nil
}

func (r *pgTestRepository) ReviewTestRun(ctx context.Context, projectID, testRunID string, status domain.TestRunReviewStatus) error {
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE test_runs 
		SET review_status = $1 
		WHERE id = $2 AND project_id = $3
	`, string(status), testRunID, projectID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
