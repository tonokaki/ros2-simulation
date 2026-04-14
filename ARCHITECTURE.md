# アーキテクチャ設計

## 設計思想

1. **レイヤー分離**: バックエンド・ROS2・フロントエンドを疎結合にし、各層を独立して開発・テ��ト・差し替え可能にした
2. **モック優先**: ROS2やLLMが未実装でもバックエンド単体で動作検証できるよう、インターフェースベースで設計
3. **rosbridge経由の接続**: GoバックエンドとROS2をWebSocketで接続することで、言語の壁を越えつつ標準的な手法を採用
4. **イベント駆動**: タスクの状態遷移をログとして記録し、監査性と可視化を両立

## コンポーネント間の通信

```
Frontend ←── REST + WebSocket ──→ Backend API
                                      │
                        ┌──────────────┼──────────────┐
                        │              │              │
                   PostgreSQL    rosbridge WS     OpenAI API
                   (lib/pq)     (gorilla/ws)     (HTTP, モック可)
                                      │
                              ROS2 bridge_node
                                      │
                        ┌─────────────┼──────────────┐
                        │                            │
                task_executor_node        navigation_controller
                        │                            │
                  /task/command              /navigation/goal
                  /task/status              /robot/state
                                            /cmd_vel
```

## バックエンド構成

```
internal/
├── config/      # 環境変数ベースの設定 (DATABASE_URL, USE_MOCK_*, etc.)
├── model/       # データモデル (Task, Robot, Location, ChatRequest/Response)
├── repository/  # DB操作 (トランザクション付きステータス更新、統計集計)
├── service/     # ビジネスロジック
│   ├── task.go      # タスク作成 → ロボット割当 → ROS2送信 → 完了処理
│   ├── llm.go       # LLMService インターフェース + MockLLMService
│   ├── rosbridge.go # ROSBridgeService インターフェース
│   └─�� robot.go     # ロボット状態管理
├── handler/     # HTTP ハンドラ + WebSocket ハブ
│   ├── chat.go      # POST /api/v1/chat
│   ├── task.go      # GET/PATCH /api/v1/tasks
│   ├── robot.go     # GET /api/v1/robots, /api/v1/locations
│   ├── websocket.go # WS /ws (フロントエンド向けリアルタイム配信)
��   └── util.go      # CORS, JSON レスポンス, パスパース
└── rosbridge/   # ROS2 ブリッジクライアント
    ├── client.go    # Client インターフェース + MockClient
    └── ws_client.go # 実 WebSocket 接続 (gorilla/websocket, リトライ/再接続付き)
```

**責務分離のルール**:
- handler: HTTPリクエストのパース・レスポンスの組み立てのみ
- service: ビジネスロジック（タスク割り当て、状態遷移、LLM呼び出し）
- repository: SQLクエリの発行のみ（トランザクション管理含む）
- rosbridge: ROS2 通信のみ（ビジネスロジックは持たない）

## ROS2 ノード構成

```
bridge_node (robotasker_bridge)
├── WebSocket サーバー (ポート 9090)
├── Pub: /task/command  ← バックエンドからの指令を変換
├── Sub: /task/status   → バックエンドにステータスを転送
├── Sub: /robot/state   → バックエンドにロボット状態を転送
└── ID マッピング: ローカルID ↔ バックエンド UUID

task_executor_node (task_executor)
├── Sub: /task/command  ← タスク受信
├── Pub: /task/status   → タスク状態報告
├── Pub: /navigation/goal → ナビゲーション目標
├── Sub: /robot/state   ← 目的地到着検知
└── タイムアウト管理 (60秒)

navigation_controller
├── Sub: /navigation/goal ← 移動目標
├── Pub: /robot/state    → 位置・速度・バッテリー (10Hz)
├── Pub: /cmd_vel        → 速度指令 (Gazebo 連携用)
└── 簡易直線移動シミュレーション (1.0 m/s)
```

## DB スキーマ

