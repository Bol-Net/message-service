package main

import (
	"fmt"
	"log"
	"net/http"

	"messaging-service/internal/api"
	"messaging-service/internal/auth"
	"messaging-service/internal/config"
	"messaging-service/internal/db"
	"messaging-service/internal/redis"
	ws "messaging-service/internal/websocket" // alias the internal package

	"github.com/gorilla/websocket" // alias Gorilla WebSocket
)

// upgrader for websocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	// load config
	cfg := config.LoadConfig()
	log.Printf("App running in %s mode", cfg.AppEnv)

	// init db
	if err := db.Init(cfg); err != nil {
		log.Fatal(err)
	}

	// init redis
	redis.Init(cfg)

	// create hub once
	hub := ws.NewHub()

	// websocket handler
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHandler(w, r, hub)
	})

	// register REST API routes (pass hub)
	api.RegisterRoutes(hub)

	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// websocket handler with JWT authentication
func wsHandler(w http.ResponseWriter, r *http.Request, hub *ws.Hub) {
	// Load public key for JWT verification
	pubKey, err := auth.LoadPublicKey()
	if err != nil {
		log.Println("Failed to load public key:", err)
		http.Error(w, "Server configuration error", http.StatusInternalServerError)
		return
	}

	// Extract JWT token from query parameter
	tokenStr, err := auth.ExtractQueryToken(r)
	if err != nil {
		log.Println("Token extraction failed:", err)
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Verify JWT token
	claims, err := auth.VerifyToken(tokenStr, pubKey)
	if err != nil {
		log.Println("Token verification failed:", err)
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Get user ID from JWT claims (not from query param)
	userID := claims.Subject
	if userID == "" {
		log.Println("user_id missing in JWT claims")
		http.Error(w, "Invalid token: missing user ID", http.StatusUnauthorized)
		return
	}

	// Upgrade to WebSocket
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade failed:", err)
		return
	}

	conn := ws.NewConnection(wsConn)

	// Create user data from JWT claims
	userData := redis.OnlineUser{
		ID:   userID,
		Name: claims.Name, // Use actual name from JWT token
		Role: claims.Role,
	}

	hub.Register(userID, conn, userData)

	fmt.Printf("Registering authenticated connection for user: %s (role: %s)\n", userID, claims.Role)
	go conn.ReadPump(hub, userID)
}
