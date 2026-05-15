package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"requirement-splitting/internal/http/handlers"
	"requirement-splitting/internal/service"
)

func NewRouter(projectService *service.ProjectService, aiDraftService *service.AIDraftService) http.Handler {
	router := chi.NewRouter()
	projectHandler := handlers.NewProjectHandler(projectService)
	aiHandler := handlers.NewAIHandler(aiDraftService)

	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	projectHandler.RegisterRoutes(router)
	aiHandler.RegisterRoutes(router)

	return router
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