```
robots          # ロボット (id, name, status, position, battery)
locations       # 移動可能地点 (id, name, x, y, floor, type)
tasks           # タスク (id, robot_id, original_text, parsed_action, status, ...)
task_logs       # 状態遷移履歴 (id, task_id, from_status, to_status, message)
schema_migrations  # マイグレーション管理
```

## タスク状態遷移

```
pending → assigned → in_progress → completed
                                 → failed
         ↓ (どの状態からも)
       cancelled
```

各遷移は `task_logs` テーブルに記録され、from_status / to_status / timestamp が保存される。

## イベントフロー（E2E）

```
1. ユーザー → Frontend: 「会議室Bに資料を届けて」
2. Frontend → Backend: POST /api/v1/chat {message: "..."}
3. Backend → LLM: 自然言語パース → {action: deliver, target: 会議室B}
4. Backend → DB: INSERT tasks (status: pending)
5. Backend → DB: FindIdleRobot → RoboTasker-01
6. Backend → DB: UPDATE tasks (robot_id, status: assigned)
7. Backend → rosbridge: WebSocket {type: task_command, ...}
8. bridge_node → /task/command: UUID→ローカルID変換、ROS2 メッセージ発行
9. task_executor → /navigation/goal: PoseStamped(7.0, 4.0)
10. navigation_controller: 直線移動シミュレーション開始
11. navigation_controller → /robot/state: 位置を10Hz で発行
12. bridge_node → Backend: WebSocket {type: robot_state, ...}
13. Backend → DB: UPDATE robots (position)
14. Backend → Frontend: WebSocket {type: robot_state, ...}
15. navigation_controller: 到着判定 (距離 < 0.3m)
16. navigation_controller → /robot/state: {current_location: "会議室B"}
17. task_executor: 目的地到着検知 → /task/status {status: completed}
18. bridge_node → Backend: WebSocket {type: task_status, status: completed}
19. Backend → DB: UPDATE tasks (status: completed, completed_at)
20. Backend → Frontend: WebSocket {type: task_status, ...}
```

## インターフェース境界（差し替えポイント）

| インターフェース | モック実装 | 本番実装 | 切り替え方法 |
|---|---|---|---|
| `LLMService` | MockLLMService (正規表現) | OpenAI Function Calling | `USE_MOCK_LLM=false` + `OPENAI_API_KEY` |
| `rosbridge.Client` | MockClient (goroutine) | WSClient (WebSocket) | `USE_MOCK_ROS=false` + `ROSBRIDGE_URL` |
| navigation_controller | 簡易直線移動 | Nav2 | launch ファイルで差し替え |
| Gazebo | なし (簡易シミュレーター) | office.world + TurtleBot3 | `simulation.launch.py` で起動 |

## Gazebo シミュレーション

`ros2_ws/src/robotasker_sim/worlds/office.world`:
- 12m x 10m のオフィスフロア
- 外壁 + 会議室A/B の仕切り壁
- 地点マーカー (色付き円盤)
- TurtleBot3 (burger) をスポーン

```bash
# Gazebo 起動
export TURTLEBOT3_MODEL=burger
ros2 launch robotasker_sim simulation.launch.py

# ヘッドレス
ros2 launch robotasker_sim simulation.launch.py gui:=false
```

## 将来の拡張ポイント

### LLM 実装の差し替え
`service.LLMService` インターフェースを実装するだけで、OpenAI API / Claude / Ollama (ローカル) に切り替え可能。

### Nav2 導入
`navigation_controller` の `/navigation/goal` インターフェースは Nav2 と互換。launch ファイルで差し替えれば、SLAM ベースのナビゲーションに移行できる。

### 複数ロボット対応
`FindIdleRobot` のロジックを拡張し、距離・バッテリー残量・スキルセットでロボットを選択するスケジューラーに進化させる。bridge_node の ID マッピングは複数ロボット対応済み。

### 実機接続
`/cmd_vel` トピックはそのまま実機のモータードライバに接続可能。TurtleBot3 の実機なら、ソフトウェア変更なしで動作する。
