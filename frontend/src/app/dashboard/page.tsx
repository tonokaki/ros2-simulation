"use client";

import { useState, useEffect, useCallback } from "react";
import { getRobots, getLocations, getDashboardStats } from "@/lib/api";
import type { Robot, Location, DashboardStats } from "@/lib/types";
import { RobotMap } from "@/components/RobotMap";
import { StatusBadge } from "@/components/StatusBadge";
import { wsClient } from "@/lib/websocket";

export default function DashboardPage() {
  const [robots, setRobots] = useState<Robot[]>([]);
  const [locations, setLocations] = useState<Location[]>([]);
  const [stats, setStats] = useState<DashboardStats | null>(null);

  const fetchAll = useCallback(async () => {
    try {
      const [r, l, s] = await Promise.all([getRobots(), getLocations(), getDashboardStats()]);
      setRobots(r || []);
      setLocations(l || []);
      setStats(s);
    } catch (err) {
      console.error("ダッシュボードデータ取得エラー:", err);
    }
  }, []);

  useEffect(() => {
    fetchAll();
    wsClient.connect();
    const unsub = wsClient.subscribe((msg) => {
      if (msg.type === "robot_state" || msg.type === "task_status") {
        fetchAll();
      }
    });
    const interval = setInterval(fetchAll, 3000);
    return () => {
      unsub();
      clearInterval(interval);
    };
  }, [fetchAll]);

  return (
    <div className="max-w-7xl mx-auto p-4">
      <div className="mb-4">
        <h1 className="text-xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-sm text-gray-500">ロボット状態とタスク統計</p>
      </div>

      {/* 統計カード */}
      {stats && (
        <div className="grid grid-cols-2 md:grid-cols-5 gap-3 mb-6">
          <StatCard label="全タスク" value={stats.total_tasks} />
          <StatCard label="完了" value={stats.completed_tasks} color="text-green-600" />
          <StatCard label="待機中" value={stats.pending_tasks} color="text-yellow-600" />
          <StatCard label="失敗" value={stats.failed_tasks} color="text-red-600" />
          <StatCard
            label="平均所要時間"
            value={stats.avg_duration_seconds > 0 ? `${stats.avg_duration_seconds.toFixed(1)}s` : "-"}
          />
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* 2Dマップ */}
        <div className="bg-white border border-gray-200 rounded-lg p-4">
          <h2 className="text-sm font-semibold text-gray-700 mb-3">Floor Map</h2>
          <RobotMap robots={robots} locations={locations} />
          <div className="flex gap-4 mt-3 text-xs text-gray-500">
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-gray-400" /> 待機
            </span>
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-blue-500" /> 移動中
            </span>
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-green-500" /> 充電中
            </span>
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-red-500" /> エラー
            </span>
          </div>
        </div>

        {/* ロボット一覧 */}
        <div className="bg-white border border-gray-200 rounded-lg p-4">
          <h2 className="text-sm font-semibold text-gray-700 mb-3">Robots</h2>
          <div className="space-y-3">
            {robots.map((robot) => (
              <div key={robot.id} className="border border-gray-100 rounded-lg p-3">
                <div className="flex items-center justify-between mb-2">
                  <span className="font-medium text-sm">{robot.name}</span>
                  <StatusBadge status={robot.status} />
                </div>
                <div className="grid grid-cols-2 gap-2 text-xs text-gray-500">
                  <div>
                    位置: ({robot.position_x.toFixed(1)}, {robot.position_y.toFixed(1)})
                  </div>
                  <div>現在地: {robot.current_location || "-"}</div>
                  <div>
                    バッテリー:
                    <span className={robot.battery_level <= 20 ? " text-red-500 font-medium" : ""}>
                      {" "}{robot.battery_level}%
                    </span>
                  </div>
                  <div>更新: {new Date(robot.updated_at).toLocaleTimeString("ja-JP")}</div>
                </div>
                {/* バッテリーバー */}
                <div className="mt-2 h-1.5 bg-gray-100 rounded-full overflow-hidden">
                  <div
                    className={`h-full rounded-full transition-all ${
                      robot.battery_level > 50 ? "bg-green-400" :
                      robot.battery_level > 20 ? "bg-yellow-400" : "bg-red-400"
                    }`}
                    style={{ width: `${robot.battery_level}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

function StatCard({
  label,
  value,
  color,
}: {
  label: string;
  value: number | string;
  color?: string;
}) {
  return (
    <div className="bg-white border border-gray-200 rounded-lg p-3">
      <p className="text-xs text-gray-500">{label}</p>
      <p className={`text-2xl font-bold ${color || "text-gray-900"}`}>{value}</p>
    </div>
  );
}
