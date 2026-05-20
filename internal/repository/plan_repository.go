package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"requirement-splitting/internal/domain"
)

type PlanRepository interface {
	// PublishPlan 将整个正式计划作为一个数据库事务插入
	PublishPlan(ctx context.Context, plan *domain.FormalPlan) error
	// GetPlan 根据项目 ID 获取该项目下所有的正式计划数据（扁平结构）
	GetPlan(ctx context.Context, projectID string) (*domain.FormalPlan, error)
}

type pgPlanRepository struct {
	pool *pgxpool.Pool
}

func NewPGPlanRepository(pool *pgxpool.Pool) PlanRepository {
	return &pgPlanRepository{pool: pool}
}

func (r *pgPlanRepository) PublishPlan(ctx context.Context, plan *domain.FormalPlan) error {
	// 1. 开启数据库事务
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	// 确保在函数退出时，如果事务没有提交，则回滚
	defer tx.Rollback(ctx)

	// 2. 批量插入 Modules
	for _, mod := range plan.Modules {
		_, err := tx.Exec(ctx,
			"INSERT INTO modules (id, project_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
			mod.ID, mod.ProjectID, mod.Name, mod.Description, mod.CreatedAt)
		if err != nil {
			return fmt.Errorf("insert module %s: %w", mod.ID, err)
		}
	}

	// 3. 批量插入 Milestones
	for _, ms := range plan.Milestones {
		_, err := tx.Exec(ctx,
			"INSERT INTO milestones (id, project_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
			ms.ID, ms.ProjectID, ms.Name, ms.Description, ms.CreatedAt)
		if err != nil {
			return fmt.Errorf("insert milestone %s: %w", ms.ID, err)
		}
	}

	// 4. 批量插入 Feature Points
	for _, fp := range plan.FeaturePoints {
		_, err := tx.Exec(ctx,
			"INSERT INTO feature_points (id, project_id, module_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
			fp.ID, fp.ProjectID, fp.ModuleID, fp.Name, fp.Description, fp.CreatedAt)
		if err != nil {
			return fmt.Errorf("insert feature point %s: %w", fp.ID, err)
		}
	}

	// 5. 批量插入 Dev Tasks
	for _, task := range plan.DevTasks {
		_, err := tx.Exec(ctx,
			"INSERT INTO dev_tasks (id, project_id, feature_point_id, name, status, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
			task.ID, task.ProjectID, task.FeaturePointID, task.Name, task.Status, task.CreatedAt)
		if err != nil {
			return fmt.Errorf("insert dev task %s: %w", task.ID, err)
		}
	}

	// 6. 批量插入 Test Cases (注意：状态默认为待人工确认)
	for _, tc := range plan.TestCases {
		_, err := tx.Exec(ctx,
			"INSERT INTO test_cases (id, project_id, feature_point_id, title, confirmation_status, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
			tc.ID, tc.ProjectID, tc.FeaturePointID, tc.Title, tc.ConfirmationStatus, tc.CreatedAt)
		if err != nil {
			return fmt.Errorf("insert test case %s: %w", tc.ID, err)
		}
	}

	// 7. 批量插入 Acceptance Items
	for _, item := range plan.AcceptanceItems {
		_, err := tx.Exec(ctx,
			"INSERT INTO acceptance_items (id, project_id, description, created_at) VALUES ($1, $2, $3, $4)",
			item.ID, item.ProjectID, item.Description, item.CreatedAt)
		if err != nil {
			return fmt.Errorf("insert acceptance item %s: %w", item.ID, err)
		}
	}

	// 8. 更新 AI 草稿状态，标记为已发布
	// 假设 ai_drafts 表有一个 status 字段，从 "draft" 变为 "published"
	_, err = tx.Exec(ctx,
		"UPDATE ai_drafts SET status = 'published' WHERE id = $1 AND project_id = $2",
		plan.DraftID, plan.ProjectID)
	if err != nil {
		return fmt.Errorf("update ai draft status: %w", err)
	}

	// 9. 提交事务
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *pgPlanRepository) GetPlan(ctx context.Context, projectID string) (*domain.FormalPlan, error) {
	plan := &domain.FormalPlan{ProjectID: projectID}

	// 1. 获取 Modules
	modRows, err := r.pool.Query(ctx, "SELECT id, project_id, name, description, created_at FROM modules WHERE project_id = $1", projectID)
	if err != nil {
		return nil, fmt.Errorf("query modules: %w", err)
	}
	for modRows.Next() {
		var m domain.Module
		if err := modRows.Scan(&m.ID, &m.ProjectID, &m.Name, &m.Description, &m.CreatedAt); err != nil {
			modRows.Close()
			return nil, err
		}
		plan.Modules = append(plan.Modules, m)
	}
	modRows.Close()

	// 2. 获取 Milestones
	msRows, err := r.pool.Query(ctx, "SELECT id, project_id, name, description, created_at FROM milestones WHERE project_id = $1", projectID)
	if err != nil {
		return nil, fmt.Errorf("query milestones: %w", err)
	}
	for msRows.Next() {
		var ms domain.Milestone
		if err := msRows.Scan(&ms.ID, &ms.ProjectID, &ms.Name, &ms.Description, &ms.CreatedAt); err != nil {
			msRows.Close()
			return nil, err
		}
		plan.Milestones = append(plan.Milestones, ms)
	}
	msRows.Close()

	// 3. 获取 Feature Points
	fpRows, err := r.pool.Query(ctx, "SELECT id, project_id, module_id, name, description, created_at FROM feature_points WHERE project_id = $1", projectID)
	if err != nil {
		return nil, fmt.Errorf("query feature points: %w", err)
	}
	for fpRows.Next() {
		var fp domain.FeaturePoint
		if err := fpRows.Scan(&fp.ID, &fp.ProjectID, &fp.ModuleID, &fp.Name, &fp.Description, &fp.CreatedAt); err != nil {
			fpRows.Close()
			return nil, err
		}
		plan.FeaturePoints = append(plan.FeaturePoints, fp)
	}
	fpRows.Close()

	// 4. 获取 Dev Tasks
	taskRows, err := r.pool.Query(ctx, "SELECT id, project_id, feature_point_id, name, status, created_at FROM dev_tasks WHERE project_id = $1", projectID)
	if err != nil {
		return nil, fmt.Errorf("query dev tasks: %w", err)
	}
	for taskRows.Next() {
		var dt domain.DevTask
		if err := taskRows.Scan(&dt.ID, &dt.ProjectID, &dt.FeaturePointID, &dt.Name, &dt.Status, &dt.CreatedAt); err != nil {
			taskRows.Close()
			return nil, err
		}
		plan.DevTasks = append(plan.DevTasks, dt)
	}
	taskRows.Close()

	// 5. 获取 Test Cases
	tcRows, err := r.pool.Query(ctx, "SELECT id, project_id, feature_point_id, title, confirmation_status, created_at FROM test_cases WHERE project_id = $1", projectID)
	if err != nil {
		return nil, fmt.Errorf("query test cases: %w", err)
	}
	for tcRows.Next() {
		var tc domain.TestCase
		if err := tcRows.Scan(&tc.ID, &tc.ProjectID, &tc.FeaturePointID, &tc.Title, &tc.ConfirmationStatus, &tc.CreatedAt); err != nil {
			tcRows.Close()
			return nil, err
		}
		plan.TestCases = append(plan.TestCases, tc)
	}
	tcRows.Close()

	// 6. 获取 Acceptance Items
	accRows, err := r.pool.Query(ctx, "SELECT id, project_id, description, created_at FROM acceptance_items WHERE project_id = $1", projectID)
	if err != nil {
		return nil, fmt.Errorf("query acceptance items: %w", err)
	}
	for accRows.Next() {
		var ai domain.AcceptanceItem
		if err := accRows.Scan(&ai.ID, &ai.ProjectID, &ai.Description, &ai.CreatedAt); err != nil {
			accRows.Close()
			return nil, err
		}
		plan.AcceptanceItems = append(plan.AcceptanceItems, ai)
	}
	accRows.Close()

	return plan, nil
}
