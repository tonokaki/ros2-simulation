package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/takaki0/robotasker-backend/internal/model"
	"github.com/takaki0/robotasker-backend/internal/repository"
)

// TaskService はタスク管理のビジネスロジック
type TaskService struct {
	taskRepo  *repository.TaskRepository
	robotRepo *repository.RobotRepository
	llm       LLMService
	ros       ROSBridgeService
}

func NewTaskService(
	taskRepo *repository.TaskRepository,
	robotRepo *repository.RobotRepository,
	llm LLMService,
	ros ROSBridgeService,
) *TaskService {
	return &TaskService{
		taskRepo:  taskRepo,
		robotRepo: robotRepo,
		llm:       llm,
		ros:       ros,
	}
}

// HandleChat は自然言語入力を処理し、タスクを生成する
func (s *TaskService) HandleChat(ctx context.Context, message string) (*model.ChatResponse, error) {
	// 1. LLMで自然言語を解析
	parsed, err := s.llm.ParseCommand(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("LLM解析エラー: %w", err)
	}

	// 信頼度が低い場合は確認を返す
	if parsed.Confidence < 0.5 {
		return &model.ChatResponse{
			Reply:      "すみません、指示を正しく理解できませんでした。目的地とアクションをもう少し具体的に教えてください。",
			ParsedTask: parsed,
		}, nil
	}

	// 2. 地点をDBから解決
	var locationID *string
	if parsed.TargetLocation != "" {
		loc, err := s.robotRepo.GetLocationByName(ctx, parsed.TargetLocation)
		if err != nil {
			if err == sql.ErrNoRows {
				return &model.ChatResponse{
					Reply:      fmt.Sprintf("「%s」という場所が見つかりません。登録されている場所を確認してください。", parsed.TargetLocation),
					ParsedTask: parsed,
				}, nil
			}
			return nil, err
		}
		locationID = &loc.ID
	}

	// 3. タスクを作成
	priority := 0
	if parsed.Priority == "high" {
		priority = 10
	}
	task := &model.Task{
		OriginalText:     message,
		ParsedAction:     parsed.Action,
		TargetLocationID: locationID,
		Priority:         priority,
		Status:           model.TaskStatusPending,
	}
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("タスク作成エラー: %w", err)
	}

	// 4. ロボットを割り当て
	reply := s.assignAndDispatch(ctx, task, parsed)

	return &model.ChatResponse{
		Reply:      reply,
		Task:       task,
		ParsedTask: parsed,
	}, nil
}

// assignAndDispatch はロボットを割り当ててタスクを送信する
func (s *TaskService) assignAndDispatch(ctx context.Context, task *model.Task, parsed *model.ParsedTask) string {
	robot, err := s.robotRepo.FindIdleRobot(ctx)
	if err != nil {
		log.Printf("アイドルロボットなし: %v", err)
		return fmt.Sprintf("タスクを受け付けました（ID: %s）。現在ロボットが空いていないため、キューに入りました。", task.ID[:8])
	}

	// ロボット割り当て
	if err := s.taskRepo.AssignRobot(ctx, task.ID, robot.ID); err != nil {
		log.Printf("ロボット割り当てエラー: %v", err)
		return "タスクを受け付けましたが、ロボットの割り当てに失敗しました。"
	}

	// ロボットステータスを更新
	s.robotRepo.UpdateStatus(ctx, robot.ID, model.RobotStatusMoving)

	// ステータスを in_progress に更新
	s.taskRepo.UpdateStatus(ctx, task.ID, model.TaskStatusAssigned, model.TaskStatusInProgress, "")

	// ROS2にタスク送信
	if err := s.ros.SendTaskCommand(ctx, task.ID, robot.ID, string(parsed.Action), parsed.TargetLocation); err != nil {
		log.Printf("ROS送信エラー: %v", err)
	}

	actionLabel := map[model.TaskAction]string{
		model.TaskActionDeliver: "配達",
		model.TaskActionPatrol:  "巡回",
		model.TaskActionReturn:  "帰還",
		model.TaskActionGoto:    "移動",
	}
	label := actionLabel[parsed.Action]
	return fmt.Sprintf("%sが「%s」への%sタスクを開始しました。（タスクID: %s）",
		robot.Name, parsed.TargetLocation, label, task.ID[:8])
}

// GetTask はタスクをIDで取得する
func (s *TaskService) GetTask(ctx context.Context, id string) (*model.Task, error) {
	return s.taskRepo.GetByID(ctx, id)
}

// ListTasks は全タスクを取得する
func (s *TaskService) ListTasks(ctx context.Context, limit int) ([]model.Task, error) {
	return s.taskRepo.List(ctx, limit)
}

// CancelTask はタスクをキャンセルする
func (s *TaskService) CancelTask(ctx context.Context, id string) error {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if task.Status == model.TaskStatusCompleted || task.Status == model.TaskStatusCancelled {
		return fmt.Errorf("タスクは既に%sです", task.Status)
	}

	// ロボットをidle に戻す
	if task.RobotID != nil {
		s.robotRepo.UpdateStatus(ctx, *task.RobotID, model.RobotStatusIdle)
	}

	return s.taskRepo.UpdateStatus(ctx, id, task.Status, model.TaskStatusCancelled, "ユーザーによりキャンセル")
}

// CompleteTask はタスクを完了にする（ROS2からのコールバック用）
func (s *TaskService) CompleteTask(ctx context.Context, taskID string, success bool, message string) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	newStatus := model.TaskStatusCompleted
	if !success {
		newStatus = model.TaskStatusFailed
	}

	// ロボットをidleに戻す
	if task.RobotID != nil {
		s.robotRepo.UpdateStatus(ctx, *task.RobotID, model.RobotStatusIdle)
	}

	return s.taskRepo.UpdateStatus(ctx, taskID, task.Status, newStatus, message)
}

// GetStats はダッシュボード統計を取得する
func (s *TaskService) GetStats(ctx context.Context) (*model.DashboardStats, error) {
	return s.taskRepo.GetStats(ctx)
}
