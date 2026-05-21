package service

import (
	"context"
	"testing"

	"requirement-splitting/internal/ai"
	"requirement-splitting/internal/domain"
	"requirement-splitting/internal/repository"
)

type capturePlanRepository struct {
	plan         *domain.FormalPlan
	updatedTask  string
	updatedState domain.TaskStatus
}

func (r *capturePlanRepository) PublishPlan(ctx context.Context, plan *domain.FormalPlan) error {
	r.plan = plan
	return nil
}

func (r *capturePlanRepository) GetPlan(ctx context.Context, projectID string) (*domain.FormalPlan, error) {
	if r.plan == nil {
		return &domain.FormalPlan{ProjectID: projectID}, nil
	}
	return r.plan, nil
}

func (r *capturePlanRepository) UpdateDevTaskStatus(ctx context.Context, projectID, taskID string, status domain.TaskStatus) error {
	r.updatedTask = taskID
	r.updatedState = status
	return nil
}

func TestPublishDraftBuildsFormalPlanFromSplitRequirementDraft(t *testing.T) {
	draftRepo := repository.NewMemoryAIDraftRepository()
	aiDraftService := NewAIDraftService(draftRepo, ai.NewStubProvider())
	draft, err := aiDraftService.SplitRequirement(context.Background(), SplitRequirementInput{
		ProjectID: "project-1",
		Content:   "做一个项目需求自动拆分系统",
	})
	if err != nil {
		t.Fatalf("split requirement: %v", err)
	}

	planRepo := &capturePlanRepository{}
	publishService := NewPlanPublishService(draftRepo, planRepo)
	if err := publishService.PublishDraft(context.Background(), "project-1", draft.ID); err != nil {
		t.Fatalf("publish draft: %v", err)
	}

	if planRepo.plan == nil {
		t.Fatal("expected formal plan to be published")
	}
	if len(planRepo.plan.Modules) == 0 {
		t.Fatal("expected modules in formal plan")
	}
	if len(planRepo.plan.FeaturePoints) == 0 {
		t.Fatal("expected feature points in formal plan")
	}
	if len(planRepo.plan.DevTasks) == 0 {
		t.Fatal("expected dev tasks in formal plan")
	}
	if len(planRepo.plan.TestCases) == 0 {
		t.Fatal("expected test cases in formal plan")
	}
	if len(planRepo.plan.AcceptanceItems) == 0 {
		t.Fatal("expected acceptance items in formal plan")
	}
}

func TestUpdateDevTaskStatusValidatesStatus(t *testing.T) {
	planRepo := &capturePlanRepository{}
	publishService := NewPlanPublishService(repository.NewMemoryAIDraftRepository(), planRepo)

	if err := publishService.UpdateDevTaskStatus(context.Background(), "project-1", "task-1", domain.TaskStatus("bad")); err == nil {
		t.Fatal("expected invalid status error")
	}

	if err := publishService.UpdateDevTaskStatus(context.Background(), "project-1", "task-1", domain.TaskStatusDeveloping); err != nil {
		t.Fatalf("update task status: %v", err)
	}
	if planRepo.updatedTask != "task-1" {
		t.Fatalf("expected task-1 to be updated, got %q", planRepo.updatedTask)
	}
	if planRepo.updatedState != domain.TaskStatusDeveloping {
		t.Fatalf("expected developing status, got %s", planRepo.updatedState)
	}
}
