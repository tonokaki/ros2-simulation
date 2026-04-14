package model

import "time"

// RobotStatus はロボットの状態
type RobotStatus string

const (
	RobotStatusIdle      RobotStatus = "idle"
	RobotStatusMoving    RobotStatus = "moving"
	RobotStatusExecuting RobotStatus = "executing"
	RobotStatusCharging  RobotStatus = "charging"
	RobotStatusError     RobotStatus = "error"
)

// Robot はロボットの状態情報
type Robot struct {
	ID              string      `json:"id" db:"id"`
	Name            string      `json:"name" db:"name"`
	Status          RobotStatus `json:"status" db:"status"`
	CurrentLocation *string     `json:"current_location,omitempty" db:"current_location"`
	BatteryLevel    int         `json:"battery_level" db:"battery_level"`
	PositionX       float64     `json:"position_x" db:"position_x"`
	PositionY       float64     `json:"position_y" db:"position_y"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" db:"updated_at"`
}

// Location はロボットが移動できる地点
type Location struct {
	ID           string  `json:"id" db:"id"`
	Name         string  `json:"name" db:"name"`
	X            float64 `json:"x" db:"x"`
	Y            float64 `json:"y" db:"y"`
	Floor        string  `json:"floor" db:"floor"`
	LocationType string  `json:"location_type" db:"location_type"`
}

// DashboardStats はダッシュボード統計
type DashboardStats struct {
	TotalTasks     int     `json:"total_tasks"`
	CompletedTasks int     `json:"completed_tasks"`
	PendingTasks   int     `json:"pending_tasks"`
	FailedTasks    int     `json:"failed_tasks"`
	ActiveRobots   int     `json:"active_robots"`
	AvgDuration    float64 `json:"avg_duration_seconds"`
}
