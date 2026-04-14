"use client";

import { useState, useRef, useEffect } from "react";
import { sendChat } from "@/lib/api";
import type { ChatResponse } from "@/lib/types";
import { StatusBadge } from "@/components/StatusBadge";

interface Message {
  role: "user" | "assistant";
  text: string;
  chatResponse?: ChatResponse;
  timestamp: Date;
}

export default function ChatPage() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const handleSend = async () => {
    const text = input.trim();
    if (!text || loading) return;

    setInput("");
    setMessages((prev) => [...prev, { role: "user", text, timestamp: new Date() }]);
    setLoading(true);

    try {
      const res = await sendChat(text);
      setMessages((prev) => [
        ...prev,
        { role: "assistant", text: res.reply, chatResponse: res, timestamp: new Date() },
      ]);
    } catch (err) {
      setMessages((prev) => [
        ...prev,
        { role: "assistant", text: `エラー: ${err}`, timestamp: new Date() },
      ]);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-3xl mx-auto p-4 flex flex-col h-[calc(100vh-56px)]">
      <div className="mb-4">
        <h1 className="text-xl font-bold text-gray-900">Chat</h1>
        <p className="text-sm text-gray-500">自然言語でロボットに指示を出せます</p>
      </div>

      {/* メッセージ一覧 */}
      <div className="flex-1 overflow-y-auto space-y-3 mb-4">
        {messages.length === 0 && (
          <div className="text-center text-gray-400 mt-20">
            <p className="text-lg mb-2">ロボットに指示を出してみましょう</p>
            <div className="space-y-1 text-sm">
              <p>例: 「会議室Bに資料を届けて」</p>
              <p>例: 「急ぎで倉庫に荷物を運んで」</p>
              <p>例: 「充電ステーションに戻って」</p>
              <p>例: 「休憩室まで行って」</p>
            </div>
          </div>
        )}

        {messages.map((msg, i) => (
          <div key={i} className={`flex ${msg.role === "user" ? "justify-end" : "justify-start"}`}>
            <div
              className={`max-w-[80%] rounded-lg px-4 py-2 ${
                msg.role === "user"
                  ? "bg-gray-900 text-white"
                  : "bg-white border border-gray-200 text-gray-900"
              }`}
            >
              <p className="text-sm">{msg.text}</p>

              {/* タスク情報カード */}
              {msg.chatResponse?.task && (
                <div className="mt-2 p-2 bg-gray-50 rounded border text-xs space-y-1">
                  <div className="flex items-center gap-2">
                    <StatusBadge status={msg.chatResponse.task.status} />
                    <span className="text-gray-500">
                      ID: {msg.chatResponse.task.id.slice(0, 8)}
                    </span>
                  </div>
                  {msg.chatResponse.parsed_task && (
                    <div className="text-gray-500">
                      Action: {msg.chatResponse.parsed_task.action} |
                      Target: {msg.chatResponse.parsed_task.target_location} |
                      Confidence: {(msg.chatResponse.parsed_task.confidence * 100).toFixed(0)}%
                    </div>
                  )}
                </div>
              )}

              <p className="text-[10px] text-gray-400 mt-1">
                {msg.timestamp.toLocaleTimeString("ja-JP")}
              </p>
            </div>
          </div>
        ))}

        {loading && (
          <div className="flex justify-start">
            <div className="bg-white border border-gray-200 rounded-lg px-4 py-2">
              <div className="flex gap-1">
                <span className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: "0ms" }} />
                <span className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: "150ms" }} />
                <span className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: "300ms" }} />
              </div>
            </div>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>

      {/* 入力エリア */}
      <div className="flex gap-2">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleSend()}
          placeholder="ロボットへの指示を入力..."
          className="flex-1 border border-gray-300 rounded-lg px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900 focus:border-transparent"
          disabled={loading}
        />
        <button
          onClick={handleSend}
          disabled={loading || !input.trim()}
          className="bg-gray-900 text-white px-6 py-2 rounded-lg text-sm font-medium hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          送信
        </button>
      </div>
    </div>
  );
}
