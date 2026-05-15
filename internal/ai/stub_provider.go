package ai

import (
	"context"
	"fmt"
)

type StubProvider struct{}

func NewStubProvider() *StubProvider {
	return &StubProvider{}
}

func (p *StubProvider) Run(ctx context.Context, input TaskInput) (TaskOutput, error) {
	switch input.Type {
	case TaskSplitRequirement:
		return TaskOutput{Type: input.Type, Result: splitRequirementStub()}, nil
	default:
		return TaskOutput{}, fmt.Errorf("stub provider does not support task %s", input.Type)
	}
}

func splitRequirementStub() map[string]any {
	return map[string]any{
		"modules": []map[string]any{
			{"key": "module_requirement", "name": "需求拆分", "description": "录入原始需求并生成结构化项目计划"},
			{"key": "module_testing", "name": "测试验收", "description": "生成测试用例、AI 辅助测试并进入人工复核"},
		},
		"milestones": []map[string]any{
			{"key": "milestone_mvp", "name": "第一版闭环", "description": "打通需求拆分到测试验收的最小闭环"},
		},
		"feature_points": []map[string]any{
			{
				"key":           "feature_split",
				"module_key":    "module_requirement",
				"milestone_key": "milestone_mvp",
				"title":         "AI 分阶段拆分需求",
				"description":   "根据原始需求生成模块、功能点、任务、测试用例和验收项草稿",
				"priority":      "high",
			},
			{
				"key":           "feature_ai_test",
				"module_key":    "module_testing",
				"milestone_key": "milestone_mvp",
				"title":         "AI 辅助执行测试",
				"description":   "AI 执行或辅助执行测试用例，生成证据和缺陷草稿，人工复核后生效",
				"priority":      "high",
			},
		},
		"dev_tasks": []map[string]any{
			{"key": "task_split_api", "feature_point_key": "feature_split", "title": "实现 AI 拆分草稿接口", "description": "保存 AI 输出和校验结果", "priority": "high"},
			{"key": "task_ai_test_api", "feature_point_key": "feature_ai_test", "title": "实现 AI 测试执行接口", "description": "记录 AI 测试结果、证据和人工复核状态", "priority": "high"},
		},
		"test_cases": []map[string]any{
			{
				"key":               "case_split_normal",
				"feature_point_key": "feature_split",
				"dev_task_key":      "task_split_api",
				"title":             "原始需求可生成结构化草稿",
				"case_type":         "normal",
				"priority":          "high",
				"preconditions":     "项目已创建且存在原始需求",
				"steps":             []string{"进入需求拆分页", "触发 AI 拆分", "查看生成草稿"},
				"test_data":         "一段项目需求文本",
				"expected_result":   "系统生成模块、功能点、任务、测试用例和验收项草稿",
			},
			{
				"key":               "case_ai_test_review",
				"feature_point_key": "feature_ai_test",
				"dev_task_key":      "task_ai_test_api",
				"title":             "AI 测试结果必须人工复核后生效",
				"case_type":         "risk",
				"priority":          "high",
				"preconditions":     "存在已确认测试用例",
				"steps":             []string{"触发 AI 测试", "查看 AI 证据", "人工确认测试结论"},
				"test_data":         "已确认的测试用例",
				"expected_result":   "未复核的 AI 测试结果不计入正式测试结论",
			},
		},
		"acceptance_items": []map[string]any{
			{"key": "acc_split", "feature_point_key": "feature_split", "milestone_key": "milestone_mvp", "title": "需求拆分草稿可人工发布", "pass_criteria": "草稿通过校验后才能发布为正式计划"},
			{"key": "acc_ai_test", "feature_point_key": "feature_ai_test", "milestone_key": "milestone_mvp", "title": "AI 测试需要人工确认", "pass_criteria": "AI 测试结果未经人工复核不能作为最终验收依据"},
		},
		"risks": []map[string]any{
			{"key": "risk_ai_overwrite", "category": "ai_control", "level": "high", "title": "AI 结果直接污染正式计划", "description": "必须通过草稿区和人工发布机制控制 AI 输出"},
		},
	}
}
