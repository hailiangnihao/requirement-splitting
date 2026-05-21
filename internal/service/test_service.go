package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"requirement-splitting/internal/ai"
	"requirement-splitting/internal/domain"
	"requirement-splitting/internal/repository"
)

type TestService struct {
	// planRepo repository.PlanRepository // 实际项目中需要它来获取 TestCase 详情
	testRepo repository.TestRepository
	provider ai.Provider
}

func NewTestService(testRepo repository.TestRepository, provider ai.Provider) *TestService {
	return &TestService{
		testRepo: testRepo,
		provider: provider,
	}
}

// ConfirmTestCase 人工确认测试用例，将其正式纳入测试计划覆盖率
func (s *TestService) ConfirmTestCase(ctx context.Context, projectID, testCaseID string) error {
	return s.testRepo.ConfirmTestCase(ctx, projectID, testCaseID)
}

func (s *TestService) ListTestRuns(ctx context.Context, projectID string) ([]domain.TestRun, error) {
	if projectID == "" {
		return nil, fieldError("project id is required")
	}
	return s.testRepo.ListTestRuns(ctx, projectID)
}

// RunAITest 触发 AI 执行指定的测试用例
func (s *TestService) RunAITest(ctx context.Context, projectID, testCaseID string) (domain.TestRun, error) {
	// 1. 获取测试用例详情
	// 真实场景下：tc, err := s.planRepo.GetTestCase(ctx, testCaseID)
	// 这里为了演示，我们先伪造一些测试上下文
	tcTitle := "验证用户登录功能"
	tcSteps := "1. 输入正确的用户名和密码 2. 点击登录按钮"
	tcExpected := "跳转到首页，且右上角显示用户头像"

	// 2. 构造 AI 任务输入
	// 注意：ai.TaskType 需要在 internal/ai/types.go 中定义 "execute_ai_test"
	taskInput := ai.TaskInput{
		Type:      ai.TaskType("execute_ai_test"),
		ProjectID: projectID,
		Payload: map[string]any{
			"test_case_id":    testCaseID,
			"title":           tcTitle,
			"steps":           tcSteps,
			"expected_result": tcExpected,
			// 如果是接口测试，这里可能会传入 API URL, Method 等
		},
	}

	// 3. 调用底层的 AI Provider 执行测试
	// 这里的 Provider 可能是真实的 LLM，也可能是当前的 StubProvider
	output, err := s.provider.Run(ctx, taskInput)
	if err != nil {
		return domain.TestRun{}, fmt.Errorf("failed to run AI test: %w", err)
	}

	// 4. 解析 AI 返回的结果
	// 假设 AI 约定好返回格式包含 actual_result, evidence, is_defect_suggested
	actualResult, _ := output.Result["actual_result"].(string)
	isDefectSuggested, _ := output.Result["is_defect_suggested"].(bool)

	evidenceBytes, err := json.Marshal(output.Result["evidence"])
	if err != nil {
		LogWarn(fmt.Sprintf("failed to marshal evidence for test case %s: %v", testCaseID, err))
		evidenceBytes = []byte("{}") // 序列化失败时给个默认空 JSON
	}

	// 5. 组装 TestRun 记录
	testRun := domain.TestRun{
		ID:                generateSimpleID("testrun"),
		ProjectID:         projectID,
		TestCaseID:        testCaseID,
		ExecutedBy:        "AI",
		ExecutionType:     "ai",
		ActualResult:      actualResult,
		Evidence:          evidenceBytes,
		IsDefectSuggested: isDefectSuggested,
		ReviewStatus:      domain.TestRunReviewPending, // 核心机制：强制标记为待人工复核
		CreatedAt:         time.Now(),
	}

	// 6. 保存到数据库
	return s.testRepo.CreateTestRun(ctx, testRun)
}

// ReviewTestRun 人工复核测试执行记录
func (s *TestService) ReviewTestRun(ctx context.Context, projectID, testRunID string, status domain.TestRunReviewStatus) error {
	// 简单的业务校验，确保传入的是合法的终态
	switch status {
	case domain.TestRunReviewPassed, domain.TestRunReviewFailed, domain.TestRunReviewRetest, domain.TestRunReviewIgnored:
		return s.testRepo.ReviewTestRun(ctx, projectID, testRunID, status)
	default:
		return fmt.Errorf("invalid review status: %s", status)
	}
}
