// service/websocket.go
package service

import (
	"context"
	"log"
	"net/http"
	"server/helpers"
	"server/internal/models"
	"server/internal/repository"
	"server/internal/ws"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Global variables for managing listener
var (
	listenerConn     *pgxpool.Conn
	listenerStarted  bool
	stopListener     chan bool
	listenerMutex    sync.Mutex
	listenerCtx      context.Context
	listenerCancel   context.CancelFunc
)

func WsHandler(w http.ResponseWriter, r *http.Request) {
	// Authenticate user from token in query parameter or Authorization header
	var userID int

	// Try to get token from query parameter first
	token := r.URL.Query().Get("token")

	// If not in query, try Authorization header
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

	// Convert userID from string to int
	userID, err = strconv.Atoi(claims.UserID)
	if err != nil {
		log.Println("Invalid user ID in token:", err)
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading: ", err)
		return
	}
	defer conn.Close()

	log.Printf("WebSocket client connected: %s (UserID: %d)", r.RemoteAddr, userID)

	// Register client with user ID
	client := &ws.Client{
		Conn:   conn,
		UserID: userID,
	}

	ws.ClientsMutex.Lock()
	ws.Clients[conn] = client
	isFirstClient := len(ws.Clients) == 1
	ws.ClientsMutex.Unlock()

	// Start listener if this is the first client
	if isFirstClient {
		go startDatabaseListener()
	}

	// Send initial data for this user only
	if err := sendCurrentModels(conn, userID); err != nil {
		log.Println("Error sending initial models:", err)
		return
	}

	// Keep connection alive and handle client messages
	for {
		// Read messages from client (or just check if connection is alive)
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		// Handle ping/pong or other messages
		if messageType == websocket.PingMessage {
			if err := conn.WriteMessage(websocket.PongMessage, nil); err != nil {
				log.Println("WebSocket pong error:", err)
				break
			}
		}

		log.Printf("Received message: %s", p)
	}

	// Unregister client
	ws.ClientsMutex.Lock()
	delete(ws.Clients, conn)
	shouldStopListener := len(ws.Clients) == 0
	ws.ClientsMutex.Unlock()

	// Stop listener if no clients left
	if shouldStopListener {
		stopDatabaseListener()
	}

	log.Println("WebSocket client disconnected:", r.RemoteAddr)
}

func startDatabaseListener() {
	listenerMutex.Lock()
	if listenerStarted {
		listenerMutex.Unlock()
		return
	}
	listenerStarted = true
	listenerCtx, listenerCancel = context.WithCancel(context.Background())
	listenerMutex.Unlock()

	log.Println("🎧 Starting PostgreSQL LISTEN for models_changes...")

	// Acquire a dedicated connection for listening
	conn, err := models.Pool.Acquire(listenerCtx)
	if err != nil {
		log.Println("❌ Failed to acquire connection for LISTEN:", err)
		listenerMutex.Lock()
		listenerStarted = false
		listenerCancel()
		listenerMutex.Unlock()
		return
	}

	listenerMutex.Lock()
	listenerConn = conn
	listenerMutex.Unlock()

	// Start listening on the channel
	_, err = conn.Exec(listenerCtx, "LISTEN models_changes")
	if err != nil {
		log.Println("❌ Failed to LISTEN:", err)
		conn.Release()
		listenerMutex.Lock()
		listenerStarted = false
		listenerConn = nil
		listenerCancel()
		listenerMutex.Unlock()
		return
	}

	log.Println("✅ Successfully started LISTEN on models_changes channel")

	// Listen for notifications in a loop
	defer func() {
		// Cleanup when exiting the listener
		conn.Exec(context.Background(), "UNLISTEN models_changes")
		conn.Release()
		listenerMutex.Lock()
		listenerStarted = false
		listenerConn = nil
		listenerMutex.Unlock()
		log.Println("✅ Database listener cleanup complete")
	}()

	for {
		select {
		case <-listenerCtx.Done():
			log.Println("🛑 Stopping database listener...")
			return

		default:
			// Wait for notification with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			notification, err := conn.Conn().WaitForNotification(ctx)
			cancel()

			if err != nil {
				// Timeout is normal, just continue
				if ctx.Err() == context.DeadlineExceeded {
					// Check if we should stop
					select {
					case <-listenerCtx.Done():
						log.Println("🛑 Stopping database listener...")
						return
					default:
						continue
					}
				}
				log.Println("❌ Error waiting for notification:", err)
				time.Sleep(1 * time.Second)
				continue
			}

			// Notification received!
			log.Printf("🔔 Received notification: %s - %s", notification.Channel, notification.Payload)

			// Fetch updated models and broadcast
			broadcastModelsToClients()
		}
	}
}

func stopDatabaseListener() {
	listenerMutex.Lock()
	defer listenerMutex.Unlock()

	if !listenerStarted {
		return
	}

	log.Println("Stopping database listener (no clients connected)...")
	if listenerCancel != nil {
		listenerCancel()
	}
}

func broadcastModelsToClients() {
	ctx := context.Background()

	// Broadcast to all connected clients - each gets only their own models
	ws.ClientsMutex.Lock()
	defer ws.ClientsMutex.Unlock()

	successCount := 0
	for conn, client := range ws.Clients {
		// Fetch models for this specific user
		userModels, err := repository.GetModelsByUserID(ctx, client.UserID)
		if err != nil {
			log.Printf("❌ GetModelsByUserID error for user %d: %v", client.UserID, err)
			continue
		}

		if userModels == nil {
			userModels = []map[string]interface{}{}
		}

		if err := conn.WriteJSON(userModels); err != nil {
			log.Println("❌ Error broadcasting to client:", err)
			conn.Close()
			delete(ws.Clients, conn)
		} else {
			successCount++
		}
	}

	log.Printf("✅ Broadcasted models update to %d clients", successCount)
}

func sendCurrentModels(conn *websocket.Conn, userID int) error {
	ctx := context.Background()
	userModels, err := repository.GetModelsByUserID(ctx, userID)
	if err != nil {
		log.Printf("❌ GetModelsByUserID error for user %d: %v", userID, err)
		return err
	}

	if userModels == nil {
		userModels = []map[string]interface{}{}
	}

	if err := conn.WriteJSON(userModels); err != nil {
		log.Println("❌ WebSocket send error:", err)
		return err
	}

	log.Printf("✅ Sent initial models to client (UserID: %d, Count: %d)", userID, len(userModels))
	return nil
}
