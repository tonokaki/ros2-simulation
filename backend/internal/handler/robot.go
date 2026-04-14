package handler

import (
	"net/http"

	"github.com/takaki0/robotasker-backend/internal/service"
)

type RobotHandler struct {
	robotService *service.RobotService
}

func NewRobotHandler(robotService *service.RobotService) *RobotHandler {
	return &RobotHandler{robotService: robotService}
}

// HandleList はロボット一覧を返す
func (h *RobotHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	robots, err := h.robotService.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, robots)
}

// HandleGet はロボット詳細を返す
func (h *RobotHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/v1/robots/")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ロボットIDが必要です"})
		return
	}

	robot, err := h.robotService.GetByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "ロボットが見つかりません"})
		return
	}
	writeJSON(w, http.StatusOK, robot)
}

// HandleLocations は全地点を返す
func (h *RobotHandler) HandleLocations(w http.ResponseWriter, r *http.Request) {
	locations, err := h.robotService.ListLocations(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, locations)
}
