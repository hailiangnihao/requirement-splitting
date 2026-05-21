package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"requirement-splitting/internal/domain"
	"requirement-splitting/internal/service"
)

// --- 返回给前端的树状结构定义 ---
type FeaturePointNode struct {
	domain.FeaturePoint
	Tasks     []domain.DevTask  `json:"tasks"`
	TestCases []domain.TestCase `json:"test_cases"`
}

type ModuleNode struct {
	domain.Module
	FeaturePoints []FeaturePointNode `json:"feature_points"`
}

type PlanTreeResponse struct {
	ProjectID       string                  `json:"project_id"`
	Modules         []ModuleNode            `json:"modules"`
	Milestones      []domain.Milestone      `json:"milestones"`
	AcceptanceItems []domain.AcceptanceItem `json:"acceptance_items"`
}

type PlanHandler struct {
	publishService *service.PlanPublishService
}

func NewPlanHandler(publishService *service.PlanPublishService) *PlanHandler {
	return &PlanHandler{
		publishService: publishService,
	}
}

func (h *PlanHandler) PublishDraft(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	draftID := chi.URLParam(r, "draft_id")

	if projectID == "" || draftID == "" {
		writeError(w, http.StatusBadRequest, "project_id and draft_id are required")
		return
	}

	if err := h.publishService.PublishDraft(r.Context(), projectID, draftID); err != nil {
		writeServiceError(w, err) // 使用统一的 service 错误处理
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "plan published successfully"})
}

func (h *PlanHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	if projectID == "" {
		writeError(w, http.StatusBadRequest, "project_id is required")
		return
	}

	plan, err := h.publishService.GetProjectPlan(r.Context(), projectID)
	if err != nil {
		writeServiceError(w, err) // 使用统一的 service 错误处理
		return
	}

	// 构建 O(N) 查询字典，避免多重嵌套循环
	tasksByFP := make(map[string][]domain.DevTask)
	for _, task := range plan.DevTasks {
		tasksByFP[task.FeaturePointID] = append(tasksByFP[task.FeaturePointID], task)
	}

	tcByFP := make(map[string][]domain.TestCase)
	for _, tc := range plan.TestCases {
		tcByFP[tc.FeaturePointID] = append(tcByFP[tc.FeaturePointID], tc)
	}

	fpByMod := make(map[string][]FeaturePointNode)
	for _, fp := range plan.FeaturePoints {
		tasks := tasksByFP[fp.ID]
		if tasks == nil {
			tasks = []domain.DevTask{}
		} // 确保 JSON 返回 [] 而不是 null

		tcs := tcByFP[fp.ID]
		if tcs == nil {
			tcs = []domain.TestCase{}
		}

		fpByMod[fp.ModuleID] = append(fpByMod[fp.ModuleID], FeaturePointNode{
			FeaturePoint: fp,
			Tasks:        tasks,
			TestCases:    tcs,
		})
	}

	// 组装最终响应
	resp := PlanTreeResponse{
		ProjectID:       projectID,
		Modules:         []ModuleNode{},
		Milestones:      plan.Milestones,
		AcceptanceItems: plan.AcceptanceItems,
	}
	if resp.Milestones == nil {
		resp.Milestones = []domain.Milestone{}
	}
	if resp.AcceptanceItems == nil {
		resp.AcceptanceItems = []domain.AcceptanceItem{}
	}

	for _, mod := range plan.Modules {
		fps := fpByMod[mod.ID]
		if fps == nil {
			fps = []FeaturePointNode{}
		}
		resp.Modules = append(resp.Modules, ModuleNode{Module: mod, FeaturePoints: fps})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *PlanHandler) ListDevTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.publishService.ListDevTasks(r.Context(), chi.URLParam(r, "project_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *PlanHandler) ListTestCases(w http.ResponseWriter, r *http.Request) {
	testCases, err := h.publishService.ListTestCases(r.Context(), chi.URLParam(r, "project_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, testCases)
}

type updateDevTaskStatusRequest struct {
	Status string `json:"status"`
}

func (h *PlanHandler) UpdateDevTaskStatus(w http.ResponseWriter, r *http.Request) {
	var req updateDevTaskStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if err := h.publishService.UpdateDevTaskStatus(r.Context(), chi.URLParam(r, "project_id"), chi.URLParam(r, "id"), domain.TaskStatus(req.Status)); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "dev task status updated"})
}
