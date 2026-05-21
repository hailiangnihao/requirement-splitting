package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"requirement-splitting/internal/ai"
	"requirement-splitting/internal/domain"
	"requirement-splitting/internal/repository"
)

type ChangeService struct {
	changeRepo repository.ChangeRepository
	planRepo   repository.PlanRepository
	provider   ai.Provider
}

func NewChangeService(changeRepo repository.ChangeRepository, planRepo repository.PlanRepository, provider ai.Provider) *ChangeService {
	return &ChangeService{
		changeRepo: changeRepo,
		planRepo:   planRepo,
		provider:   provider,
	}
}

type SubmitChangeInput struct {
	ProjectID string
	Title     string
	Content   string
	CreatedBy string
}

// SubmitChangeRequest 提交一个新的需求变更申请
func (s *ChangeService) SubmitChangeRequest(ctx context.Context, input SubmitChangeInput) (domain.ChangeRequest, error) {
	if strings.TrimSpace(input.ProjectID) == "" || strings.TrimSpace(input.Content) == "" {
		return domain.ChangeRequest{}, fmt.Errorf("%w: project_id and content are required", ErrValidation)
	}

	// 输入长度验证
	if len(input.Title) > 200 {
		return domain.ChangeRequest{}, fmt.Errorf("%w: title too long (max 200 chars)", ErrValidation)
	}
	if len(input.Content) > 5000 {
		return domain.ChangeRequest{}, fmt.Errorf("%w: content too long (max 5000 chars)", ErrValidation)
	}

	now := time.Now()
	cr := domain.ChangeRequest{
		ID:        generateID(),
		ProjectID: input.ProjectID,
		Title:     input.Title,
		Content:   input.Content,
		Status:    domain.ChangeRequestStatusSubmitted,
		CreatedBy: input.CreatedBy,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return s.changeRepo.CreateChangeRequest(ctx, cr)
}

func (s *ChangeService) ListChangeRequests(ctx context.Context, projectID string) ([]domain.ChangeRequest, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, fmt.Errorf("%w: project_id is required", ErrValidation)
	}
	return s.changeRepo.ListChangeRequests(ctx, projectID)
}

// AnalyzeChangeImpact 触发 AI 分析该变更对当前计划的影响面
func (s *ChangeService) AnalyzeChangeImpact(ctx context.Context, projectID, changeID string) error {
	// 1. 获取变更申请
	cr, err := s.changeRepo.GetChangeRequest(ctx, changeID)
	if err != nil {
		return fmt.Errorf("failed to get change request: %w", err)
	}

	// 2. 获取当前项目的正式计划 (The "Context" for AI)
	currentPlan, err := s.planRepo.GetPlan(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get current plan: %w", err)
	}

	// [轻量级摘要策略]：将庞大的完整计划精简为只有 ID 和标题的树状摘要，防止 Token 超限
	planSummary := buildPlanSummary(currentPlan)

	// 3. 构建 AI 分析任务
	taskInput := ai.TaskInput{
		Type:      ai.TaskAnalyzeChangeImpact,
		ProjectID: projectID,
		Payload: map[string]any{
			"change_title":   cr.Title,
			"change_content": cr.Content,
			"current_plan":   planSummary, // 传入轻量级摘要代替全量数据
		},
	}

	// 4. 调用 AI Provider
	output, err := s.provider.Run(ctx, taskInput)
	if err != nil {
		return fmt.Errorf("ai analysis failed: %w", err)
	}

	// 5. 序列化 AI 报告并更新数据库状态
	analysisBytes, _ := json.Marshal(output.Result)

	return s.changeRepo.UpdateChangeAnalysis(ctx, projectID, changeID, analysisBytes, domain.ChangeRequestStatusAnalyzed)
}

func (s *ChangeService) UpdateChangeStatus(ctx context.Context, projectID, changeID string, status domain.ChangeRequestStatus) error {
	switch status {
	case domain.ChangeRequestStatusSubmitted, domain.ChangeRequestStatusAnalyzed, domain.ChangeRequestStatusAccepted,
		domain.ChangeRequestStatusApplied, domain.ChangeRequestStatusRejected:
		return s.changeRepo.UpdateChangeStatus(ctx, projectID, changeID, status)
	default:
		return fmt.Errorf("%w: invalid change status", ErrValidation)
	}
}

// --- 轻量级摘要结构定义 ---

type nodeSummary struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type fpSummary struct {
	nodeSummary
	Tasks     []nodeSummary `json:"tasks"`
	TestCases []nodeSummary `json:"test_cases"`
}

type moduleSummary struct {
	nodeSummary
	FeaturePoints []fpSummary `json:"feature_points"`
}

// buildPlanSummary 将扁平且庞大的 FormalPlan 组装为轻量级的摘要树
func buildPlanSummary(plan *domain.FormalPlan) []moduleSummary {
	tasksByFP := make(map[string][]nodeSummary)
	for _, task := range plan.DevTasks {
		tasksByFP[task.FeaturePointID] = append(tasksByFP[task.FeaturePointID], nodeSummary{ID: task.ID, Title: task.Name})
	}

	tcByFP := make(map[string][]nodeSummary)
	for _, tc := range plan.TestCases {
		tcByFP[tc.FeaturePointID] = append(tcByFP[tc.FeaturePointID], nodeSummary{ID: tc.ID, Title: tc.Title})
	}

	fpByMod := make(map[string][]fpSummary)
	for _, fp := range plan.FeaturePoints {
		fpByMod[fp.ModuleID] = append(fpByMod[fp.ModuleID], fpSummary{
			nodeSummary: nodeSummary{ID: fp.ID, Title: fp.Name}, // 去掉了冗长的 Description
			Tasks:       tasksByFP[fp.ID],
			TestCases:   tcByFP[fp.ID],
		})
	}

	var modules []moduleSummary
	for _, mod := range plan.Modules {
		fps := fpByMod[mod.ID]
		if fps == nil {
			fps = []fpSummary{}
		}
		modules = append(modules, moduleSummary{
			nodeSummary:   nodeSummary{ID: mod.ID, Title: mod.Name},
			FeaturePoints: fps,
		})
	}

	return modules
}
