package api

import (
	"messaging-service/internal/auth"
	ws "messaging-service/internal/websocket"
	"net/http"
)

func RegisterRoutes(hub *ws.Hub) {
	// Protected endpoints with JWT authentication and CORS
	// Messages endpoint - get message history between users
	http.Handle("/api/messages", enableCORS(auth.JWTMiddleware(getMessagesHandler)))

	// Online users endpoint - get list of currently online users
	http.Handle("/api/online_users", enableCORS(auth.JWTMiddleware(getOnlineUsersHandler)))

	// Send message endpoint - send a message via REST API
	http.Handle("/api/send_message", enableCORS(auth.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		sendMessageHandler(w, r, hub)
	})))

	// Database info endpoint - shows which database is connected (no auth required for debugging)
	http.Handle("/api/db-info", enableCORS(http.HandlerFunc(DatabaseInfoHandler)))
}
