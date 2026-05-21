package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"requirement-splitting/internal/domain"
	"requirement-splitting/internal/repository"
)

// AITestRunner 定义了触发自动化测试的接口，用于解耦 Service 间的强依赖
type AITestRunner interface {
	RunAITest(ctx context.Context, projectID, testCaseID string) (domain.TestRun, error)
}

type DefectService struct {
	defectRepo repository.DefectRepository
	testRepo   repository.TestRepository // 注入测试仓储，用于追溯证据
	testRunner AITestRunner              // 注入测试执行器，用于自动回归
}

func NewDefectService(defectRepo repository.DefectRepository, testRepo repository.TestRepository, testRunner AITestRunner) *DefectService {
	return &DefectService{
		defectRepo: defectRepo,
		testRepo:   testRepo,
		testRunner: testRunner,
	}
}

type CreateDefectInput struct {
	ProjectID   string
	Title       string
	Description string
	TestRunID   *string
	CreatedBy   string
}

func (s *DefectService) CreateDefect(ctx context.Context, input CreateDefectInput) (domain.Defect, error) {
	if strings.TrimSpace(input.ProjectID) == "" || strings.TrimSpace(input.Title) == "" {
		return domain.Defect{}, fmt.Errorf("%w: project_id and title are required", ErrValidation)
	}

	// 输入长度验证
	if len(input.Title) > 200 {
		return domain.Defect{}, fmt.Errorf("%w: title too long (max 200 chars)", ErrValidation)
	}
	if len(input.Description) > 5000 {
		return domain.Defect{}, fmt.Errorf("%w: description too long (max 5000 chars)", ErrValidation)
	}

	finalDescription := input.Description

	// 【核心体验增强】如果关联了 TestRun，自动追加 AI 抓取的现场证据
	if input.TestRunID != nil && *input.TestRunID != "" {
		testRun, err := s.testRepo.GetTestRun(ctx, *input.TestRunID)
		if err == nil { // 即便没查到也别阻断流程，尽量容错
			appendStr := fmt.Sprintf("\n\n--- 🤖 AI 测试执行证据 ---\n[测试执行ID]: %s\n[实际结果]: %s\n[现场数据]: %s",
				testRun.ID, testRun.ActualResult, string(testRun.Evidence))
			finalDescription = strings.TrimSpace(finalDescription) + appendStr

			// 自动将原 TestRun 标记为 failed，移出待复核队列
			_ = s.testRepo.ReviewTestRun(ctx, input.ProjectID, *input.TestRunID, domain.TestRunReviewFailed)
		}
	}

	now := time.Now()
	defect := domain.Defect{
		ID:          generateID(),
		ProjectID:   input.ProjectID,
		Title:       input.Title,
		Description: finalDescription,
		Status:      domain.DefectStatusPendingFix, // 初始状态为待修复
		TestRunID:   input.TestRunID,
		CreatedBy:   "", // 暂时留空，后续可以从认证上下文获取
		AssignedTo:  "", // 暂时留空
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return s.defectRepo.CreateDefect(ctx, defect)
}

func (s *DefectService) ListDefects(ctx context.Context, projectID string) ([]domain.Defect, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, fmt.Errorf("%w: project_id is required", ErrValidation)
	}
	return s.defectRepo.ListDefects(ctx, projectID)
}

// UpdateDefectStatus 变更缺陷状态
func (s *DefectService) UpdateDefectStatus(ctx context.Context, projectID, defectID string, status domain.DefectStatus) error {
	switch status {
	case domain.DefectStatusPendingConfirm, domain.DefectStatusPendingFix, domain.DefectStatusFixing,
		domain.DefectStatusPendingRegression, domain.DefectStatusRegressionPassed, domain.DefectStatusClosed, domain.DefectStatusRejected:
	default:
		return fmt.Errorf("%w: invalid defect status", ErrValidation)
	}

	err := s.defectRepo.UpdateDefectStatus(ctx, projectID, defectID, status)
	if err != nil {
		return err
	}

	// 【自动化回归魔法 ✨】
	// 当开发将 Bug 状态变更为 "待回归" (pending_regression) 时，自动触发 AI 重新跑一遍原用例
	if status == domain.DefectStatusPendingRegression {
		defect, err := s.defectRepo.GetDefect(ctx, defectID)
		if err == nil && defect.TestRunID != nil && *defect.TestRunID != "" {
			testRun, err := s.testRepo.GetTestRun(ctx, *defect.TestRunID)
			if err == nil && testRun.TestCaseID != "" {
				// 使用 go func 异步触发，不阻塞当前的 HTTP 响应
				go func() {
					// 添加超时控制，防止 goroutine 泄漏
					bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
					defer cancel()
					_, _ = s.testRunner.RunAITest(bgCtx, projectID, testRun.TestCaseID)
				}()
			}
		}
	}

	return nil
}
