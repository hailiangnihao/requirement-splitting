package handlers

import (
	"net/http"

	"requirement-splitting/internal/service"

	"github.com/go-chi/chi/v5"
)

type HealthHandler struct {
	service *service.HealthService
}

func NewHealthHandler(service *service.HealthService) *HealthHandler {
	return &HealthHandler{service: service}
}

func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	health, err := h.service.GetProjectHealth(r.Context(), chi.URLParam(r, "project_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, health)
}
