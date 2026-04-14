#!/bin/bash
# RoboTasker 初回セットアップスクリプト
set -e

echo "============================================"
echo "  RoboTasker - セットアップ"
echo "============================================"

# Docker チェック
if ! command -v docker &> /dev/null; then
    echo "[ERROR] Docker がインストールされていません"
    echo "  → https://docs.docker.com/get-docker/"
    exit 1
fi

if ! docker compose version &> /dev/null; then
    echo "[ERROR] Docker Compose が利用できません"
    exit 1
fi

echo "[OK] Docker & Docker Compose 検出"

# 1. バックエンド + DB 起動
echo ""
echo "--- Step 1: バックエンド + DB を起動 ---"
docker compose up -d --build db backend
echo "[OK] バックエンド起動完了"

# 2. DB が起動するまで待つ
echo ""
echo "--- Step 2: DB 接続待機 ---"
for i in $(seq 1 30); do
    if docker compose exec -T db pg_isready -U robotasker > /dev/null 2>&1; then
        echo "[OK] DB 接続成功"
        break
    fi
    echo "  待機中... ($i/30)"
    sleep 1
done

# 3. シードデータ投入
echo ""
echo "--- Step 3: 初期データ投入 ---"
docker compose exec -T db psql -U robotasker robotasker < scripts/seed.sql
echo "[OK] シードデータ投入完了"

# 4. ヘルスチェック
echo ""
echo "--- Step 4: ヘルスチェック ---"
for i in $(seq 1 10); do
    if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
        echo "[OK] API ヘルスチェック成功"
        break
    fi
    sleep 1
done

# 5. フロントエンド
echo ""
echo "--- Step 5: フロントエンド ---"
if [ -d "frontend/node_modules" ]; then
    echo "[OK] node_modules 検出済み"
else
    echo "  npm install 実行中..."
    cd frontend && npm install && cd ..
    echo "[OK] npm install 完了"
fi

echo ""
echo "============================================"
echo "  セットアップ完了"
echo "============================================"
echo ""
echo "  バックエンド API: http://localhost:8080"
echo "  フロントエンド:   cd frontend && npm run dev"
echo "                    → http://localhost:3000"
echo ""
echo "  ROS2接続モード:   make up-ros2"
echo ""
echo "  使い方:"
echo "    1. ブラウザで http://localhost:3000 を開く"
echo "    2. Chat 画面で「会議室Bに資料を届けて」と入力"
echo "    3. Tasks / Dashboard で状態を確認"
echo ""
