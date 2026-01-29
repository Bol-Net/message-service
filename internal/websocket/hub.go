package websocket

import (
	"log"
	"messaging-service/internal/redis"
	"sync"
)

// Hub maintains active WebSocket connections.
type Hub struct {
	connections map[string]*Connection
	mu          sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		connections: make(map[string]*Connection),
	}
}

func (h *Hub) Register(userID string, conn *Connection, userData redis.OnlineUser) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connections[userID] = conn

	// Mark user as online with user data
	if err := redis.MarkOnline(userID, userData); err != nil {
		log.Println("error marking user as online:", err)
	}
}

func (h *Hub) Unregister(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.connections, userID)

	// Mark user as offline
	if err := redis.MarkOffline(userID); err != nil {
		log.Println("error marking user as offline:", err)
	}
}

func (h *Hub) SendMessage(receiverID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if conn, ok := h.connections[receiverID]; ok {
		conn.Send(message)
	}
}

// helper: check if user is online
func (h *Hub) IsOnline(userID string) bool {
	return redis.IsOnline(userID)
}

func GetOnlineUsers() ([]redis.OnlineUser, error) {
	return redis.GetOnlineUsers()
}
