"use client";

import type { Robot, Location } from "@/lib/types";

interface Props {
  robots: Robot[];
  locations: Location[];
}

// 座標をSVG描画用に変換 (実座標 → SVG座標)
const SCALE = 35;
const OFFSET_X = 30;
const OFFSET_Y = 30;

function toSvg(x: number, y: number): [number, number] {
  return [x * SCALE + OFFSET_X, y * SCALE + OFFSET_Y];
}

const robotStatusColor: Record<string, string> = {
  idle: "#6B7280",
  moving: "#3B82F6",
  executing: "#F59E0B",
  charging: "#10B981",
  error: "#EF4444",
};

export function RobotMap({ robots, locations }: Props) {
  const width = 420;
  const height = 350;

  return (
    <svg viewBox={`0 0 ${width} ${height}`} className="w-full bg-gray-50 rounded-lg border border-gray-200">
      {/* グリッド */}
      {Array.from({ length: 12 }).map((_, i) => (
        <line key={`gx-${i}`} x1={i * SCALE + OFFSET_X} y1={OFFSET_Y - 10} x2={i * SCALE + OFFSET_X} y2={height - 10} stroke="#E5E7EB" strokeWidth={0.5} />
      ))}
      {Array.from({ length: 10 }).map((_, i) => (
        <line key={`gy-${i}`} x1={OFFSET_X - 10} y1={i * SCALE + OFFSET_Y} x2={width - 10} y2={i * SCALE + OFFSET_Y} stroke="#E5E7EB" strokeWidth={0.5} />
      ))}

      {/* 地点 */}
      {locations.map((loc) => {
        const [sx, sy] = toSvg(loc.x, loc.y);
        const isStation = loc.location_type === "station";
        return (
          <g key={loc.id}>
            <rect
              x={sx - 14}
              y={sy - 14}
              width={28}
              height={28}
              rx={isStation ? 14 : 4}
              fill={isStation ? "#DBEAFE" : "#F3F4F6"}
              stroke={isStation ? "#3B82F6" : "#9CA3AF"}
              strokeWidth={1.5}
            />
            <text x={sx} y={sy + 24} textAnchor="middle" fontSize={9} fill="#374151" fontWeight={500}>
              {loc.name}
            </text>
          </g>
        );
      })}

      {/* ロボット */}
      {robots.map((robot) => {
        const [sx, sy] = toSvg(robot.position_x, robot.position_y);
        const color = robotStatusColor[robot.status] || "#6B7280";
        return (
          <g key={robot.id}>
            {/* 移動中は脈動アニメーション */}
            {robot.status === "moving" && (
              <circle cx={sx} cy={sy} r={12} fill={color} opacity={0.2}>
                <animate attributeName="r" values="12;18;12" dur="1.5s" repeatCount="indefinite" />
                <animate attributeName="opacity" values="0.2;0.05;0.2" dur="1.5s" repeatCount="indefinite" />
              </circle>
            )}
            <circle cx={sx} cy={sy} r={8} fill={color} stroke="white" strokeWidth={2} />
            <text x={sx} y={sy - 14} textAnchor="middle" fontSize={8} fill={color} fontWeight={600}>
              {robot.name.replace("RoboTasker-", "R")}
            </text>
            {/* バッテリー */}
            <rect x={sx - 8} y={sy + 12} width={16} height={4} rx={1} fill="#E5E7EB" />
            <rect x={sx - 8} y={sy + 12} width={Math.max(0, robot.battery_level * 0.16)} height={4} rx={1}
              fill={robot.battery_level > 20 ? "#10B981" : "#EF4444"} />
          </g>
        );
      })}
    </svg>
  );
}
