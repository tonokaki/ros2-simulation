package rosbridge

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSMessage はbridge_nodeとの通信メッセージ
type WSMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// TaskCommandData はタスク指令データ
type TaskCommandData struct {
	TaskID         string  `json:"task_id"`
	RobotID        string  `json:"robot_id"`
	Action         string  `json:"action"`
	TargetLocation string  `json:"target_location"`
	TargetX        float64 `json:"target_x"`
	TargetY        float64 `json:"target_y"`
	Priority       int     `json:"priority"`
}

// TaskStatusData はタスク状態データ
type TaskStatusData struct {
	TaskID  string `json:"task_id"`
	RobotID string `json:"robot_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// RobotStateData はロボット状態データ
type RobotStateData struct {
	RobotID         string  `json:"robot_id"`
	Status          string  `json:"status"`
	PositionX       float64 `json:"position_x"`
	PositionY       float64 `json:"position_y"`
	Velocity        float64 `json:"velocity"`
	CurrentLocation string  `json:"current_location"`
	BatteryLevel    int     `json:"battery_level"`
}

// WSClient はROS2 bridge_nodeへのWebSocket接続クライアント
type WSClient struct {
	url            string
	conn           *websocket.Conn
	mu             sync.Mutex
	onTaskComplete func(taskID string, success bool, message string)
	onRobotState   func(robotID string, x, y float64, location string)
	done           chan struct{}
	// 地点座標マップ（タスク指令時に座標を付加するため）
	locationCoords map[string][2]float64
}

// NewWSClient はWebSocketクライアントを生成する
func NewWSClient(url string) *WSClient {
	return &WSClient{
		url:  url,
		done: make(chan struct{}),
		locationCoords: map[string][2]float64{
			"充電ステーション": {0.0, 0.0},
			"受付":           {5.0, 0.0},
			"会議室A":        {3.0, 4.0},
			"会議室B":        {7.0, 4.0},
			"休憩室":         {5.0, 8.0},
			"倉庫":           {0.0, 8.0},
			"エントランス":    {10.0, 0.0},
		},
	}
}

func (c *WSClient) SetOnTaskComplete(fn func(taskID string, success bool, message string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onTaskComplete = fn
}

func (c *WSClient) SetOnRobotState(fn func(robotID string, x, y float64, location string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onRobotState = fn
}

// Connect はbridge_nodeに接続し、受信ループを開始する
func (c *WSClient) Connect() error {
	var err error
	for i := 0; i < 30; i++ {
		c.conn, _, err = websocket.DefaultDialer.Dial(c.url, nil)
		if err == nil {
			log.Printf("[ROSBridge] 接続成功: %s", c.url)
			go c.receiveLoop()
			return nil
		}
		log.Printf("[ROSBridge] 接続待機中... (%d/30): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("ROSBridge接続失敗: %w", err)
}

// receiveLoop はbridge_nodeからのメッセージを受信し続ける
func (c *WSClient) receiveLoop() {
	defer close(c.done)
	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("[ROSBridge] 受信エラー（再接続試行）: %v", err)
			c.reconnect()
			return
		}

		var msg WSMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Printf("[ROSBridge] JSONパースエラー: %v", err)
			continue
		}

		switch msg.Type {
		case "task_status":
			c.handleTaskStatus(msg.Data)
		case "robot_state":
			c.handleRobotState(msg.Data)
		default:
			log.Printf("[ROSBridge] 不明なメッセージタイプ: %s", msg.Type)
		}
	}
}

func (c *WSClient) handleTaskStatus(data json.RawMessage) {
	var status TaskStatusData
	if err := json.Unmarshal(data, &status); err != nil {
		log.Printf("[ROSBridge] TaskStatusパースエラー: %v", err)
		return
	}

	c.mu.Lock()
	fn := c.onTaskComplete
	c.mu.Unlock()

	if fn != nil {
		success := status.Status == "completed"
		fn(status.TaskID, success, status.Message)
	}
	log.Printf("[ROSBridge] タスク状態受信: task=%s status=%s", status.TaskID[:8], status.Status)
}

func (c *WSClient) handleRobotState(data json.RawMessage) {
	var state RobotStateData
	if err := json.Unmarshal(data, &state); err != nil {
		log.Printf("[ROSBridge] RobotStateパースエラー: %v", err)
		return
	}

	c.mu.Lock()
	fn := c.onRobotState
	c.mu.Unlock()

	if fn != nil {
		fn(state.RobotID, state.PositionX, state.PositionY, state.CurrentLocation)
	}
}

func (c *WSClient) reconnect() {
	log.Println("[ROSBridge] 再接続中...")
	for i := 0; i < 10; i++ {
		time.Sleep(3 * time.Second)
		conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
		if err == nil {
			c.mu.Lock()
			c.conn = conn
			c.mu.Unlock()
			log.Println("[ROSBridge] 再接続成功")
			c.done = make(chan struct{})
			go c.receiveLoop()
			return
		}
		log.Printf("[ROSBridge] 再接続失敗 (%d/10): %v", i+1, err)
	}
	log.Println("[ROSBridge] 再接続を断念")
}

// SendTaskCommand はタスク指令をbridge_nodeに送信する
func (c *WSClient) SendTaskCommand(ctx context.Context, taskID, robotID, action, targetLocation string) error {
	coords, ok := c.locationCoords[targetLocation]
	if !ok {
		coords = [2]float64{5.0, 5.0}
	}

	cmdData := TaskCommandData{
		TaskID:         taskID,
		RobotID:        robotID,
		Action:         action,
		TargetLocation: targetLocation,
		TargetX:        coords[0],
		TargetY:        coords[1],
	}

	data, err := json.Marshal(cmdData)
	if err != nil {
		return fmt.Errorf("タスク指令シリアライズエラー: %w", err)
	}

	msg := WSMessage{
		Type: "task_command",
		Data: data,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("メッセージシリアライズエラー: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return fmt.Errorf("WebSocket未接続")
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
		return fmt.Errorf("WebSocket送信エラー: %w", err)
	}

	log.Printf("[ROSBridge] タスク指令送信: task=%s action=%s target=%s", taskID[:8], action, targetLocation)
	return nil
}

// Close はWebSocket接続を閉じる
func (c *WSClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.conn.Close()
	}
	log.Println("[ROSBridge] クライアント終了")
}
