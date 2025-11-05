// service/trainingWS.go
package service

import (
	"log"
	"net/http"
	"server/aiAgent"
	"server/helpers"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// TrainingClient represents a WebSocket connection for training updates
type TrainingClient struct {
	Conn       *websocket.Conn
	UserID     int
	TrainingID string // Optional: filter updates for specific training
}

// TrainingBroadcaster manages WebSocket connections for training updates
type TrainingBroadcaster struct {
	clients      map[*websocket.Conn]*TrainingClient
	clientsMutex sync.RWMutex
	upgrader     websocket.Upgrader
}

// Global broadcaster instance
var trainingBroadcaster *TrainingBroadcaster
var broadcasterOnce sync.Once

// GetTrainingBroadcaster returns the singleton broadcaster instance
func GetTrainingBroadcaster() *TrainingBroadcaster {
	broadcasterOnce.Do(func() {
		trainingBroadcaster = &TrainingBroadcaster{
			clients: make(map[*websocket.Conn]*TrainingClient),
			upgrader: websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			},
		}
	})
	return trainingBroadcaster
}

// TrainingWSHandler handles WebSocket connections for training updates
func TrainingWSHandler(w http.ResponseWriter, r *http.Request) {
	broadcaster := GetTrainingBroadcaster()

	// Authenticate user from token
	var userID int
	token := r.URL.Query().Get("token")

	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if token == "" {
		http.Error(w, "Missing authentication token", http.StatusUnauthorized)
		return
	}

	// Validate JWT and extract user ID
	claims, err := helpers.ValidateJWT(token)
	if err != nil {
		log.Println("Invalid JWT token:", err)
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	userID, err = strconv.Atoi(claims.UserID)
	if err != nil {
		log.Println("Invalid user ID in token:", err)
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// Get optional training ID filter
	trainingID := r.URL.Query().Get("training_id")

	// Upgrade connection
	conn, err := broadcaster.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	log.Printf("🔌 Training WebSocket connected: UserID=%d, TrainingID=%s", userID, trainingID)

	// Register client
	client := &TrainingClient{
		Conn:       conn,
		UserID:     userID,
		TrainingID: trainingID,
	}

	broadcaster.clientsMutex.Lock()
	broadcaster.clients[conn] = client
	broadcaster.clientsMutex.Unlock()

	// Send initial connection success message
	conn.WriteJSON(map[string]interface{}{
		"type":    "connected",
		"message": "Connected to training updates",
		"user_id": userID,
	})

	// Keep connection alive
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Training WebSocket read error:", err)
			break
		}

		// Handle ping/pong
		if messageType == websocket.PingMessage {
			if err := conn.WriteMessage(websocket.PongMessage, nil); err != nil {
				log.Println("Training WebSocket pong error:", err)
				break
			}
		}

		log.Printf("Received training WS message: %s", p)
	}

	// Unregister client
	broadcaster.clientsMutex.Lock()
	delete(broadcaster.clients, conn)
	broadcaster.clientsMutex.Unlock()

	log.Printf("🔌 Training WebSocket disconnected: UserID=%d", userID)
}

// BroadcastTrainingUpdate sends a training update to all connected clients
func (b *TrainingBroadcaster) BroadcastTrainingUpdate(trainingID string, updateType string, data interface{}) {
	b.clientsMutex.RLock()
	defer b.clientsMutex.RUnlock()

	message := map[string]interface{}{
		"type":        updateType,
		"training_id": trainingID,
		"data":        data,
	}

	// Send to all clients (or filter by trainingID if they subscribed to specific training)
	for conn, client := range b.clients {
		// If client subscribed to specific training, only send updates for that training
		if client.TrainingID != "" && client.TrainingID != trainingID {
			continue
		}

		if err := conn.WriteJSON(message); err != nil {
			log.Printf("❌ Error broadcasting training update to client %d: %v", client.UserID, err)
			conn.Close()
			delete(b.clients, conn)
		}
	}
}

// BroadcastLog sends a log message to all connected clients
func (b *TrainingBroadcaster) BroadcastLog(trainingID string, logLine string, isError bool) {
	b.BroadcastTrainingUpdate(trainingID, "log", map[string]interface{}{
		"message":  logLine,
		"is_error": isError,
	})
}

// BroadcastMetrics sends metrics update to all connected clients
func (b *TrainingBroadcaster) BroadcastMetrics(trainingID string, metrics *aiAgent.TrainingMetrics) {
	b.BroadcastTrainingUpdate(trainingID, "metrics", metrics)
}

// BroadcastStatus sends status update to all connected clients
func (b *TrainingBroadcaster) BroadcastStatus(trainingID string, status aiAgent.TrainingStatus, errorMessage string) {
	b.BroadcastTrainingUpdate(trainingID, "status", map[string]interface{}{
		"status":        status,
		"error_message": errorMessage,
	})
}

// BroadcastProgress sends overall progress update to all connected clients
func (b *TrainingBroadcaster) BroadcastProgress(trainingID string, progress *aiAgent.TrainingProgress) {
	b.BroadcastTrainingUpdate(trainingID, "progress", map[string]interface{}{
		"status":        progress.Status,
		"current_epoch": progress.CurrentEpoch,
		"total_epochs":  progress.TotalEpochs,
		"start_time":    progress.StartTime,
		"end_time":      progress.EndTime,
	})
}
