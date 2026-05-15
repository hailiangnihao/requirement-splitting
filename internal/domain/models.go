package domain

import (
	"encoding/json"
	"time"
)

type Project struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Objective string        `json:"objective"`
	Scope     string        `json:"scope"`
	Status    ProjectStatus `json:"status"`
	OwnerID   string        `json:"owner_id"`
	Health    HealthLevel   `json:"health"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProjectMember struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Requirement struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"project_id"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	SourceType     string    `json:"source_type"`
	SourceFilename string    `json:"source_filename"`
	CreatedBy      string    `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type AIDraft struct {
	ID               string          `json:"id"`
	ProjectID        string          `json:"project_id"`
	RequirementID    string          `json:"requirement_id"`
	TaskType         string          `json:"task_type"`
	Provider         string          `json:"provider"`
	Model            string          `json:"model"`
	InputJSON        json.RawMessage `json:"input_json"`
	OutputJSON       json.RawMessage `json:"output_json"`
	ValidationErrors json.RawMessage `json:"validation_errors"`
	Status           AIDraftStatus   `json:"status"`
	CreatedBy        string          `json:"created_by"`
	PublishedAt      *time.Time      `json:"published_at"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}
