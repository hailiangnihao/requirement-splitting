package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"requirement-splitting/internal/domain"
	"requirement-splitting/internal/service"
)

type ChangeHandler struct {
	service *service.ChangeService
}

func NewChangeHandler(service *service.ChangeService) *ChangeHandler {
	return &ChangeHandler{service: service}
}

type submitChangeRequest struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedBy string `json:"created_by"`
}

func (h *ChangeHandler) SubmitChange(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")

	var req submitChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	change, err := h.service.SubmitChangeRequest(r.Context(), service.SubmitChangeInput{
		ProjectID: projectID,
		Title:     req.Title,
		Content:   req.Content,
		CreatedBy: req.CreatedBy,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, change)
}

func (h *ChangeHandler) ListChanges(w http.ResponseWriter, r *http.Request) {
	changes, err := h.service.ListChangeRequests(r.Context(), chi.URLParam(r, "project_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, changes)
}

func (h *ChangeHandler) AnalyzeChangeImpact(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	changeID := chi.URLParam(r, "id")

	if err := h.service.AnalyzeChangeImpact(r.Context(), projectID, changeID); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "change impact analysis completed"})
}

type updateChangeStatusRequest struct {
	Status string `json:"status"`
}

func (h *ChangeHandler) UpdateChangeStatus(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	changeID := chi.URLParam(r, "id")

	var req updateChangeStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := h.service.UpdateChangeStatus(r.Context(), projectID, changeID, domain.ChangeRequestStatus(req.Status)); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "change status updated"})
}
