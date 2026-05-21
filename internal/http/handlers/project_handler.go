package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"requirement-splitting/internal/service"
)

type ProjectHandler struct {
	service *service.ProjectService
}

func NewProjectHandler(service *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

func (h *ProjectHandler) RegisterRoutes(router chi.Router) {
	router.Post("/api/projects", h.createProject)
	router.Get("/api/projects", h.listProjects)
	router.Get("/api/projects/{id}", h.getProject)
	router.Post("/api/projects/{id}/requirements", h.addRequirement)
	router.Get("/api/projects/{id}/requirements", h.listRequirements)
}

type createProjectRequest struct {
	Name      string `json:"name"`
	Objective string `json:"objective"`
	Scope     string `json:"scope"`
	OwnerID   string `json:"owner_id"`
	OwnerRole string `json:"owner_role"`
}

func (h *ProjectHandler) createProject(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	project, err := h.service.CreateProject(r.Context(), service.CreateProjectInput(req))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, project)
}

func (h *ProjectHandler) listProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.ListProjects(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func (h *ProjectHandler) getProject(w http.ResponseWriter, r *http.Request) {
	project, err := h.service.GetProject(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

type addRequirementRequest struct {
	Title          string `json:"title"`
	Content        string `json:"content"`
	SourceType     string `json:"source_type"`
	SourceFilename string `json:"source_filename"`
	CreatedBy      string `json:"created_by"`
}

func (h *ProjectHandler) addRequirement(w http.ResponseWriter, r *http.Request) {
	var req addRequirementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	requirement, err := h.service.AddRequirement(r.Context(), service.AddRequirementInput{
		ProjectID:      chi.URLParam(r, "id"),
		Title:          req.Title,
		Content:        req.Content,
		SourceType:     req.SourceType,
		SourceFilename: req.SourceFilename,
		CreatedBy:      req.CreatedBy,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, requirement)
}

func (h *ProjectHandler) listRequirements(w http.ResponseWriter, r *http.Request) {
	requirements, err := h.service.ListRequirements(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, requirements)
}

// Helper functions moved to helpers.go
