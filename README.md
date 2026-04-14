# RoboTasker - 自然言語タスク管理 + 自律移動ロボットシステム

自然言語でロボットに指示を出し、タスクの生成・実行・記録・可視化を行う統合システムです。

## コンセプト

「会議室Bに資料を届けて」のような自然言語指示を、LLMで構造化タスクに変換し、ROS2ロボットが自律的に実行します。すべてのタスクはバックエンドで管理・記録され、ダッシュボードでリアルタイムに可視化されます。

## アーキテクチャ

```
[Chat UI] → [Go API] → [LLM Service] → タスク構造化
                ↓
           [PostgreSQL] ← タスク・状態記録
                ↓
         [ROS2 Bridge] → [ROS2 Nodes] → [Gazebo Simulation]
                ↑
           [WebSocket] → [Dashboard UI] リアルタイム更新
```

詳細は [ARCHITECTURE.md](ARCHITECTURE.md) を参照してください。

## 技術スタック

| レイヤー | 技術 |
|---|---|
| Backend API | Go (net/http) |
| Database | PostgreSQL 16 |
| LLM | OpenAI API (モック実装あり) |
| ROS2 | ROS2 Humble + rclpy |
| Simulation | Gazebo (Phase 4) / モックシミュレーター |
| Frontend | Next.js + TypeScript (Phase 3) |
| Infrastructure | Docker Compose |

## クイックスタート

### 必要なもの

- Docker & Docker Compose

### 起動

```bash
# 全体起動（Go API + PostgreSQL）
make up

# 初期データ投入
make seed

# 動作確認
make api-test
```

### API テスト例

```bash
# 自然言語でタスク作成
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"会議室Bに資料を届けて"}'

# タスク一覧
curl http://localhost:8080/api/v1/tasks

# ロボット一覧
curl http://localhost:8080/api/v1/robots

# 地点一覧
curl http://localhost:8080/api/v1/locations

# ダッシュボード統計
curl http://localhost:8080/api/v1/dashboard/stats
```

### ROS2接続モードで起動

```bash
# ROS2ノード含む全体起動
make up-ros2

# 初期データ投入
make seed

# ROS2トピック確認
make ros2-topics

# ロボット状態をリアルタイム表示
make ros2-echo-state
```

### 停止

```bash
make down
```

## プロジェクト構成

```
ros2-simulation/
├── backend/              # Go バックエンドAPI
│   ├── cmd/server/       # エントリポイント
│   ├── internal/         # ビジネスロジック
│   │   ├── handler/      # HTTPハンドラ
│   │   ├── service/      # ビジネスロジック
│   │   ├── repository/   # DB操作
│   │   ├── model/        # データモデル
│   │   └── rosbridge/    # ROS2接続（モック/実装）
│   └── migrations/       # DBマイグレーション
├── ros2_ws/src/           # ROS2 ワークスペース
│   ├── robotasker_msgs/  # カスタムメッセージ定義
│   └── robotasker_core/  # コアノード群
│       ├── bridge_node         # バックエンド↔ROS2ブリッジ
│       ├── task_executor_node  # タスク実行管理
│       └── navigation_controller # 移動シミュレーション
├── frontend/             # Next.js フロントエンド (Phase 3)
├── scripts/              # ユーティリティスクリプト
├── infra/docker/         # Dockerfiles
├── docs/                 # ドキュメント
└── docker-compose.yml
```

## ROS2 トピック構成

```
/task/command     (robotasker_msgs/TaskCommand)  バックエンド→ロボット
/task/status      (robotasker_msgs/TaskStatus)   ロボット→バックエンド
/robot/state      (robotasker_msgs/RobotState)   ロボット状態（10Hz）
/navigation/goal  (geometry_msgs/PoseStamped)    ナビゲーション目標
/cmd_vel          (geometry_msgs/Twist)          速度指令
```

## 開発ロードマップ

- [x] **Phase 1**: バックエンドAPI + モックROS2/LLM
- [x] **Phase 2**: ROS2ノード + 簡易シミュレーション
- [ ] **Phase 3**: フロントエンド + WebSocket + デモ品質
- [ ] **Phase 4**: Gazebo + ドキュメント + ポートフォリオ仕上げ

## 設計判断

このプロジェクトの設計判断については [ARCHITECTURE.md](ARCHITECTURE.md) を参照してください。

## ライセンス

MIT
