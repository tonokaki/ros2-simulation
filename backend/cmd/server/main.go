package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/takaki0/robotasker-backend/internal/config"
	"github.com/takaki0/robotasker-backend/internal/handler"
	"github.com/takaki0/robotasker-backend/internal/repository"
	"github.com/takaki0/robotasker-backend/internal/rosbridge"
	"github.com/takaki0/robotasker-backend/internal/service"
)

func main() {
	cfg := config.Load()

	// DB接続
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		log.Fatalf("DB接続エラー: %v", err)
	}
	defer db.Close()

	// 接続確認（リトライ付き）
	for i := 0; i < 30; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		log.Printf("DB接続待機中... (%d/30)", i+1)
		time.Sleep(time.Second)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("DB接続失敗: %v", err)
	}
	log.Println("DB接続成功")

	// マイグレーション実行
	if err := runMigrations(db); err != nil {
		log.Fatalf("マイグレーションエラー: %v", err)
	}

	// リポジトリ
	taskRepo := repository.NewTaskRepository(db)
	robotRepo := repository.NewRobotRepository(db)

	// ROS Bridge クライアント
	var rosClient rosbridge.Client
	if cfg.UseMockROS {
		log.Println("モックROS2ブリッジを使用")
		rosClient = rosbridge.NewMockClient()
	} else {
		log.Printf("実ROS2ブリッジに接続: %s", cfg.RosBridgeURL)
		wsClient := rosbridge.NewWSClient(cfg.RosBridgeURL)
		if err := wsClient.Connect(); err != nil {
			log.Printf("ROSBridge接続失敗（モックにフォールバック）: %v", err)
			rosClient = rosbridge.NewMockClient()
		} else {
			rosClient = wsClient
		}
	}

	// LLMサービス
	var llmService service.LLMService
	if cfg.UseMockLLM {
		log.Println("モックLLMサービスを使用")
		llmService = service.NewMockLLMService()
	} else {
		log.Println("OpenAI LLMへの接続は未実装（モックを使用）")
		llmService = service.NewMockLLMService()
	}

	// サービス
	taskService := service.NewTaskService(taskRepo, robotRepo, llmService, rosClient)
	robotService := service.NewRobotService(robotRepo)

	// WebSocketハブ（フロントエンドへのリアルタイム配信）
	wsHub := handler.NewWSHub()

	// ROS2コールバック登録（タスク完了時にDB更新 + フロントエンド通知）
	rosClient.SetOnTaskComplete(func(taskID string, success bool, message string) {
		ctx := context.Background()
		if err := taskService.CompleteTask(ctx, taskID, success, message); err != nil {
			log.Printf("タスク完了処理エラー: %v", err)
		}
		// フロントエンドに通知
		status := "completed"
		if !success {
			status = "failed"
		}
		wsHub.Broadcast("task_status", map[string]string{
			"task_id": taskID, "status": status, "message": message,
		})
	})
	rosClient.SetOnRobotState(func(robotID string, x, y float64, location string) {
		// UUID形式でない場合はマッピング未完了なのでスキップ
		if len(robotID) < 36 {
			return
		}
		ctx := context.Background()
		if err := robotService.UpdatePosition(ctx, robotID, x, y, location); err != nil {
			log.Printf("ロボット位置更新エラー: %v", err)
		}
		// フロントエンドに通知
		wsHub.Broadcast("robot_state", map[string]any{
			"robot_id": robotID, "position_x": x, "position_y": y, "current_location": location,
		})
	})

	// ハンドラ
	chatHandler := handler.NewChatHandler(taskService)
	taskHandler := handler.NewTaskHandler(taskService)
	robotHandler := handler.NewRobotHandler(robotService)

	// ルーティング
	mux := http.NewServeMux()

	// ヘルスチェック
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API v1
	mux.HandleFunc("/api/v1/chat", chatHandler.HandleChat)
	mux.HandleFunc("/api/v1/tasks", taskHandler.HandleList)
	mux.HandleFunc("/api/v1/tasks/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/cancel") {
			taskHandler.HandleCancel(w, r)
		} else {
			taskHandler.HandleGet(w, r)
		}
	})
	mux.HandleFunc("/api/v1/robots", robotHandler.HandleList)
	mux.HandleFunc("/api/v1/robots/", robotHandler.HandleGet)
	mux.HandleFunc("/api/v1/locations", robotHandler.HandleLocations)
	mux.HandleFunc("/api/v1/dashboard/stats", taskHandler.HandleStats)

	// WebSocket（フロントエンド向けリアルタイム配信）
	mux.HandleFunc("/ws", wsHub.HandleWS)

	// サーバー起動
	server := &http.Server{
		Addr:    cfg.ListenAddr(),
		Handler: handler.CORSMiddleware(mux),
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("シャットダウン中...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	log.Printf("RoboTasker API サーバー起動: http://localhost%s", cfg.ListenAddr())
	log.Println("エンドポイント:")
	log.Println("  POST /api/v1/chat          - 自然言語でタスク作成")
	log.Println("  GET  /api/v1/tasks         - タスク一覧")
	log.Println("  GET  /api/v1/tasks/:id     - タスク詳細")
	log.Println("  PATCH /api/v1/tasks/:id/cancel - タスクキャンセル")
	log.Println("  GET  /api/v1/robots        - ロボット一覧")
	log.Println("  GET  /api/v1/locations      - 地点一覧")
	log.Println("  GET  /api/v1/dashboard/stats - 統計")

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("サーバーエラー: %v", err)
	}
	log.Println("サーバー停止")
}

