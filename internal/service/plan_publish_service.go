package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"requirement-splitting/internal/domain"
	"requirement-splitting/internal/repository"
)

// DraftOutputJSON 定义了我们期望从 AI 获取的结构化 JSON 格式
// 这里假设了 AI 返回的大致层级结构，你需要确保你的 AI Stub Provider 也是按这个结构返回的
type DraftOutputJSON struct {
	Modules []struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		FeaturePoints []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Tasks       []struct {
				Name string `json:"name"`
			} `json:"tasks"`
			TestCases []struct {
				Title string `json:"title"`
			} `json:"test_cases"`
		} `json:"feature_points"`
	} `json:"modules"`
	Milestones []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"milestones"`
	AcceptanceItems []struct {
		Description string `json:"description"`
	} `json:"acceptance_items"`
}

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
	var outData DraftOutputJSON
	if err := json.Unmarshal(targetDraft.OutputJSON, &outData); err != nil {
		return fmt.Errorf("failed to parse draft output JSON: %w", err)
	}

	// 3. 组装 FormalPlan 领域模型
	now := time.Now()
	plan := &domain.FormalPlan{
		ProjectID: projectID,
		DraftID:   draftID,
	}

	for i, m := range outData.Modules {
		modID := generateSimpleID(fmt.Sprintf("mod-%d", i))
		plan.Modules = append(plan.Modules, domain.Module{
			ID: modID, ProjectID: projectID, Name: m.Name, Description: m.Description, CreatedAt: now,
		})

		for j, fp := range m.FeaturePoints {
			fpID := generateSimpleID(fmt.Sprintf("fp-%d-%d", i, j))
			plan.FeaturePoints = append(plan.FeaturePoints, domain.FeaturePoint{
				ID: fpID, ProjectID: projectID, ModuleID: modID, Name: fp.Name, Description: fp.Description, CreatedAt: now,
			})

			for k, t := range fp.Tasks {
				plan.DevTasks = append(plan.DevTasks, domain.DevTask{
					ID:        generateSimpleID(fmt.Sprintf("task-%d-%d-%d", i, j, k)),
					ProjectID: projectID, FeaturePointID: fpID, Name: t.Name, Status: string(domain.TaskStatusPendingDev), CreatedAt: now,
				})
			}

			for k, tc := range fp.TestCases {
				plan.TestCases = append(plan.TestCases, domain.TestCase{
					ID:        generateSimpleID(fmt.Sprintf("tc-%d-%d-%d", i, j, k)),
					ProjectID: projectID, FeaturePointID: fpID, Title: tc.Title,
					ConfirmationStatus: string(domain.TestCaseConfirmationPending), // 核心设定：待人工确认
					CreatedAt:          now,
				})
			}
		}
	}

	for i, ms := range outData.Milestones {
		plan.Milestones = append(plan.Milestones, domain.Milestone{
			ID: generateSimpleID(fmt.Sprintf("ms-%d", i)), ProjectID: projectID, Name: ms.Name, Description: ms.Description, CreatedAt: now,
		})
	}

	for i, acc := range outData.AcceptanceItems {
		plan.AcceptanceItems = append(plan.AcceptanceItems, domain.AcceptanceItem{
			ID: generateSimpleID(fmt.Sprintf("acc-%d", i)), ProjectID: projectID, Description: acc.Description, CreatedAt: now,
		})
	}

	// 4. 调用仓储层事务，一并落库
	return s.planRepo.PublishPlan(ctx, plan)
}

// GetProjectPlan 获取指定项目的完整正式计划
func (s *PlanPublishService) GetProjectPlan(ctx context.Context, projectID string) (*domain.FormalPlan, error) {
	return s.planRepo.GetPlan(ctx, projectID)
}

// generateSimpleID 生成一个简易的 ID，实际项目中建议使用 google/uuid 库生成标准 UUID
func generateSimpleID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
