package model

import (
	"messaging-service/internal/db"
	"time"
)

type Message struct {
	ID         int       `db:"id" json:"id"`
	SenderID   string    `db:"sender_id" json:"sender_id"`
	ReceiverID string    `db:"receiver_id" json:"receiver_id"`
	Content    string    `db:"content" json:"content"`
	Status     string    `db:"status" json:"status"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// Save the message to the database
func (m *Message) Save() error {
	query := `
		INSERT INTO messages (sender_id, receiver_id, content, status, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at
	`
	return db.DB.QueryRow(query, m.SenderID, m.ReceiverID, m.Content, m.Status).
		Scan(&m.ID, &m.CreatedAt)
}

func (m *Message) UpdateStatus() error {
	query := `
		UPDATE messages
		SET status = $1
		WHERE id = $2
	`
	_, err := db.DB.Exec(query, m.Status, m.ID)
	return err
}

func GetMessagesBetween(userA, userB string) ([]Message, error) {
	query := `
			SELECT id, sender_id, receiver_id, content, status, created_at
			FROM messages
			WHERE (sender_id=$1 AND receiver_id=$2)
				 OR (sender_id=$2 AND receiver_id=$1)
			ORDER BY created_at ASC
	`
	rows, err := db.DB.Query(query, userA, userB)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Content, &m.Status, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}
