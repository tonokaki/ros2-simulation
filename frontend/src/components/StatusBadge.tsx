"use client";

const statusColors: Record<string, string> = {
  idle: "bg-gray-100 text-gray-700",
  moving: "bg-blue-100 text-blue-700",
  executing: "bg-yellow-100 text-yellow-700",
  charging: "bg-green-100 text-green-700",
  error: "bg-red-100 text-red-700",
  pending: "bg-gray-100 text-gray-700",
  assigned: "bg-blue-100 text-blue-700",
  in_progress: "bg-yellow-100 text-yellow-700",
  completed: "bg-green-100 text-green-700",
  failed: "bg-red-100 text-red-700",
  cancelled: "bg-gray-200 text-gray-500",
};

const statusLabels: Record<string, string> = {
  idle: "待機中",
  moving: "移動中",
  executing: "実行中",
  charging: "充電中",
  error: "エラー",
  pending: "待機中",
  assigned: "割当済",
  in_progress: "実行中",
  completed: "完了",
  failed: "失敗",
  cancelled: "キャンセル",
};

export function StatusBadge({ status }: { status: string }) {
  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${statusColors[status] || "bg-gray-100 text-gray-700"}`}
    >
      {statusLabels[status] || status}
    </span>
  );
}
