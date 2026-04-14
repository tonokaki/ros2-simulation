// バックエンドAPI クライアント

import type { Robot, Location, Task, ChatResponse, DashboardStats } from "./types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

async function fetchJSON<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: { "Content-Type": "application/json", ...options?.headers },
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`API error ${res.status}: ${body}`);
  }
  return res.json();
}

export async function sendChat(message: string): Promise<ChatResponse> {
  return fetchJSON("/api/v1/chat", {
    method: "POST",
    body: JSON.stringify({ message }),
  });
}

export async function getTasks(limit = 50): Promise<Task[]> {
  return fetchJSON(`/api/v1/tasks?limit=${limit}`);
}

export async function getTask(id: string): Promise<Task> {
  return fetchJSON(`/api/v1/tasks/${id}`);
}

export async function cancelTask(id: string): Promise<void> {
  await fetchJSON(`/api/v1/tasks/${id}/cancel`, { method: "PATCH" });
}

export async function getRobots(): Promise<Robot[]> {
  return fetchJSON("/api/v1/robots");
}

export async function getLocations(): Promise<Location[]> {
  return fetchJSON("/api/v1/locations");
}

export async function getDashboardStats(): Promise<DashboardStats> {
  return fetchJSON("/api/v1/dashboard/stats");
}
