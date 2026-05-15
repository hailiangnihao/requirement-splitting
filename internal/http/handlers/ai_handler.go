package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"requirement-splitting/internal/service"
)

type AIHandler struct {
	service *service.AIDraftService
}

func NewAIHandler(service *service.AIDraftService) *AIHandler {
	return &AIHandler{service: service}
}

func (h *AIHandler) RegisterRoutes(router chi.Router) {
	router.Post("/api/projects/{id}/ai/split-requirement", h.splitRequirement)
	router.Get("/api/projects/{id}/ai-drafts", h.listDrafts)
}

type splitRequirementRequest struct {
	RequirementID string `json:"requirement_id"`
	Content       string `json:"content"`
	CreatedBy     string `json:"created_by"`
}

func (h *AIHandler) splitRequirement(w http.ResponseWriter, r *http.Request) {
	var req splitRequirementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	draft, err := h.service.SplitRequirement(r.Context(), service.SplitRequirementInput{
		ProjectID:     chi.URLParam(r, "id"),
		RequirementID: req.RequirementID,
		Content:       req.Content,
		CreatedBy:     req.CreatedBy,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, draft)
}

func (h *AIHandler) listDrafts(w http.ResponseWriter, r *http.Request) {
	drafts, err := h.service.ListDrafts(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, drafts)
}
