package domain

import (
	"encoding/json"
	"time"
)

// TestRun 代表一次测试执行记录（可能是人工执行，也可能是 AI 执行）
type TestRun struct {
	ID                string              `json:"id"`
	ProjectID         string              `json:"project_id"`
	TestCaseID        string              `json:"test_case_id"`
	ExecutedBy        string              `json:"executed_by"`         // 执行人 ID 或 "AI"
	ExecutionType     string              `json:"execution_type"`      // "manual" 或 "ai"
	ActualResult      string              `json:"actual_result"`       // 实际结果描述
	Evidence          json.RawMessage     `json:"evidence"`            // 证据：截图、日志、接口响应等 (JSON)
	IsDefectSuggested bool                `json:"is_defect_suggested"` // AI 是否建议针对此结果提 Bug
	ReviewStatus      TestRunReviewStatus `json:"review_status"`       // 核心字段：人工复核状态
	CreatedAt         time.Time           `json:"created_at"`
}
