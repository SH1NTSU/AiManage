package ws

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a WebSocket connection with its associated user ID
type Client struct {
	Conn   *websocket.Conn
	UserID int
}

// Global variables for managing clients
var (
	ClientsMutex sync.Mutex
	Clients      = make(map[*websocket.Conn]*Client)
)

// BroadcastAgentStatus broadcasts agent status to all WebSocket clients for a specific user
func BroadcastAgentStatus(userID int, status map[string]interface{}) {
	ClientsMutex.Lock()
	defer ClientsMutex.Unlock()

	// Add a type field to distinguish from model updates
	message := map[string]interface{}{
		"type": "agent_status",
		"data": status,
	}

	successCount := 0
	for conn, client := range Clients {
		if client.UserID == userID {
			if err := conn.WriteJSON(message); err != nil {
				log.Printf("❌ Error broadcasting agent status to client: %v", err)
				conn.Close()
				delete(Clients, conn)
			} else {
				successCount++
			}
		}
	}

	if successCount > 0 {
		log.Printf("✅ Broadcasted agent status to %d client(s) for user %d", successCount, userID)
	}
}

// BroadcastToUser broadcasts a message to all WebSocket clients for a specific user
func BroadcastToUser(userID int, message map[string]interface{}) {
	ClientsMutex.Lock()
	defer ClientsMutex.Unlock()

	successCount := 0
	for conn, client := range Clients {
		if client.UserID == userID {
			if err := conn.WriteJSON(message); err != nil {
				log.Printf("❌ Error broadcasting to client: %v", err)
				conn.Close()
				delete(Clients, conn)
			} else {
				successCount++
			}
		}
	}

	if successCount > 0 {
		msgType := message["type"]
		log.Printf("✅ Broadcasted %v to %d client(s) for user %d", msgType, successCount, userID)
	}
}
