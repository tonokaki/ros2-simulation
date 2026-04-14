package repository

import (
	"context"
	"database/sql"

	"github.com/takaki0/robotasker-backend/internal/model"
)

type RobotRepository struct {
	db *sql.DB
}

func NewRobotRepository(db *sql.DB) *RobotRepository {
	return &RobotRepository{db: db}
}

// List は全ロボットを取得する
func (r *RobotRepository) List(ctx context.Context) ([]model.Robot, error) {
	query := `SELECT id, name, status, current_location, battery_level, position_x, position_y, created_at, updated_at
		FROM robots ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var robots []model.Robot
	for rows.Next() {
		var rb model.Robot
		if err := rows.Scan(&rb.ID, &rb.Name, &rb.Status, &rb.CurrentLocation, &rb.BatteryLevel,
			&rb.PositionX, &rb.PositionY, &rb.CreatedAt, &rb.UpdatedAt); err != nil {
			return nil, err
		}
		robots = append(robots, rb)
	}
	return robots, rows.Err()
}

// GetByID はロボットをIDで取得する
func (r *RobotRepository) GetByID(ctx context.Context, id string) (*model.Robot, error) {
	query := `SELECT id, name, status, current_location, battery_level, position_x, position_y, created_at, updated_at
		FROM robots WHERE id = $1`
	var rb model.Robot
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rb.ID, &rb.Name, &rb.Status, &rb.CurrentLocation, &rb.BatteryLevel,
		&rb.PositionX, &rb.PositionY, &rb.CreatedAt, &rb.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rb, nil
}

// FindIdleRobot はアイドル状態のロボットを1台返す
func (r *RobotRepository) FindIdleRobot(ctx context.Context) (*model.Robot, error) {
	query := `SELECT id, name, status, current_location, battery_level, position_x, position_y, created_at, updated_at
		FROM robots WHERE status = 'idle' ORDER BY battery_level DESC LIMIT 1`
	var rb model.Robot
	err := r.db.QueryRowContext(ctx, query).Scan(
		&rb.ID, &rb.Name, &rb.Status, &rb.CurrentLocation, &rb.BatteryLevel,
		&rb.PositionX, &rb.PositionY, &rb.CreatedAt, &rb.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rb, nil
}

// UpdateStatus はロボットのステータスを更新する
func (r *RobotRepository) UpdateStatus(ctx context.Context, id string, status model.RobotStatus) error {
	query := `UPDATE robots SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

// UpdatePosition はロボットの位置を更新する
func (r *RobotRepository) UpdatePosition(ctx context.Context, id string, x, y float64, location string) error {
	query := `UPDATE robots SET position_x = $1, position_y = $2, current_location = $3, updated_at = NOW() WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, x, y, location, id)
	return err
}

// GetLocationByName は名前から地点を取得する
func (r *RobotRepository) GetLocationByName(ctx context.Context, name string) (*model.Location, error) {
	query := `SELECT id, name, x, y, floor, location_type FROM locations WHERE name = $1`
	var loc model.Location
	err := r.db.QueryRowContext(ctx, query, name).Scan(&loc.ID, &loc.Name, &loc.X, &loc.Y, &loc.Floor, &loc.LocationType)
	if err != nil {
		return nil, err
	}
	return &loc, nil
}

// ListLocations は全地点を取得する
func (r *RobotRepository) ListLocations(ctx context.Context) ([]model.Location, error) {
	query := `SELECT id, name, x, y, floor, location_type FROM locations ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []model.Location
	for rows.Next() {
		var loc model.Location
		if err := rows.Scan(&loc.ID, &loc.Name, &loc.X, &loc.Y, &loc.Floor, &loc.LocationType); err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}
	return locations, rows.Err()
}
