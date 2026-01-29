package websocket

import (
	"encoding/json"
	"log"
	"messaging-service/internal/model"

	"github.com/gorilla/websocket"
)

type Connection struct {
	WS       *websocket.Conn
	SendChan chan []byte
}

func NewConnection(ws *websocket.Conn) *Connection {
	c := &Connection{
		WS:       ws,
		SendChan: make(chan []byte, 256),
	}
	go c.writePump()
	return c
}

func (c *Connection) Send(message []byte) {
	select {
	case c.SendChan <- message:
	default:
		log.Println("Send channel full, dropping message")
	}
}

func (c *Connection) writePump() {
	for msg := range c.SendChan {
		err := c.WS.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("Write error:", err)
			break
		}
	}
	c.WS.Close()
}

func (c *Connection) ReadPump(hub *Hub, userID string) {
	defer func() {
		hub.Unregister(userID)
		c.WS.Close()
	}()

	for {
		_, msg, err := c.WS.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		var m model.Message
		if err := json.Unmarshal(msg, &m); err != nil {
			log.Println("invalid message:", err)
			continue
		}

		// Enforce sender
		m.SenderID = userID
		m.Status = "sent"

		// Save message to DB
		if err := m.Save(); err != nil {
			log.Println("error saving message:", err)
			continue
		}

		// Skip sending back to self
		if m.ReceiverID == userID {
			continue
		}

		// Marshal JSON to send to recipient
		data, _ := json.Marshal(m)

		// Send to recipient
		hub.SendMessage(m.ReceiverID, data)

		// Update message status to delivered
		m.Status = "delivered"
		if err := m.UpdateStatus(); err != nil {
			log.Println("error updating status:", err)
		}

		// Notify sender that message was delivered
		ack := map[string]interface{}{
			"type":       "delivery",
			"message_id": m.ID,
			"to":         m.ReceiverID,
			"status":     m.Status,
		}

		ackData, _ := json.Marshal(ack)
		hub.SendMessage(userID, ackData)
	}
}
