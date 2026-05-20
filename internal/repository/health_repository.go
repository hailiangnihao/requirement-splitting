package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"requirement-splitting/internal/domain"
)

type HealthRepository interface {
	GetMetrics(ctx context.Context, projectID string) (domain.HealthMetrics, error)
}

type pgHealthRepository struct {
	pool *pgxpool.Pool
}

func NewPGHealthRepository(pool *pgxpool.Pool) HealthRepository {
	return &pgHealthRepository{pool: pool}
}

func (r *pgHealthRepository) GetMetrics(ctx context.Context, projectID string) (domain.HealthMetrics, error) {
	query := `
		SELECT 
			(SELECT COUNT(*) FROM feature_points WHERE project_id = $1) as fp_count,
			(SELECT COUNT(*) FROM feature_points fp WHERE project_id = $1 AND NOT EXISTS (SELECT 1 FROM test_cases tc WHERE tc.feature_point_id = fp.id)) as untested_fp_count,
			(SELECT COUNT(*) FROM dev_tasks WHERE project_id = $1) as task_total,
			(SELECT COUNT(*) FROM dev_tasks WHERE project_id = $1 AND status = 'done') as task_done,
			(SELECT COUNT(*) FROM defects WHERE project_id = $1 AND status NOT IN ('closed', 'rejected')) as active_defects,
			(SELECT COUNT(*) FROM change_requests WHERE project_id = $1 AND created_at >= NOW() - INTERVAL '7 days') as recent_changes
	`

	var metrics domain.HealthMetrics
	err := r.pool.QueryRow(ctx, query, projectID).Scan(
		&metrics.FeaturePointCount,
		&metrics.UntestedFeatureCount,
		&metrics.DevTaskTotal,
		&metrics.DevTaskDone,
		&metrics.ActiveDefects,
		&metrics.RecentChanges,
	)

	return metrics, err
}
