package repository

import (
	"context"
	"errors"
	"time"

	"requirement-splitting/internal/domain"
)

var ErrNotFound = errors.New("not found")

type ProjectRepository interface {
	CreateProject(ctx context.Context, project domain.Project) (domain.Project, error)
	ListProjects(ctx context.Context) ([]domain.Project, error)
	GetProject(ctx context.Context, id string) (domain.Project, error)
	CreateProjectMember(ctx context.Context, member domain.ProjectMember) (domain.ProjectMember, error)
	CreateRequirement(ctx context.Context, requirement domain.Requirement) (domain.Requirement, error)
	ListRequirements(ctx context.Context, projectID string) ([]domain.Requirement, error)
}

type MemoryProjectRepository struct {
	projects     map[string]domain.Project
	members      map[string]domain.ProjectMember
	requirements map[string]domain.Requirement
	now          func() time.Time
	newID        func() string
}

func NewMemoryProjectRepository() *MemoryProjectRepository {
	return &MemoryProjectRepository{
		projects:     map[string]domain.Project{},
		members:      map[string]domain.ProjectMember{},
		requirements: map[string]domain.Requirement{},
		now:          time.Now,
		newID:        newMemoryID,
	}
}

func (r *MemoryProjectRepository) CreateProject(ctx context.Context, project domain.Project) (domain.Project, error) {
	if project.ID == "" {
		project.ID = r.newID()
	}
	now := r.now()
	project.CreatedAt = now
	project.UpdatedAt = now
	r.projects[project.ID] = project
	return project, nil
}

func (r *MemoryProjectRepository) ListProjects(ctx context.Context) ([]domain.Project, error) {
	projects := make([]domain.Project, 0, len(r.projects))
	for _, project := range r.projects {
		projects = append(projects, project)
	}
	return projects, nil
}

func (r *MemoryProjectRepository) GetProject(ctx context.Context, id string) (domain.Project, error) {
	project, ok := r.projects[id]
	if !ok {
		return domain.Project{}, ErrNotFound
	}
	return project, nil
}

func (r *MemoryProjectRepository) CreateProjectMember(ctx context.Context, member domain.ProjectMember) (domain.ProjectMember, error) {
	if member.ID == "" {
		member.ID = r.newID()
	}
	now := r.now()
	member.CreatedAt = now
	member.UpdatedAt = now
	r.members[member.ID] = member
	return member, nil
}

func (r *MemoryProjectRepository) CreateRequirement(ctx context.Context, requirement domain.Requirement) (domain.Requirement, error) {
	if _, ok := r.projects[requirement.ProjectID]; !ok {
		return domain.Requirement{}, ErrNotFound
	}
	if requirement.ID == "" {
		requirement.ID = r.newID()
	}
	now := r.now()
	requirement.CreatedAt = now
	requirement.UpdatedAt = now
	r.requirements[requirement.ID] = requirement
	return requirement, nil
}

func (r *MemoryProjectRepository) ListRequirements(ctx context.Context, projectID string) ([]domain.Requirement, error) {
	if _, ok := r.projects[projectID]; !ok {
		return nil, ErrNotFound
	}
	requirements := make([]domain.Requirement, 0)
	for _, requirement := range r.requirements {
		if requirement.ProjectID == projectID {
			requirements = append(requirements, requirement)
		}
	}
	return requirements, nil
}

func newMemoryID() string {
	return time.Now().UTC().Format("20060102150405.000000000")
}
