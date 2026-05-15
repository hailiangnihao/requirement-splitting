package service

import (
	"context"
	"errors"
	"testing"

	"requirement-splitting/internal/repository"
)

func TestCreateProjectValidatesName(t *testing.T) {
	svc := NewProjectService(repository.NewMemoryProjectRepository())

	_, err := svc.CreateProject(context.Background(), CreateProjectInput{Name: " "})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestCreateProjectCreatesOwnerMembership(t *testing.T) {
	repo := repository.NewMemoryProjectRepository()
	svc := NewProjectService(repo)

	project, err := svc.CreateProject(context.Background(), CreateProjectInput{
		Name:      "AI PM",
		OwnerID:   "user-1",
		OwnerRole: "owner",
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if project.ID == "" {
		t.Fatal("expected project id")
	}
}

func TestAddRequirementRequiresContent(t *testing.T) {
	svc := NewProjectService(repository.NewMemoryProjectRepository())

	_, err := svc.AddRequirement(context.Background(), AddRequirementInput{
		ProjectID: "project-1",
		Content:   " ",
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestAddRequirementStoresRequirement(t *testing.T) {
	repo := repository.NewMemoryProjectRepository()
	svc := NewProjectService(repo)

	project, err := svc.CreateProject(context.Background(), CreateProjectInput{Name: "AI PM"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	requirement, err := svc.AddRequirement(context.Background(), AddRequirementInput{
		ProjectID: project.ID,
		Content:   "做一个项目需求自动拆分系统",
	})
	if err != nil {
		t.Fatalf("add requirement: %v", err)
	}
	if requirement.Title != "原始需求" {
		t.Fatalf("unexpected title %q", requirement.Title)
	}
}
