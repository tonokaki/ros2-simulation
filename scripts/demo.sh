#!/bin/bash
# RoboTasker デモ実行スクリプト
# 一連のタスクを自動送信し、システムの動作を確認する
set -e

API="http://localhost:8080"

echo "============================================"
echo "  RoboTasker - デモ実行"
echo "============================================"
echo ""

# ヘルスチェック
echo "--- ヘルスチェック ---"
curl -sf "$API/health" | python3 -m json.tool
echo ""

# 現在のロボット状態
echo "--- ロボット状態（初期） ---"
curl -s "$API/api/v1/robots" | python3 -m json.tool
echo ""

sleep 1

# タスク1: 配達
echo "--- タスク1: 配達 ---"
echo '> 「会議室Bに資料を届けて」'
curl -s -X POST "$API/api/v1/chat" \
  -H 'Content-Type: application/json' \
  -d '{"message":"会議室Bに資料を届けて"}' | python3 -m json.tool
echo ""

sleep 2

# タスク2: 緊急配達
echo "--- タスク2: 緊急配達 ---"
echo '> 「急ぎで倉庫に荷物を運んで」'
curl -s -X POST "$API/api/v1/chat" \
  -H 'Content-Type: application/json' \
  -d '{"message":"急ぎで倉庫に荷物を運んで"}' | python3 -m json.tool
echo ""

sleep 2

# タスク一覧
echo "--- タスク一覧（実行中） ---"
curl -s "$API/api/v1/tasks" | python3 -m json.tool
echo ""

# ロボット状態
echo "--- ロボット状態（移動中） ---"
curl -s "$API/api/v1/robots" | python3 -m json.tool
echo ""

echo "移動完了を待機中 (8秒)..."
sleep 8

# 完了後の状態
echo "--- タスク一覧（完了後） ---"
curl -s "$API/api/v1/tasks" | python3 -m json.tool
echo ""

# タスク3: 帰還
echo "--- タスク3: 帰還 ---"
echo '> 「充電ステーションに戻って」'
curl -s -X POST "$API/api/v1/chat" \
  -H 'Content-Type: application/json' \
  -d '{"message":"充電ステーションに戻って"}' | python3 -m json.tool
echo ""

sleep 8

# ダッシュボード統計
echo "--- ダッシュボード統計 ---"
curl -s "$API/api/v1/dashboard/stats" | python3 -m json.tool
echo ""

# 最終ロボット状態
echo "--- ロボット状態（最終） ---"
curl -s "$API/api/v1/robots" | python3 -m json.tool
echo ""

echo "============================================"
echo "  デモ完了"
echo "============================================"
