"use client";

import { useState, useEffect, useCallback } from "react";
import { getTasks, cancelTask } from "@/lib/api";
import type { Task } from "@/lib/types";
import { StatusBadge } from "@/components/StatusBadge";
import { wsClient } from "@/lib/websocket";

const actionLabels: Record<string, string> = {
  deliver: "配達",
  patrol: "巡回",
  return: "帰還",
  goto: "移動",
};

export default function TasksPage() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchTasks = useCallback(async () => {
    try {
      const data = await getTasks();
      setTasks(data || []);
    } catch (err) {
      console.error("タスク取得エラー:", err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTasks();
    // WebSocket経由でリアルタイム更新
    wsClient.connect();
    const unsub = wsClient.subscribe((msg) => {
      if (msg.type === "task_status") {
        fetchTasks();
      }
    });
    // 5秒ごとにポーリング（フォールバック）
    const interval = setInterval(fetchTasks, 5000);
    return () => {
      unsub();
      clearInterval(interval);
    };
  }, [fetchTasks]);

  const handleCancel = async (id: string) => {
    try {
      await cancelTask(id);
      fetchTasks();
    } catch (err) {
      alert(`キャンセルエラー: ${err}`);
    }
  };

  const formatTime = (iso: string | null) => {
    if (!iso) return "-";
    return new Date(iso).toLocaleTimeString("ja-JP");
  };

  const formatDuration = (start: string | null, end: string | null) => {
    if (!start || !end) return "-";
    const ms = new Date(end).getTime() - new Date(start).getTime();
    return `${(ms / 1000).toFixed(1)}s`;
  };

  if (loading) {
    return (
      <div className="max-w-5xl mx-auto p-4">
        <p className="text-gray-500">読み込み中...</p>
      </div>
    );
  }

  return (
    <div className="max-w-5xl mx-auto p-4">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Tasks</h1>
          <p className="text-sm text-gray-500">タスク一覧 ({tasks.length}件)</p>
        </div>
        <button
          onClick={fetchTasks}
          className="text-sm text-gray-600 hover:text-gray-900 border border-gray-300 rounded px-3 py-1"
        >
          更新
        </button>
      </div>

      {tasks.length === 0 ? (
        <p className="text-gray-400 text-center mt-10">タスクがありません</p>
      ) : (
        <div className="space-y-2">
          {tasks.map((task) => (
            <div
              key={task.id}
              className="bg-white border border-gray-200 rounded-lg p-4 flex items-center gap-4"
            >
              {/* ステータス */}
              <div className="w-20">
                <StatusBadge status={task.status} />
              </div>

              {/* 内容 */}
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-gray-900 truncate">
                  {task.original_text}
                </p>
                <div className="flex gap-3 text-xs text-gray-500 mt-1">
                  <span>{actionLabels[task.parsed_action] || task.parsed_action}</span>
                  {task.target_location_name && <span>→ {task.target_location_name}</span>}
                  {task.robot_name && <span>by {task.robot_name}</span>}
                  {task.priority > 0 && (
                    <span className="text-red-500 font-medium">HIGH</span>
                  )}
                </div>
              </div>

              {/* 時刻 */}
              <div className="text-xs text-gray-400 text-right space-y-0.5 w-28">
                <div>作成: {formatTime(task.created_at)}</div>
                <div>所要: {formatDuration(task.started_at, task.completed_at)}</div>
              </div>

              {/* 結果メッセージ */}
              {task.result_message && (
                <div className="text-xs text-gray-500 w-40 truncate">{task.result_message}</div>
              )}

              {/* キャンセルボタン */}
              {(task.status === "pending" || task.status === "assigned" || task.status === "in_progress") && (
                <button
                  onClick={() => handleCancel(task.id)}
                  className="text-xs text-red-500 hover:text-red-700 border border-red-200 rounded px-2 py-1"
                >
                  Cancel
                </button>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
