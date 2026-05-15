package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"requirement-splitting/internal/domain"
)

type PGProjectRepository struct {
	pool *pgxpool.Pool
}

func NewPGProjectRepository(pool *pgxpool.Pool) *PGProjectRepository {
	return &PGProjectRepository{pool: pool}
}

func (r *PGProjectRepository) CreateProject(ctx context.Context, project domain.Project) (domain.Project, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO projects (name, objective, scope, status, owner_id, health_level)
		VALUES ($1, $2, $3, $4, NULLIF($5, '')::uuid, $6)
		RETURNING id::text, name, objective, scope, status, COALESCE(owner_id::text, ''), health_level, created_at, updated_at
	`, project.Name, project.Objective, project.Scope, project.Status, project.OwnerID, project.Health)

	var created domain.Project
	err := row.Scan(&created.ID, &created.Name, &created.Objective, &created.Scope, &created.Status, &created.OwnerID, &created.Health, &created.CreatedAt, &created.UpdatedAt)
	return created, err
}

func (r *PGProjectRepository) ListProjects(ctx context.Context) ([]domain.Project, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, name, objective, scope, status, COALESCE(owner_id::text, ''), health_level, created_at, updated_at
		FROM projects
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []domain.Project
	for rows.Next() {
		var project domain.Project
		if err := rows.Scan(&project.ID, &project.Name, &project.Objective, &project.Scope, &project.Status, &project.OwnerID, &project.Health, &project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, rows.Err()
}

func (r *PGProjectRepository) GetProject(ctx context.Context, id string) (domain.Project, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id::text, name, objective, scope, status, COALESCE(owner_id::text, ''), health_level, created_at, updated_at
		FROM projects
		WHERE id = $1
	`, id)

	var project domain.Project
	if err := row.Scan(&project.ID, &project.Name, &project.Objective, &project.Scope, &project.Status, &project.OwnerID, &project.Health, &project.CreatedAt, &project.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Project{}, ErrNotFound
		}
		return domain.Project{}, err
	}
	return project, nil
}

func (r *PGProjectRepository) CreateProjectMember(ctx context.Context, member domain.ProjectMember) (domain.ProjectMember, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, $3)
		RETURNING id::text, project_id::text, user_id::text, role, created_at, updated_at
	`, member.ProjectID, member.UserID, member.Role)

	var created domain.ProjectMember
	err := row.Scan(&created.ID, &created.ProjectID, &created.UserID, &created.Role, &created.CreatedAt, &created.UpdatedAt)
	return created, err
}

func (r *PGProjectRepository) CreateRequirement(ctx context.Context, requirement domain.Requirement) (domain.Requirement, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO requirements (project_id, title, content, source_type, source_filename, created_by)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, '')::uuid)
		RETURNING id::text, project_id::text, title, content, source_type, source_filename, COALESCE(created_by::text, ''), created_at, updated_at
	`, requirement.ProjectID, requirement.Title, requirement.Content, requirement.SourceType, requirement.SourceFilename, requirement.CreatedBy)

	var created domain.Requirement
	if err := row.Scan(&created.ID, &created.ProjectID, &created.Title, &created.Content, &created.SourceType, &created.SourceFilename, &created.CreatedBy, &created.CreatedAt, &created.UpdatedAt); err != nil {
		if isForeignKeyViolation(err) {
			return domain.Requirement{}, ErrNotFound
		}
		return domain.Requirement{}, err
	}
	return created, nil
}

func (r *PGProjectRepository) ListRequirements(ctx context.Context, projectID string) ([]domain.Requirement, error) {
	if _, err := r.GetProject(ctx, projectID); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id::text, project_id::text, title, content, source_type, source_filename, COALESCE(created_by::text, ''), created_at, updated_at
		FROM requirements
		WHERE project_id = $1
		ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requirements []domain.Requirement
	for rows.Next() {
		var requirement domain.Requirement
		if err := rows.Scan(&requirement.ID, &requirement.ProjectID, &requirement.Title, &requirement.Content, &requirement.SourceType, &requirement.SourceFilename, &requirement.CreatedBy, &requirement.CreatedAt, &requirement.UpdatedAt); err != nil {
			return nil, err
		}
		requirements = append(requirements, requirement)
	}
	return requirements, rows.Err()
}

func isForeignKeyViolation(err error) bool {
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23503"
	}
	return false
}
