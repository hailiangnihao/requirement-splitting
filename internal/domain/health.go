package domain

import "encoding/json"

// HealthMetrics 包含项目健康度的硬核量化指标
type HealthMetrics struct {
	FeaturePointCount    int `json:"feature_point_count"`
	UntestedFeatureCount int `json:"untested_feature_count"` // 没写用例的裸奔功能点
	DevTaskTotal         int `json:"dev_task_total"`
	DevTaskDone          int `json:"dev_task_done"`
	ActiveDefects        int `json:"active_defects"` // 待修复的 Bug 数
	RecentChanges        int `json:"recent_changes"` // 最近 7 天变更数
	BaseScore            int `json:"base_score"`     // 综合基础得分 (0-100)
}

// ProjectHealth 代表带有 AI 洞察的最终健康度报告
type ProjectHealth struct {
	ProjectID string          `json:"project_id"`
	Metrics   HealthMetrics   `json:"metrics"`
	Insight   json.RawMessage `json:"insight"` // AI 生成的管理洞察
}
