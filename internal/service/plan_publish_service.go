package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"requirement-splitting/internal/ai"
	"requirement-splitting/internal/domain"
	"requirement-splitting/internal/repository"
)

type PlanPublishService struct {
	draftRepo repository.AIDraftRepository
	planRepo  repository.PlanRepository
}

func NewPlanPublishService(draftRepo repository.AIDraftRepository, planRepo repository.PlanRepository) *PlanPublishService {
	return &PlanPublishService{
		draftRepo: draftRepo,
		planRepo:  planRepo,
	}
}

// PublishDraft 将特定的 AI 草稿解析并发布为正式计划
func (s *PlanPublishService) PublishDraft(ctx context.Context, projectID, draftID string) error {
	// 1. 获取草稿 (因为当前接口没有 GetByID，我们先从 List 里过滤)
	drafts, err := s.draftRepo.ListAIDrafts(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to list drafts: %w", err)
	}

	var targetDraft *domain.AIDraft
	for _, d := range drafts {
		if d.ID == draftID {
			targetDraft = &d
			break
		}
	}

	if targetDraft == nil {
		return errors.New("draft not found")
	}

	// 可选：校验草稿状态，只有特定状态（如 validated）才能发布
	// if targetDraft.Status != domain.AIDraftStatusValidated { ... }

	// 2. 解析 JSON
	var outData ai.SplitRequirementResult
	if err := json.Unmarshal(targetDraft.OutputJSON, &outData); err != nil {
		return fmt.Errorf("failed to parse draft output JSON: %w", err)
	}

	// 3. 组装 FormalPlan 领域模型
	now := time.Now()
	plan := &domain.FormalPlan{
		ProjectID: projectID,
		DraftID:   draftID,
	}

	moduleIDsByKey := make(map[string]string)
	featurePointIDsByKey := make(map[string]string)

	for _, m := range outData.Modules {
		modID := generateID()
		moduleIDsByKey[m.Key] = modID
		plan.Modules = append(plan.Modules, domain.Module{
			ID: modID, ProjectID: projectID, Name: m.Name, Description: m.Description, CreatedAt: now,
		})
	}

	for _, ms := range outData.Milestones {
		plan.Milestones = append(plan.Milestones, domain.Milestone{
			ID: generateID(), ProjectID: projectID, Name: ms.Name, Description: ms.Description, CreatedAt: now,
		})
	}

	for _, fp := range outData.FeaturePoints {
		fpID := generateID()
		featurePointIDsByKey[fp.Key] = fpID
		plan.FeaturePoints = append(plan.FeaturePoints, domain.FeaturePoint{
			ID:          fpID,
			ProjectID:   projectID,
			ModuleID:    moduleIDsByKey[fp.ModuleKey],
			Name:        fp.Title,
			Description: fp.Description,
			CreatedAt:   now,
		})
	}

	for _, task := range outData.DevTasks {
		plan.DevTasks = append(plan.DevTasks, domain.DevTask{
			ID:             generateID(),
			ProjectID:      projectID,
			FeaturePointID: featurePointIDsByKey[task.FeaturePointKey],
			Name:           task.Title,
			Status:         string(domain.TaskStatusPendingDev),
			CreatedAt:      now,
		})
	}

	for _, tc := range outData.TestCases {
		plan.TestCases = append(plan.TestCases, domain.TestCase{
			ID:                 generateID(),
			ProjectID:          projectID,
			FeaturePointID:     featurePointIDsByKey[tc.FeaturePointKey],
			Title:              tc.Title,
			ConfirmationStatus: string(domain.TestCaseConfirmationPending),
			CreatedAt:          now,
		})
	}

	for _, acc := range outData.AcceptanceItems {
		plan.AcceptanceItems = append(plan.AcceptanceItems, domain.AcceptanceItem{
			ID: generateID(), ProjectID: projectID, Description: acc.PassCriteria, CreatedAt: now,
		})
	}

	// 4. 调用仓储层事务，一并落库
	return s.planRepo.PublishPlan(ctx, plan)
}

// GetProjectPlan 获取指定项目的完整正式计划
func (s *PlanPublishService) GetProjectPlan(ctx context.Context, projectID string) (*domain.FormalPlan, error) {
	return s.planRepo.GetPlan(ctx, projectID)
}

func (s *PlanPublishService) ListTestCases(ctx context.Context, projectID string) ([]domain.TestCase, error) {
	if projectID == "" {
		return nil, fieldError("project id is required")
	}
	plan, err := s.planRepo.GetPlan(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return plan.TestCases, nil
}

func (s *PlanPublishService) ListDevTasks(ctx context.Context, projectID string) ([]domain.DevTask, error) {
	if projectID == "" {
		return nil, fieldError("project id is required")
	}
	plan, err := s.planRepo.GetPlan(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return plan.DevTasks, nil
}

func (s *PlanPublishService) UpdateDevTaskStatus(ctx context.Context, projectID, taskID string, status domain.TaskStatus) error {
	if projectID == "" || taskID == "" {
		return fieldError("project id and task id are required")
	}
	switch status {
	case domain.TaskStatusPendingDev, domain.TaskStatusDeveloping, domain.TaskStatusPendingTest,
		domain.TaskStatusTesting, domain.TaskStatusPendingCheck, domain.TaskStatusAccepted, domain.TaskStatusLaunched:
		return s.planRepo.UpdateDevTaskStatus(ctx, projectID, taskID, status)
	default:
		return fieldError("invalid task status")
	}
}

// generateSimpleID moved to id_generator.go
