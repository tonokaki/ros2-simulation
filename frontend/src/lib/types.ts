// バックエンドAPIのデータモデル

export interface Robot {
  id: string;
  name: string;
  status: "idle" | "moving" | "executing" | "charging" | "error";
  current_location: string | null;
  battery_level: number;
  position_x: number;
  position_y: number;
  created_at: string;
  updated_at: string;
}

export interface Location {
  id: string;
  name: string;
  x: number;
  y: number;
  floor: string;
  location_type: string;
}

export interface Task {
  id: string;
  robot_id: string | null;
  original_text: string;
  parsed_action: "deliver" | "patrol" | "return" | "goto";
  target_location_id: string | null;
  priority: number;
  status: "pending" | "assigned" | "in_progress" | "completed" | "failed" | "cancelled";
  result_message: string | null;
  created_at: string;
  started_at: string | null;
  completed_at: string | null;
  robot_name: string;
  target_location_name: string;
}

export interface ParsedTask {
  action: string;
  target_location: string;
  item: string;
  priority: string;
  confidence: number;
}

export interface ChatResponse {
  reply: string;
  task: Task | null;
  parsed_task: ParsedTask | null;
}

export interface DashboardStats {
  total_tasks: number;
  completed_tasks: number;
  pending_tasks: number;
  failed_tasks: number;
  active_robots: number;
  avg_duration_seconds: number;
}

// WebSocketメッセージ
export interface WSMessage {
  type: "task_status" | "robot_state";
  data: Record<string, unknown>;
}
