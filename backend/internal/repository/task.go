package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/takaki0/robotasker-backend/internal/model"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// Create は新しいタスクを作成する
func (r *TaskRepository) Create(ctx context.Context, task *model.Task) error {
	query := `
		INSERT INTO tasks (original_text, parsed_action, target_location_id, priority, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query,
		task.OriginalText, task.ParsedAction, task.TargetLocationID, task.Priority, task.Status,
	).Scan(&task.ID, &task.CreatedAt)
}

// GetByID はタスクをIDで取得する
func (r *TaskRepository) GetByID(ctx context.Context, id string) (*model.Task, error) {
	query := `
		SELECT t.id, t.robot_id, t.original_text, t.parsed_action, t.target_location_id,
		       t.priority, t.status, t.result_message, t.created_at, t.started_at, t.completed_at,
		       COALESCE(r.name, '') as robot_name,
		       COALESCE(l.name, '') as target_location_name
		FROM tasks t
		LEFT JOIN robots r ON t.robot_id = r.id
		LEFT JOIN locations l ON t.target_location_id = l.id
		WHERE t.id = $1`
	var task model.Task
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.RobotID, &task.OriginalText, &task.ParsedAction, &task.TargetLocationID,
		&task.Priority, &task.Status, &task.ResultMessage, &task.CreatedAt, &task.StartedAt, &task.CompletedAt,
		&task.RobotName, &task.TargetLocationName,
	)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// List は全タスクを新しい順に取得する
func (r *TaskRepository) List(ctx context.Context, limit int) ([]model.Task, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT t.id, t.robot_id, t.original_text, t.parsed_action, t.target_location_id,
		       t.priority, t.status, t.result_message, t.created_at, t.started_at, t.completed_at,
		       COALESCE(r.name, '') as robot_name,
		       COALESCE(l.name, '') as target_location_name
		FROM tasks t
		LEFT JOIN robots r ON t.robot_id = r.id
		LEFT JOIN locations l ON t.target_location_id = l.id
		ORDER BY t.created_at DESC
		LIMIT $1`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(
			&t.ID, &t.RobotID, &t.OriginalText, &t.ParsedAction, &t.TargetLocationID,
			&t.Priority, &t.Status, &t.ResultMessage, &t.CreatedAt, &t.StartedAt, &t.CompletedAt,
			&t.RobotName, &t.TargetLocationName,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// UpdateStatus はタスクのステータスを更新し、ログを記録する
func (r *TaskRepository) UpdateStatus(ctx context.Context, id string, fromStatus, toStatus model.TaskStatus, message string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// ステータス更新
	updateQuery := `UPDATE tasks SET status = $1`
	args := []any{toStatus}
	argIdx := 2

	switch toStatus {
	case model.TaskStatusInProgress:
		updateQuery += `, started_at = NOW()`
	case model.TaskStatusCompleted, model.TaskStatusFailed:
		updateQuery += `, completed_at = NOW()`
		if message != "" {
			updateQuery += `, result_message = $` + itoa(argIdx)
			args = append(args, message)
			argIdx++
		}
	}
	updateQuery += ` WHERE id = $` + itoa(argIdx)
	args = append(args, id)

	if _, err := tx.ExecContext(ctx, updateQuery, args...); err != nil {
		return err
	}

	// ログ記録
	var msgPtr *string
	if message != "" {
		msgPtr = &message
	}
	var fromPtr *string
	if fromStatus != "" {
		s := string(fromStatus)
		fromPtr = &s
	}
	logQuery := `INSERT INTO task_logs (task_id, from_status, to_status, message) VALUES ($1, $2, $3, $4)`
	if _, err := tx.ExecContext(ctx, logQuery, id, fromPtr, toStatus, msgPtr); err != nil {
		return err
	}

	return tx.Commit()
}

// AssignRobot はタスクにロボットを割り当てる
func (r *TaskRepository) AssignRobot(ctx context.Context, taskID, robotID string) error {
	query := `UPDATE tasks SET robot_id = $1, status = 'assigned' WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, robotID, taskID)
	return err
}

// GetStats はダッシュボード統計を取得する
func (r *TaskRepository) GetStats(ctx context.Context) (*model.DashboardStats, error) {
	stats := &model.DashboardStats{}
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'completed') as completed,
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COUNT(*) FILTER (WHERE status = 'failed') as failed,
			COALESCE(AVG(EXTRACT(EPOCH FROM (completed_at - started_at))) FILTER (WHERE completed_at IS NOT NULL AND started_at IS NOT NULL), 0) as avg_duration
		FROM tasks`
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalTasks, &stats.CompletedTasks, &stats.PendingTasks, &stats.FailedTasks, &stats.AvgDuration,
	)
	if err != nil {
		return nil, err
	}

	// アクティブロボット数
	robotQuery := `SELECT COUNT(*) FROM robots WHERE status != 'idle' AND status != 'error'`
	err = r.db.QueryRowContext(ctx, robotQuery).Scan(&stats.ActiveRobots)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// FindPendingTasks は割り当て待ちタスクを取得する
func (r *TaskRepository) FindPendingTasks(ctx context.Context) ([]model.Task, error) {
	query := `SELECT id, original_text, parsed_action, target_location_id, priority, status, created_at
		FROM tasks WHERE status = 'pending' ORDER BY priority DESC, created_at ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.OriginalText, &t.ParsedAction, &t.TargetLocationID, &t.Priority, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// itoa は数字を文字列に変換する（fmtインポート回避）
func itoa(i int) string {
	_ = time.Now() // timeパッケージ利用のため
	digits := "0123456789"
	if i < 10 {
		return string(digits[i])
	}
	return string(digits[i/10]) + string(digits[i%10])
}
