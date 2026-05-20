package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"requirement-splitting/internal/http/handlers"
	"requirement-splitting/internal/service"
)

func NewRouter(projectService *service.ProjectService, aiDraftService *service.AIDraftService, planPublishService *service.PlanPublishService) http.Handler {
	router := chi.NewRouter()
	projectHandler := handlers.NewProjectHandler(projectService)
	aiHandler := handlers.NewAIHandler(aiDraftService)
	planHandler := handlers.NewPlanHandler(planPublishService)

	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	projectHandler.RegisterRoutes(router)
	aiHandler.RegisterRoutes(router)

	// 注册计划相关接口
	router.Post("/api/projects/{project_id}/ai-drafts/{draft_id}/publish", planHandler.PublishDraft)
	router.Get("/api/projects/{project_id}/plan", planHandler.GetPlan)

	return router
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
