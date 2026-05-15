package service

import (
	"context"
	"errors"
	"strings"

	"requirement-splitting/internal/domain"
	"requirement-splitting/internal/repository"
)

var ErrValidation = errors.New("validation failed")

type ProjectService struct {
	repo repository.ProjectRepository
}

func NewProjectService(repo repository.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

type CreateProjectInput struct {
	Name      string
	Objective string
	Scope     string
	OwnerID   string
	OwnerRole string
}

func (s *ProjectService) CreateProject(ctx context.Context, input CreateProjectInput) (domain.Project, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return domain.Project{}, fieldError("name is required")
	}

	project, err := s.repo.CreateProject(ctx, domain.Project{
		Name:      name,
		Objective: strings.TrimSpace(input.Objective),
		Scope:     strings.TrimSpace(input.Scope),
		Status:    domain.ProjectStatusPlanning,
		OwnerID:   input.OwnerID,
		Health:    domain.HealthLevelAttention,
	})
	if err != nil {
		return domain.Project{}, err
	}

	if strings.TrimSpace(input.OwnerID) != "" {
		role := strings.TrimSpace(input.OwnerRole)
		if role == "" {
			role = "owner"
		}
		if _, err := s.repo.CreateProjectMember(ctx, domain.ProjectMember{
			ProjectID: project.ID,
			UserID:    input.OwnerID,
			Role:      role,
		}); err != nil {
			return domain.Project{}, err
		}
	}

	return project, nil
}

func (s *ProjectService) ListProjects(ctx context.Context) ([]domain.Project, error) {
	return s.repo.ListProjects(ctx)
}

func (s *ProjectService) GetProject(ctx context.Context, id string) (domain.Project, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Project{}, fieldError("project id is required")
	}
	return s.repo.GetProject(ctx, id)
}

type AddRequirementInput struct {
	ProjectID      string
	Title          string
	Content        string
	SourceType     string
	SourceFilename string
	CreatedBy      string
}

func (s *ProjectService) AddRequirement(ctx context.Context, input AddRequirementInput) (domain.Requirement, error) {
	projectID := strings.TrimSpace(input.ProjectID)
	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	if projectID == "" {
		return domain.Requirement{}, fieldError("project id is required")
	}
	if content == "" {
		return domain.Requirement{}, fieldError("content is required")
	}
	if title == "" {
		title = "原始需求"
	}
	sourceType := strings.TrimSpace(input.SourceType)
	if sourceType == "" {
		sourceType = "manual"
	}

	return s.repo.CreateRequirement(ctx, domain.Requirement{
		ProjectID:      projectID,
		Title:          title,
		Content:        content,
		SourceType:     sourceType,
		SourceFilename: strings.TrimSpace(input.SourceFilename),
		CreatedBy:      input.CreatedBy,
	})
}

func (s *ProjectService) ListRequirements(ctx context.Context, projectID string) ([]domain.Requirement, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, fieldError("project id is required")
	}
	return s.repo.ListRequirements(ctx, projectID)
}

func fieldError(message string) error {
	return errors.Join(ErrValidation, errors.New(message))
}
