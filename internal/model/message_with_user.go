package model

import (
	"messaging-service/internal/db"
	"time"
)

// MessageWithUser represents a message with sender and receiver user information
type MessageWithUser struct {
	ID           int       `json:"id"`
	SenderID     string    `json:"sender_id"`
	ReceiverID   string    `json:"receiver_id"`
	Content      string    `json:"content"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	SenderName   string    `json:"sender_name"`
	SenderEmail  string    `json:"sender_email"`
	ReceiverName string    `json:"receiver_name"`
	ReceiverEmail string   `json:"receiver_email"`
}

// GetMessagesBetweenWithUsers fetches messages between two users with user information
func GetMessagesBetweenWithUsers(userA, userB string) ([]MessageWithUser, error) {
	query := `
		SELECT 
			m.id, m.sender_id, m.receiver_id, m.content, m.status, m.created_at,
			s.name as sender_name, s.email as sender_email,
			r.name as receiver_name, r.email as receiver_email
		FROM messages m
		JOIN users s ON m.sender_id::int = s.id
		JOIN users r ON m.receiver_id::int = r.id
		WHERE (m.sender_id = $1 AND m.receiver_id = $2)
			 OR (m.sender_id = $2 AND m.receiver_id = $1)
		ORDER BY m.created_at ASC
	`
	
	rows, err := db.DB.Query(query, userA, userB)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []MessageWithUser
	for rows.Next() {
		var m MessageWithUser
		if err := rows.Scan(
			&m.ID, &m.SenderID, &m.ReceiverID, &m.Content, &m.Status, &m.CreatedAt,
			&m.SenderName, &m.SenderEmail, &m.ReceiverName, &m.ReceiverEmail,
		); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	
	return messages, nil
}

// GetAllMessagesWithUsers fetches all messages with user information (for admin/debugging)
func GetAllMessagesWithUsers() ([]MessageWithUser, error) {
	query := `
		SELECT 
			m.id, m.sender_id, m.receiver_id, m.content, m.status, m.created_at,
			s.name as sender_name, s.email as sender_email,
			r.name as receiver_name, r.email as receiver_email
		FROM messages m
		JOIN users s ON m.sender_id::int = s.id
		JOIN users r ON m.receiver_id::int = r.id
		ORDER BY m.created_at DESC
	`
	
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []MessageWithUser
	for rows.Next() {
		var m MessageWithUser
		if err := rows.Scan(
			&m.ID, &m.SenderID, &m.ReceiverID, &m.Content, &m.Status, &m.CreatedAt,
			&m.SenderName, &m.SenderEmail, &m.ReceiverName, &m.ReceiverEmail,
		); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	
	return messages, nil
}
