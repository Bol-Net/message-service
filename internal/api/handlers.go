package api

import (
	"encoding/json"
	"fmt"
	"messaging-service/internal/auth"
	"messaging-service/internal/db"
	"messaging-service/internal/model"
	"messaging-service/internal/websocket"
	"net/http"
)

func getMessagesHandler(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from JWT context
	userA, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Authentication error", http.StatusUnauthorized)
		return
	}

	// Get the other user from query parameter
	userB := r.URL.Query().Get("with")
	if userB == "" {
		http.Error(w, "with query param required", http.StatusBadRequest)
		return
	}

	messages, err := model.GetMessagesBetweenWithUsers(userA, userB)
	fmt.Println("error", err)
	if err != nil {
		http.Error(w, "Error fetching messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func getOnlineUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Get online users from Redis (which already has user data)
	onlineUsers, err := websocket.GetOnlineUsers()
	if err != nil {
		http.Error(w, "Error fetching online users", http.StatusInternalServerError)
		return
	}

	// Extract user IDs to get full user details from database
	var userIDStrings []string
	for _, user := range onlineUsers {
		userIDStrings = append(userIDStrings, user.ID)
	}

	// Get complete user details from database
	userMap, err := model.GetUsersByIDs(userIDStrings)
	if err != nil {
		// Fallback to Redis data if database query fails
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(onlineUsers)
		return
	}

	// Convert map to slice for JSON response with full user details
	var usersWithDetails []model.User
	for _, user := range userMap {
		usersWithDetails = append(usersWithDetails, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usersWithDetails)
}

func sendMessageHandler(w http.ResponseWriter, r *http.Request, hub *websocket.Hub) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user ID from JWT context
	senderID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Authentication error", http.StatusUnauthorized)
		return
	}

	var m model.Message
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Enforce sender ID from JWT (prevent spoofing)
	m.SenderID = senderID
	m.Status = "sent"

	// Save to DB
	if err := m.Save(); err != nil {
		http.Error(w, "Error saving message", http.StatusInternalServerError)
		return
	}

	// Send via WebSocket if recipient online
	data, _ := json.Marshal(m)
	hub.SendMessage(m.ReceiverID, data)

	// Update status to delivered
	m.Status = "delivered"
	m.UpdateStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// DatabaseInfoHandler returns information about the connected database
func DatabaseInfoHandler(w http.ResponseWriter, r *http.Request) {
	type DatabaseInfo struct {
		DatabaseName    string   `json:"database_name"`
		Version         string   `json:"version"`
		TableCount      int      `json:"table_count"`
		KeyTables       []string `json:"key_tables"`
		ConnectionValid bool     `json:"connection_valid"`
	}

	info := DatabaseInfo{
		ConnectionValid: false,
	}

	// Test connection
	if err := db.DB.Ping(); err != nil {
		http.Error(w, fmt.Sprintf("Database connection failed: %v", err), http.StatusInternalServerError)
		return
	}
	info.ConnectionValid = true

	// Get database name
	if err := db.DB.QueryRow("SELECT current_database()").Scan(&info.DatabaseName); err != nil {
		http.Error(w, fmt.Sprintf("Failed to get database name: %v", err), http.StatusInternalServerError)
		return
	}

	// Get PostgreSQL version
	if err := db.DB.QueryRow("SELECT version()").Scan(&info.Version); err != nil {
		http.Error(w, fmt.Sprintf("Failed to get database version: %v", err), http.StatusInternalServerError)
		return
	}

	// Get table count
	if err := db.DB.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&info.TableCount); err != nil {
		http.Error(w, fmt.Sprintf("Failed to get table count: %v", err), http.StatusInternalServerError)
		return
	}

	// Get key tables
	rows, err := db.DB.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get table list: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		info.KeyTables = append(info.KeyTables, tableName)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
