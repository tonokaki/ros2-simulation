package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/takaki0/robotasker-backend/internal/service"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// HandleList はタスク一覧を返す
func (h *TaskHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	tasks, err := h.taskService.ListTasks(r.Context(), limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

// HandleGet はタスク詳細を返す
func (h *TaskHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/v1/tasks/")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "タスクIDが必要です"})
		return
	}

	task, err := h.taskService.GetTask(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "タスクが見つかりません"})
		return
	}
	writeJSON(w, http.StatusOK, task)
}

// HandleCancel はタスクをキャンセルする
func (h *TaskHandler) HandleCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// /api/v1/tasks/{id}/cancel からidを抽出
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks/")
	parts := strings.Split(path, "/")
	if len(parts) < 1 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "タスクIDが必要です"})
		return
	}
	id := parts[0]

	if err := h.taskService.CancelTask(r.Context(), id); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "タスクをキャンセルしました"})
}

// HandleStats はダッシュボード統計を返す
func (h *TaskHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.taskService.GetStats(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, stats)
}
