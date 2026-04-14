package handler

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WSHub はフロントエンドへのWebSocket配信を管理する
type WSHub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]bool
}

func NewWSHub() *WSHub {
	return &WSHub{clients: make(map[*websocket.Conn]bool)}
}

// HandleWS はWebSocket接続をアップグレードし、クライアントを登録する
func (h *WSHub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocketアップグレードエラー: %v", err)
		return
	}

	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()
	log.Printf("フロントエンドWebSocket接続: %s (計%d)", conn.RemoteAddr(), len(h.clients))

	// 読み取りループ（切断検知用）
	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.clients, conn)
			h.mu.Unlock()
			conn.Close()
			log.Printf("フロントエンドWebSocket切断: %s", conn.RemoteAddr())
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}

// Broadcast は全クライアントにJSONメッセージを送信する
func (h *WSHub) Broadcast(msgType string, data any) {
	payload := map[string]any{
		"type": msgType,
		"data": data,
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range h.clients {
		if err := conn.WriteJSON(payload); err != nil {
			log.Printf("WebSocket送信エラー: %v", err)
			conn.Close()
			delete(h.clients, conn)
		}
	}
}
