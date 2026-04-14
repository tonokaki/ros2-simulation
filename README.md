# RoboTasker - 自然言語タスク管理 + 自律移動ロボットシステム

自然言語でロボットに指示を出し、タスクの生成・実行・記録・可視化を行う統合システムです。

> **ポートフォリオの狙い**: 「バックエンド × AI × ロボット」をつなぐ設計力を1つのプロジェクトで示す

## コンセプト

「会議室Bに資料を届けて」のような自然言語指示を、LLMで構造化タスクに変換し、ROS2ロボットが自律的に実行します。すべてのタスクはバックエンドで管理・記録され、ダッシュボードでリアルタイムに可視化されます。

### 特徴

- **自然言語 → 構造化タスク変換**: LLM (モック/OpenAI) でアクション・目的地・優先度を自動判定
- **タスクキュー + 自動ロボット割当**: バッテリー残量を考慮したスケジューリング
- **ROS2 ネイティブ通信**: カスタムメッセージ定義による型安全な通信
- **リアルタイム可視化**: WebSocket 経由でフロントエンドにロボット状態を配信
- **インターフェース設計**: モック↔実装を環境変数で切り替え可能

## アーキテクチャ

```
┌──────────────────────────────────────────────────────────────┐
│ Frontend (Next.js)                                           │
│  Chat UI │ Task Board │ Dashboard (2D Map)                   │
└────────┬──────────────────────────────────┬──────────────────┘
         │ REST API                         │ WebSocket
┌────────┴──────────────────────────────────┴──────────────────┐
│ Backend API (Go)                                             │
│  LLM Service │ Task Manager │ Robot Manager │ WS Hub         │
│       │              │                                       │
│  [PostgreSQL]   [rosbridge Client]                           │
└───────┼──────────────┼───────────────────────────────────────┘
        │ HTTP         │ WebSocket
   [OpenAI API]   ┌────┴──────────────────────────────────────┐
   (or Mock)      │ ROS2 Layer                                │
                  │  bridge_node ←→ task_executor_node         │
                  │                    ↓                       │
                  │  navigation_controller → [Gazebo / Simple] │
                  └────────────────────────────────────────────┘
```

詳細設計: [ARCHITECTURE.md](ARCHITECTURE.md) | 設計判断: [docs/design-decisions.md](docs/design-decisions.md)

## 技術スタック

| レイヤー | 技術 | 選定理由 |
|---|---|---|
| Backend API | Go (net/http + gorilla/websocket) | 並行処理・軽量バイナリ・型安全性 |
| Database | PostgreSQL 16 | 信頼性・UUID対応・JSON対応 |
| LLM | OpenAI API (モック実装あり) | Function Calling で構造化出力 |
| ROS2 | ROS2 Humble + rclpy | 業界標準・豊富なエコシステム |
| Simulation | Gazebo + 簡易シミュレーター | 段階的に精度を上げられる構成 |
| Frontend | Next.js 16 + TypeScript + Tailwind CSS | React経験活用・素早いUI構築 |
| Infrastructure | Docker Compose | 一発起動・環境差異なし |

## クイックスタート

### 必要なもの

- Docker & Docker Compose
- Node.js 20+ (フロントエンドのローカル開発時)

### 起動 (最短手順)

```bash
# 1. セットアップ (バックエンド + DB 起動 + シードデータ)
./scripts/setup.sh

# 2. フロントエンド起動
cd frontend && npm run dev

# 3. ブラウザで開く
open http://localhost:3000
```

### Docker で全体起動

```bash
make up      # バックエンド + DB + フロントエンド
make seed    # 初期データ投入
```

### ROS2 接続モード

```bash
make up-ros2    # ROS2ノード含む全体起動
make seed
make ros2-topics      # トピック一覧確認
make ros2-echo-state  # ロボット状態リアルタイム表示
```

### デモ実行

```bash
./scripts/demo.sh   # 一連のタスクを自動送信して動作確認
```

### API テスト

```bash
# 自然言語でタスク作成
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"会議室Bに資料を届けて"}'

# タスク一覧 / ロボット一覧 / 統計
curl http://localhost:8080/api/v1/tasks
curl http://localhost:8080/api/v1/robots
curl http://localhost:8080/api/v1/dashboard/stats
```

