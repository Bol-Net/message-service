package service

import (
	"messaging-service/internal/websocket"
	"time"
)

type Message struct {
	From    string
	To      string
	Content string
	Created time.Time
}

var batchInterval = 100 * time.Millisecond

func StartMessageBatching(hub *websocket.Hub, incoming <-chan Message) {
	ticker := time.NewTicker(batchInterval)
	defer ticker.Stop()

	batch := make(map[string][]Message)

	for {
		select {
		case msg := <-incoming:
			batch[msg.To] = append(batch[msg.To], msg)
		case <-ticker.C:
			for userID, messages := range batch {
				if len(messages) > 0 {
					// send all messages in batch
					for _, m := range messages {
						hub.SendMessage(userID, []byte(m.Content))
					}
					batch[userID] = nil
				}
			}
		}
	}
}
