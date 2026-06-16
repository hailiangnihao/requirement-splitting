package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"requirement-splitting/internal/domain"
	"requirement-splitting/internal/service"
)

type DefectHandler struct {
	service *service.DefectService
}

func NewDefectHandler(service *service.DefectService) *DefectHandler {
	return &DefectHandler{service: service}
}

type createDefectRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	TestRunID   *string `json:"test_run_id"`
	CreatedBy   string  `json:"created_by"`
}

func (h *DefectHandler) CreateDefect(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")

	var req createDefectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	defect, err := h.service.CreateDefect(r.Context(), service.CreateDefectInput{
		ProjectID:   projectID,
		Title:       req.Title,
		Description: req.Description,
		TestRunID:   req.TestRunID,
		CreatedBy:   req.CreatedBy,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, defect)
}

func (h *DefectHandler) ListDefects(w http.ResponseWriter, r *http.Request) {
	defects, err := h.service.ListDefects(r.Context(), chi.URLParam(r, "project_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, defects)
}

type updateDefectStatusRequest struct {
	Status string `json:"status"`
}

func (h *DefectHandler) UpdateDefectStatus(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	defectID := chi.URLParam(r, "id")

	var req updateDefectStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := h.service.UpdateDefectStatus(r.Context(), projectID, defectID, domain.DefectStatus(req.Status)); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "defect status updated"})
}