## プロジェクト構成

```
ros2-simulation/
├── backend/                   # Go バックエンドAPI
│   ├── cmd/server/            # エントリポイント
│   ├── internal/
│   │   ├── handler/           # HTTP ハンドラ + WebSocket ハブ
│   │   ├── service/           # ビジネスロジック (タスク管理, LLM, ロボット管理)
│   │   ├── repository/        # DB 操作 (PostgreSQL)
│   │   ├── model/             # データモデル
│   │   └── rosbridge/         # ROS2 接続 (MockClient / WSClient)
│   └── migrations/            # SQL マイグレーション
├── ros2_ws/src/
│   ├── robotasker_msgs/       # カスタムメッセージ定義
│   │   └── msg/               # TaskCommand, TaskStatus, RobotState
│   ├── robotasker_core/       # コアノード群
│   │   ├── bridge_node.py           # バックエンド ↔ ROS2 ブリッジ
│   │   ├── task_executor_node.py    # タスク実行管理
│   │   └── navigation_controller.py # 移動シミュレーション
│   └── robotasker_sim/        # Gazebo シミュレーション環境
│       └── worlds/office.world
├── frontend/                  # Next.js フロントエンド
│   └── src/
│       ├── app/               # Chat / Tasks / Dashboard ページ
│       ├── components/        # RobotMap, StatusBadge, Navigation
│       └── lib/               # API クライアント, WebSocket, 型定義
├── scripts/
│   ├── setup.sh               # 初回セットアップ
│   ├── demo.sh                # デモ実行
│   └── seed.sql               # 初期データ
├── docs/
│   ├── demo-script.md         # デモ動画台本
│   └── design-decisions.md    # 設計判断の記録
├── infra/docker/              # Dockerfiles (backend, ros2, frontend)
├── docker-compose.yml
├── Makefile
└── ARCHITECTURE.md
```

## ROS2 トピック構成

```
/task/command     (robotasker_msgs/TaskCommand)  バックエンド → ロボット
/task/status      (robotasker_msgs/TaskStatus)   ロボット → バックエンド
/robot/state      (robotasker_msgs/RobotState)   ロボット状態 (10Hz)
/navigation/goal  (geometry_msgs/PoseStamped)    ナビゲーション目標
/cmd_vel          (geometry_msgs/Twist)          速度指令 (Gazebo 連携用)
```

## API エンドポイント

| Method | Path | 説明 |
|---|---|---|
| POST | `/api/v1/chat` | 自然言語でタスク作成 |
| GET | `/api/v1/tasks` | タスク一覧 |
| GET | `/api/v1/tasks/:id` | タスク詳細 |
| PATCH | `/api/v1/tasks/:id/cancel` | タスクキャンセル |
| GET | `/api/v1/robots` | ロボット一覧・状態 |
| GET | `/api/v1/locations` | 移動可能地点一覧 |
| GET | `/api/v1/dashboard/stats` | ダッシュボード統計 |
| WS | `/ws` | リアルタイム更新 |

## 開発ロードマップ

- [x] **Phase 1**: バックエンド API + モック ROS2/LLM
- [x] **Phase 2**: ROS2 ノード + 簡易シミュレーション
- [x] **Phase 3**: フロントエンド + WebSocket + デモ品質
- [x] **Phase 4**: Gazebo ワールド + ドキュメント + ポートフォリオ仕上げ

## 将来の拡張

| 拡張 | 内容 | 難易度 |
|---|---|---|
| OpenAI API 実接続 | `LLMService` インターフェースの実装追加 | 低 |
| Nav2 導入 | navigation_controller を Nav2 に差し替え | 中 |
| 複数ロボット協調 | スケジューラーの高度化、衝突回避 | 中 |
| 実機接続 (TurtleBot3) | `cmd_vel` トピック経由で接続 | 中 |
| カメラ + 異常検知 | 点検機能の追加、Vision API 連携 | 高 |
| 音声入力 | Whisper API → チャット入力 | 低 |

## 設計判断

主要な設計判断とその理由は以下に記録しています:
- [ARCHITECTURE.md](ARCHITECTURE.md) — アーキテクチャ全体設計
- [docs/design-decisions.md](docs/design-decisions.md) — 個別の設計判断と根拠

## ライセンス

MIT
