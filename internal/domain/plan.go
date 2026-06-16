package domain

import "time"

// FormalPlan 代表从 AI 草稿解析并确认后的正式发布计划
type FormalPlan struct {
	ProjectID       string
	DraftID         string
	Modules         []Module
	Milestones      []Milestone
	FeaturePoints   []FeaturePoint
	DevTasks        []DevTask
	TestCases       []TestCase
	AcceptanceItems []AcceptanceItem
}

type Module struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Milestone struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type FeaturePoint struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	ModuleID    string    `json:"module_id"` // 关联到 Module
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type DevTask struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"project_id"`
	FeaturePointID string    `json:"feature_point_id"` // 关联到 FeaturePoint
	Name           string    `json:"name"`
	Status         string    `json:"status"` // e.g., "pending", "in_progress", "done"
	CreatedAt      time.Time `json:"created_at"`
}

type TestCase struct {
	ID                 string    `json:"id"`
	ProjectID          string    `json:"project_id"`
	FeaturePointID     string    `json:"feature_point_id"`
	Title              string    `json:"title"`
	ConfirmationStatus string    `json:"confirmation_status"` // 核心设计："pending_human_confirmation", "confirmed"
	CreatedAt          time.Time `json:"created_at"`
}

type AcceptanceItem struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
