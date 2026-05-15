package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"requirement-splitting/internal/ai"
	"requirement-splitting/internal/domain"
	"requirement-splitting/internal/repository"
)

func TestSplitRequirementRequiresContent(t *testing.T) {
	svc := NewAIDraftService(repository.NewMemoryAIDraftRepository(), ai.NewStubProvider())

	_, err := svc.SplitRequirement(context.Background(), SplitRequirementInput{ProjectID: "project-1"})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestSplitRequirementCreatesValidatedDraft(t *testing.T) {
	svc := NewAIDraftService(repository.NewMemoryAIDraftRepository(), ai.NewStubProvider())

	draft, err := svc.SplitRequirement(context.Background(), SplitRequirementInput{
		ProjectID: "project-1",
		Content:   "做一个项目需求自动拆分系统",
	})
	if err != nil {
		t.Fatalf("split requirement: %v", err)
	}
	if draft.Status != domain.AIDraftStatusValidated {
		t.Fatalf("expected validated draft, got %s", draft.Status)
	}

	var output map[string]any
	if err := json.Unmarshal(draft.OutputJSON, &output); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(output["test_cases"].([]any)) == 0 {
		t.Fatal("expected test cases in AI draft output")
	}
}
