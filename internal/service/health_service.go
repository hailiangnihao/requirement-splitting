package service

import (
	"context"
	"encoding/json"
	"fmt"

	"requirement-splitting/internal/ai"
	"requirement-splitting/internal/domain"
	"requirement-splitting/internal/repository"
)

type HealthService struct {
	healthRepo repository.HealthRepository
	provider   ai.Provider
}

func NewHealthService(healthRepo repository.HealthRepository, provider ai.Provider) *HealthService {
	return &HealthService{
		healthRepo: healthRepo,
		provider:   provider,
	}
}

func (s *HealthService) GetProjectHealth(ctx context.Context, projectID string) (domain.ProjectHealth, error) {
	metrics, err := s.healthRepo.GetMetrics(ctx, projectID)
	if err != nil {
		return domain.ProjectHealth{}, fmt.Errorf("failed to get metrics: %w", err)
	}

	// 简单的规则引擎：计算基准健康分
	score := 100
	// 扣分项 1: 每个活跃缺陷扣 5 分
	score -= metrics.ActiveDefects * 5
	// 扣分项 2: 最近有变更扣 2 分 (代表波动性)
	score -= metrics.RecentChanges * 2
	// 扣分项 3: 覆盖率不足惩罚
	if metrics.FeaturePointCount > 0 {
		untestedRatio := float64(metrics.UntestedFeatureCount) / float64(metrics.FeaturePointCount)
		score -= int(untestedRatio * 30) // 如果全是裸奔的，最高扣 30 分
	}

	if score < 0 {
		score = 0
	}
	metrics.BaseScore = score

	// 调用 AI 获取洞察
	taskInput := ai.TaskInput{
		Type:      ai.TaskType("generate_health_insight"),
		ProjectID: projectID,
		Payload: map[string]any{
			"feature_point_count":    metrics.FeaturePointCount,
			"untested_feature_count": metrics.UntestedFeatureCount,
			"dev_task_total":         metrics.DevTaskTotal,
			"dev_task_done":          metrics.DevTaskDone,
			"active_defects":         metrics.ActiveDefects,
			"recent_changes":         metrics.RecentChanges,
			"base_score":             metrics.BaseScore,
		},
	}

	output, err := s.provider.Run(ctx, taskInput)
	insightBytes, _ := json.Marshal(output.Result)
	if err != nil {
		insightBytes = []byte(`{"error": "AI insight generation failed"}`)
	}

	return domain.ProjectHealth{
		ProjectID: projectID,
		Metrics:   metrics,
		Insight:   insightBytes,
	}, nil
}
