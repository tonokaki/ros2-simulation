.PHONY: up down build seed logs backend-logs ros2-logs db-shell api-test up-ros2 ros2-topics

# === 基本操作 ===

# バックエンド + DB 起動（モックROS2モード）
up:
	docker compose up -d --build

# ROS2含む全体起動（実ROS2接続モード）
up-ros2:
	docker compose --profile ros2 up -d --build

# 全体停止
down:
	docker compose --profile ros2 down

# 再ビルド
build:
	docker compose --profile ros2 build

# 初期データ投入
seed:
	docker compose exec -T db psql -U robotasker robotasker < scripts/seed.sql

# === ログ ===

logs:
	docker compose --profile ros2 logs -f

backend-logs:
	docker compose logs -f backend

ros2-logs:
	docker compose logs -f ros2

# === デバッグ ===

db-shell:
	docker compose exec db psql -U robotasker robotasker

# ROS2トピック一覧
ros2-topics:
	docker compose exec ros2 bash -c "source /ros2_ws/install/setup.bash && ros2 topic list"

# ROS2ロボット状態をリアルタイム表示
ros2-echo-state:
	docker compose exec ros2 bash -c "source /ros2_ws/install/setup.bash && ros2 topic echo /robot/state"

# === テスト ===

# API テスト（curl）
api-test:
	@echo "=== ヘルスチェック ==="
	curl -s http://localhost:8080/health | python3 -m json.tool
	@echo "\n=== 地点一覧 ==="
	curl -s http://localhost:8080/api/v1/locations | python3 -m json.tool
	@echo "\n=== ロボット一覧 ==="
	curl -s http://localhost:8080/api/v1/robots | python3 -m json.tool
	@echo "\n=== チャット（タスク作成） ==="
	curl -s -X POST http://localhost:8080/api/v1/chat \
		-H 'Content-Type: application/json' \
		-d '{"message":"会議室Bに資料を届けて"}' | python3 -m json.tool
	@echo "\n=== タスク一覧 ==="
	curl -s http://localhost:8080/api/v1/tasks | python3 -m json.tool
