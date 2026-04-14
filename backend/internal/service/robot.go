package service

import (
	"context"

	"github.com/takaki0/robotasker-backend/internal/model"
	"github.com/takaki0/robotasker-backend/internal/repository"
)

// RobotService はロボット管理のビジネスロジック
type RobotService struct {
	robotRepo *repository.RobotRepository
}

func NewRobotService(robotRepo *repository.RobotRepository) *RobotService {
	return &RobotService{robotRepo: robotRepo}
}

// List は全ロボットを取得する
func (s *RobotService) List(ctx context.Context) ([]model.Robot, error) {
	return s.robotRepo.List(ctx)
}

// GetByID はロボットをIDで取得する
func (s *RobotService) GetByID(ctx context.Context, id string) (*model.Robot, error) {
	return s.robotRepo.GetByID(ctx, id)
}

// ListLocations は全地点を取得する
func (s *RobotService) ListLocations(ctx context.Context) ([]model.Location, error) {
	return s.robotRepo.ListLocations(ctx)
}

// UpdatePosition はロボットの位置を更新する（ROS2からのコールバック用）
func (s *RobotService) UpdatePosition(ctx context.Context, robotID string, x, y float64, location string) error {
	return s.robotRepo.UpdatePosition(ctx, robotID, x, y, location)
}
