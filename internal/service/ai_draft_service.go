package service

import (
	"context"
	"encoding/json"
	"strings"

	"requirement-splitting/internal/ai"
	"requirement-splitting/internal/domain"
	"requirement-splitting/internal/repository"
)

type AIDraftService struct {
	repo     repository.AIDraftRepository
	provider ai.Provider
}

func NewAIDraftService(repo repository.AIDraftRepository, provider ai.Provider) *AIDraftService {
	return &AIDraftService{repo: repo, provider: provider}
}

type SplitRequirementInput struct {
	ProjectID     string
	RequirementID string
	Content       string
	CreatedBy     string
}

func (s *AIDraftService) SplitRequirement(ctx context.Context, input SplitRequirementInput) (domain.AIDraft, error) {
	if strings.TrimSpace(input.ProjectID) == "" {
		return domain.AIDraft{}, fieldError("project id is required")
	}
	if strings.TrimSpace(input.Content) == "" {
		return domain.AIDraft{}, fieldError("content is required")
	}

	taskInput := ai.TaskInput{
		Type:      ai.TaskSplitRequirement,
		ProjectID: input.ProjectID,
		Payload: map[string]any{
			"requirement_id": input.RequirementID,
			"content":        input.Content,
		},
	}
	output, err := s.provider.Run(ctx, taskInput)
	if err != nil {
		return domain.AIDraft{}, err
	}

	inputJSON, err := json.Marshal(taskInput)
	if err != nil {
		return domain.AIDraft{}, err
	}
	outputJSON, err := json.Marshal(output.Result)
	if err != nil {
		return domain.AIDraft{}, err
	}
	validationErrors := validateSplitRequirementOutput(output.Result)
	validationJSON, err := json.Marshal(validationErrors)
	if err != nil {
		return domain.AIDraft{}, err
	}

	status := domain.AIDraftStatusValidated
	if len(validationErrors) > 0 {
		status = domain.AIDraftStatusInvalid
	}

	return s.repo.CreateAIDraft(ctx, domain.AIDraft{
		ProjectID:        input.ProjectID,
		RequirementID:    input.RequirementID,
		TaskType:         string(ai.TaskSplitRequirement),
		Provider:         "stub",
		Model:            "stub-v1",
		InputJSON:        inputJSON,
		OutputJSON:       outputJSON,
		ValidationErrors: validationJSON,
		Status:           status,
		CreatedBy:        input.CreatedBy,
	})
}

func (s *AIDraftService) ListDrafts(ctx context.Context, projectID string) ([]domain.AIDraft, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, fieldError("project id is required")
	}
	return s.repo.ListAIDrafts(ctx, projectID)
}

func validateSplitRequirementOutput(result map[string]any) []string {
	errs := []string{}
	requiredArrays := []string{"modules", "milestones", "feature_points", "dev_tasks", "test_cases", "acceptance_items"}
	for _, key := range requiredArrays {
		value, ok := result[key]
		if !ok {
			errs = append(errs, key+" is required")
			continue
		}
		items, ok := value.([]map[string]any)
		if ok && len(items) > 0 {
			continue
		}
		if rawItems, ok := value.([]any); ok && len(rawItems) > 0 {
			continue
		}
		errs = append(errs, key+" must not be empty")
	}
	return errs
}
