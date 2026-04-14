package handler

import (
	"encoding/json"
	"net/http"

	"github.com/takaki0/robotasker-backend/internal/model"
	"github.com/takaki0/robotasker-backend/internal/service"
)

type ChatHandler struct {
	taskService *service.TaskService
}

func NewChatHandler(taskService *service.TaskService) *ChatHandler {
	return &ChatHandler{taskService: taskService}
}

// HandleChat は自然言語入力を受け取り、タスクを生成する
func (h *ChatHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "リクエストの形式が不正です"})
		return
	}
	if req.Message == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "メッセージを入力してください"})
		return
	}

	resp, err := h.taskService.HandleChat(r.Context(), req.Message)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
