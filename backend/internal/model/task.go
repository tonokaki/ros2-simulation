package model

import "time"

// TaskStatus はタスクの状態遷移を表す
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusAssigned   TaskStatus = "assigned"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// TaskAction はロボットが実行するアクション種別
type TaskAction string

const (
	TaskActionDeliver TaskAction = "deliver"
	TaskActionPatrol  TaskAction = "patrol"
	TaskActionReturn  TaskAction = "return"
	TaskActionGoto    TaskAction = "goto"
)

// Task はロボットに割り当てるタスク
type Task struct {
	ID               string     `json:"id" db:"id"`
	RobotID          *string    `json:"robot_id,omitempty" db:"robot_id"`
	OriginalText     string     `json:"original_text" db:"original_text"`
	ParsedAction     TaskAction `json:"parsed_action" db:"parsed_action"`
	TargetLocationID *string    `json:"target_location_id,omitempty" db:"target_location_id"`
	Priority         int        `json:"priority" db:"priority"`
	Status           TaskStatus `json:"status" db:"status"`
	ResultMessage    *string    `json:"result_message,omitempty" db:"result_message"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	StartedAt        *time.Time `json:"started_at,omitempty" db:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty" db:"completed_at"`

	// JOIN用フィールド（DB列ではない）
	RobotName          string `json:"robot_name,omitempty" db:"robot_name"`
	TargetLocationName string `json:"target_location_name,omitempty" db:"target_location_name"`
}

// TaskLog はタスクの状態遷移履歴
type TaskLog struct {
	ID         string     `json:"id" db:"id"`
	TaskID     string     `json:"task_id" db:"task_id"`
	FromStatus *string    `json:"from_status,omitempty" db:"from_status"`
	ToStatus   string     `json:"to_status" db:"to_status"`
	Message    *string    `json:"message,omitempty" db:"message"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// ParsedTask はLLMが自然言語から解析した結果
type ParsedTask struct {
	Action         TaskAction `json:"action"`
	TargetLocation string     `json:"target_location"`
	Item           string     `json:"item,omitempty"`
	Priority       string     `json:"priority"`
	Confidence     float64    `json:"confidence"`
}

// ChatRequest はチャットAPIのリクエスト
type ChatRequest struct {
	Message string `json:"message"`
}

// ChatResponse はチャットAPIのレスポンス
type ChatResponse struct {
	Reply      string      `json:"reply"`
	Task       *Task       `json:"task,omitempty"`
	ParsedTask *ParsedTask `json:"parsed_task,omitempty"`
}
