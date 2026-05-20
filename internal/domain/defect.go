package domain

import (
	"time"
)

type DefectStatus string

const (
	DefectStatusPendingFix    DefectStatus = "pending_fix"    // 待修复
	DefectStatusFixing        DefectStatus = "fixing"         // 修复中
	DefectStatusPendingRetest DefectStatus = "pending_retest" // 待回归
	DefectStatusClosed        DefectStatus = "closed"         // 已关闭
	DefectStatusRejected      DefectStatus = "rejected"       // 已拒绝
)

// Defect 代表项目中的一个缺陷/Bug
type Defect struct {
	ID          string       `json:"id"`
	ProjectID   string       `json:"project_id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Status      DefectStatus `json:"status"`      // "pending_fix", "fixing", "closed" 等，已在 enums.go 中定义
	TestRunID   *string      `json:"test_run_id"` // 核心字段：关联的测试执行记录，实现 AI 证据溯源
	CreatedBy   string       `json:"created_by"`  // 提单人 UUID
	AssignedTo  string       `json:"assigned_to"` // 修复人 UUID
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}
