package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"requirement-splitting/internal/http/handlers"
	"requirement-splitting/internal/service"
)

func NewRouter(projectService *service.ProjectService, aiDraftService *service.AIDraftService, planPublishService *service.PlanPublishService, testService *service.TestService, defectService *service.DefectService) http.Handler {
	router := chi.NewRouter()
	projectHandler := handlers.NewProjectHandler(projectService)
	aiHandler := handlers.NewAIHandler(aiDraftService)
	planHandler := handlers.NewPlanHandler(planPublishService)
	testHandler := handlers.NewTestHandler(testService)
	defectHandler := handlers.NewDefectHandler(defectService)

	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	projectHandler.RegisterRoutes(router)
	aiHandler.RegisterRoutes(router)

	// 注册计划相关接口
	router.Post("/api/projects/{project_id}/ai-drafts/{draft_id}/publish", planHandler.PublishDraft)
	router.Get("/api/projects/{project_id}/plan", planHandler.GetPlan)

	// 注册测试相关接口
	router.Post("/api/projects/{project_id}/test-cases/{id}/confirm", testHandler.ConfirmTestCase)
	router.Post("/api/projects/{project_id}/test-cases/{id}/ai-run", testHandler.RunAITest)
	router.Post("/api/projects/{project_id}/test-runs/{id}/review", testHandler.ReviewTestRun)

	// 注册缺陷相关接口
	router.Post("/api/projects/{project_id}/defects", defectHandler.CreateDefect)
	router.Patch("/api/projects/{project_id}/defects/{id}/status", defectHandler.UpdateDefectStatus)

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
