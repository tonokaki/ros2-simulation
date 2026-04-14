package rosbridge

import (
	"context"
	"log"
	"math"
	"sync"
	"time"
)

// Client はROS2 rosbridge WebSocketクライアントのインターフェース
type Client interface {
	SendTaskCommand(ctx context.Context, taskID, robotID, action, targetLocation string) error
	// OnTaskComplete はタスク完了時のコールバックを登録する
	SetOnTaskComplete(fn func(taskID string, success bool, message string))
	SetOnRobotState(fn func(robotID string, x, y float64, location string))
	Close()
}

// MockClient はROS2接続のモック実装（タスクを受信→一定時間後に完了を返す）
type MockClient struct {
	mu              sync.Mutex
	onTaskComplete  func(taskID string, success bool, message string)
	onRobotState    func(robotID string, x, y float64, location string)
}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (c *MockClient) SetOnTaskComplete(fn func(taskID string, success bool, message string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onTaskComplete = fn
}

func (c *MockClient) SetOnRobotState(fn func(robotID string, x, y float64, location string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onRobotState = fn
}

// SendTaskCommand はタスクをROS2に送信する（モック: goroutineで移動をシミュレート）
func (c *MockClient) SendTaskCommand(ctx context.Context, taskID, robotID, action, targetLocation string) error {
	log.Printf("[MockROS] タスク送信: task=%s robot=%s action=%s target=%s", taskID[:8], robotID[:8], action, targetLocation)

	// シミュレーション用の地点座標（seed.sqlと一致）
	locationCoords := map[string][2]float64{
		"充電ステーション": {0.0, 0.0},
		"受付":           {5.0, 0.0},
		"会議室A":        {3.0, 4.0},
		"会議室B":        {7.0, 4.0},
		"休憩室":         {5.0, 8.0},
		"倉庫":           {0.0, 8.0},
		"エントランス":    {10.0, 0.0},
	}

	targetCoord, ok := locationCoords[targetLocation]
	if !ok {
		targetCoord = [2]float64{5.0, 5.0}
	}

	// バックグラウンドで移動シミュレーション
	go func() {
		steps := 10
		startX, startY := 0.0, 0.0 // 簡易: 原点から出発
		dx := (targetCoord[0] - startX) / float64(steps)
		dy := (targetCoord[1] - startY) / float64(steps)

		for i := 1; i <= steps; i++ {
			time.Sleep(500 * time.Millisecond)
			x := startX + dx*float64(i)
			y := startY + dy*float64(i)

			c.mu.Lock()
			if c.onRobotState != nil {
				loc := ""
				// 最終ステップで到着地点名を設定
				if i == steps {
					loc = targetLocation
				}
				c.onRobotState(robotID, roundTo(x, 2), roundTo(y, 2), loc)
			}
			c.mu.Unlock()
		}

		// 到着後、タスク完了を通知
		time.Sleep(500 * time.Millisecond)
		c.mu.Lock()
		if c.onTaskComplete != nil {
			c.onTaskComplete(taskID, true, targetLocation+"に到着しました")
		}
		c.mu.Unlock()
		log.Printf("[MockROS] タスク完了: task=%s", taskID[:8])
	}()

	return nil
}

func (c *MockClient) Close() {
	log.Println("[MockROS] クライアント終了")
}

func roundTo(val float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(val*pow) / pow
}