// runMigrations はSQLマイグレーションを実行する（簡易版）
func runMigrations(db *sql.DB) error {
	// マイグレーション管理テーブル
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		filename VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT NOW()
	)`)
	if err != nil {
		return fmt.Errorf("マイグレーション管理テーブル作成エラー: %w", err)
	}

	// マイグレーションファイルを順番に実行
	migrations := []struct {
		filename string
		upSQL    string
	}{
		{"001_create_robots.sql", `
			CREATE EXTENSION IF NOT EXISTS "pgcrypto";
			CREATE TABLE IF NOT EXISTS robots (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				name VARCHAR(100) NOT NULL,
				status VARCHAR(20) NOT NULL DEFAULT 'idle',
				current_location VARCHAR(100),
				battery_level INT NOT NULL DEFAULT 100,
				position_x FLOAT NOT NULL DEFAULT 0.0,
				position_y FLOAT NOT NULL DEFAULT 0.0,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			);`},
		{"002_create_locations.sql", `
			CREATE TABLE IF NOT EXISTS locations (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				name VARCHAR(100) NOT NULL UNIQUE,
				x FLOAT NOT NULL,
				y FLOAT NOT NULL,
				floor VARCHAR(10) NOT NULL DEFAULT '1F',
				location_type VARCHAR(50) NOT NULL DEFAULT 'room'
			);`},
		{"003_create_tasks.sql", `
			CREATE TABLE IF NOT EXISTS tasks (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				robot_id UUID REFERENCES robots(id),
				original_text TEXT NOT NULL,
				parsed_action VARCHAR(50) NOT NULL,
				target_location_id UUID REFERENCES locations(id),
				priority INT NOT NULL DEFAULT 0,
				status VARCHAR(20) NOT NULL DEFAULT 'pending',
				result_message TEXT,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				started_at TIMESTAMP WITH TIME ZONE,
				completed_at TIMESTAMP WITH TIME ZONE
			);
			CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
			CREATE INDEX IF NOT EXISTS idx_tasks_robot_id ON tasks(robot_id);`},
		{"004_create_task_logs.sql", `
			CREATE TABLE IF NOT EXISTS task_logs (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
				from_status VARCHAR(20),
				to_status VARCHAR(20) NOT NULL,
				message TEXT,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			);
			CREATE INDEX IF NOT EXISTS idx_task_logs_task_id ON task_logs(task_id);`},
	}

	for _, m := range migrations {
		var count int
		db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE filename = $1", m.filename).Scan(&count)
		if count > 0 {
			continue
		}
		log.Printf("マイグレーション実行: %s", m.filename)
		if _, err := db.Exec(m.upSQL); err != nil {
			return fmt.Errorf("マイグレーション %s 失敗: %w", m.filename, err)
		}
		db.Exec("INSERT INTO schema_migrations (filename) VALUES ($1)", m.filename)
	}

	return nil
}
