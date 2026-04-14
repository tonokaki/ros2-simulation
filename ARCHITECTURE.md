# アーキテクチャ設計

## 設計思想

### なぜこの構成にしたか

1. **レイヤー分離**: バックエンド・ROS2・フロントエンドを疎結合にし、各層を独立して開発・テスト・差し替え可能にした
2. **モック優先**: ROS2やLLMが未実装でもバックエンド単体で動作検証できるよう、インターフェースベースで設計
3. **rosbridge経由の接続**: GoバックエンドとROS2をWebSocketで接続することで、言語の壁を越えつつ標準的な手法を採用
4. **イベント駆動**: タスクの状態遷移をログとして記録し、監査性と可視化を両立

### コンポーネント間の通信

```
Frontend ←→ Backend API: REST + WebSocket
Backend API ←→ ROS2: rosbridge (WebSocket)
Backend API ←→ LLM: HTTP (OpenAI API)
Backend API ←→ DB: SQL (lib/pq)
```

## バックエンド構成

```
internal/
├── config/      # 環境変数ベースの設定
├── model/       # データモデル（DB + API）
├── repository/  # DB操作（SQLクエリ）
├── service/     # ビジネスロジック
├── handler/     # HTTPハンドラ
└── rosbridge/   # ROS2ブリッジクライアント
```

**責務分離のルール**:
- handler: HTTPリクエストのパース・レスポンスの組み立てのみ
- service: ビジネスロジック（タスク割り当て、状態遷移、LLM呼び出し）
- repository: SQLクエリの発行のみ

## タスク状態遷移

```
pending → assigned → in_progress → completed
                                 → failed
         ↓ (どの状態からも)
       cancelled
```

各遷移は `task_logs` テーブルに記録される。

## 将来の拡張ポイント

### LLM実装の差し替え
`service.LLMService` インターフェースを実装するだけで、OpenAI API / Claude / ローカルLLM に切り替え可能。

### ROS2実接続
`rosbridge.Client` インターフェースを実装し、実際のrosbridge WebSocketサーバーに接続する。

### 複数ロボット対応
`FindIdleRobot` のロジックを拡張し、距離・バッテリー残量・スキルセットでロボットを選択するスケジューラーに進化させる。
