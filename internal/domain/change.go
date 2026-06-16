package domain

import (
	"encoding/json"
	"time"
)

// ChangeRequest 代表一次需求变更申请
type ChangeRequest struct {
	ID             string              `json:"id"`
	ProjectID      string              `json:"project_id"`
	Title          string              `json:"title"`
	Content        string              `json:"content"`         // 变更的具体诉求
	Status         ChangeRequestStatus `json:"status"`          // submitted, analyzed, accepted, applied, rejected
	ImpactAnalysis json.RawMessage     `json:"impact_analysis"` // AI 生成的影响面分析报告 (JSON)
	CreatedBy      string              `json:"created_by"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}
