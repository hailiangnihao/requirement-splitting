package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"requirement-splitting/internal/domain"
	"requirement-splitting/internal/service"
)

type TestHandler struct {
	service *service.TestService
}

func NewTestHandler(service *service.TestService) *TestHandler {
	return &TestHandler{service: service}
}

func (h *TestHandler) ConfirmTestCase(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	testCaseID := chi.URLParam(r, "id")

	if err := h.service.ConfirmTestCase(r.Context(), projectID, testCaseID); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "test case confirmed"})
}

func (h *TestHandler) ListTestRuns(w http.ResponseWriter, r *http.Request) {
	testRuns, err := h.service.ListTestRuns(r.Context(), chi.URLParam(r, "project_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, testRuns)
}

func (h *TestHandler) RunAITest(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	testCaseID := chi.URLParam(r, "id")

	testRun, err := h.service.RunAITest(r.Context(), projectID, testCaseID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, testRun)
}

type reviewTestRunRequest struct {
	Status string `json:"status"`
}

func (h *TestHandler) ReviewTestRun(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	testRunID := chi.URLParam(r, "id")

	var req reviewTestRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := h.service.ReviewTestRun(r.Context(), projectID, testRunID, domain.TestRunReviewStatus(req.Status)); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "test run reviewed"})
}
