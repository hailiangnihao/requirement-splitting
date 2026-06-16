package handlers

import (
	"encoding/json"
	"fmt"
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
	router.Post("/api/projects/{id}/ai/split-requirement/stream", h.splitRequirementStream) // 新增流式endpoint
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

// splitRequirementStream 流式拆分需求 (SSE)
func (h *AIHandler) splitRequirementStream(w http.ResponseWriter, r *http.Request) {
	var req splitRequirementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	// 设置 SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	// 进度回调函数
	progressCallback := func(progress interface{}) {
		data, _ := json.Marshal(progress)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	// 调用流式服务
	draft, err := h.service.SplitRequirementStream(r.Context(), service.SplitRequirementInput{
		ProjectID:     chi.URLParam(r, "id"),
		RequirementID: req.RequirementID,
		Content:       req.Content,
		CreatedBy:     req.CreatedBy,
	}, progressCallback)

	if err != nil {
		errorData, _ := json.Marshal(map[string]string{
			"type":    "error",
			"message": err.Error(),
		})
		fmt.Fprintf(w, "data: %s\n\n", errorData)
		flusher.Flush()
		return
	}

	// 发送最终结果
	resultData, _ := json.Marshal(map[string]interface{}{
		"type":    "complete",
		"message": "需求拆分完成",
		"data":    draft,
	})
	fmt.Fprintf(w, "data: %s\n\n", resultData)
	flusher.Flush()
}
