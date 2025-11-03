// service/websocket.go
package service

import (
	"context"
	"log"
	"net/http"
	"server/helpers"
	"server/internal/models"
	"server/internal/repository"
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

// Client represents a WebSocket connection with its associated user ID
type Client struct {
	Conn   *websocket.Conn
	UserID int
}

// Global variables for managing clients and listener
var (
	clientsMutex     sync.Mutex
	clients          = make(map[*websocket.Conn]*Client)
	listenerConn     *pgxpool.Conn
	listenerStarted  bool
	stopListener     chan bool
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
	client := &Client{
		Conn:   conn,
		UserID: userID,
	}

	clientsMutex.Lock()
	clients[conn] = client
	isFirstClient := len(clients) == 1
	clientsMutex.Unlock()

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
	clientsMutex.Lock()
	delete(clients, conn)
	shouldStopListener := len(clients) == 0
	clientsMutex.Unlock()

	// Stop listener if no clients left
	if shouldStopListener {
		stopDatabaseListener()
	}

	log.Println("WebSocket client disconnected:", r.RemoteAddr)
}

func startDatabaseListener() {
	clientsMutex.Lock()
	if listenerStarted {
		clientsMutex.Unlock()
		return
	}
	listenerStarted = true
	stopListener = make(chan bool)
	clientsMutex.Unlock()

	log.Println("🎧 Starting PostgreSQL LISTEN for models_changes...")

	ctx := context.Background()

	// Acquire a dedicated connection for listening
	conn, err := models.Pool.Acquire(ctx)
	if err != nil {
		log.Println("❌ Failed to acquire connection for LISTEN:", err)
		clientsMutex.Lock()
		listenerStarted = false
		clientsMutex.Unlock()
		return
	}

	clientsMutex.Lock()
	listenerConn = conn
	clientsMutex.Unlock()

	// Start listening on the channel
	_, err = conn.Exec(ctx, "LISTEN models_changes")
	if err != nil {
		log.Println("❌ Failed to LISTEN:", err)
		conn.Release()
		clientsMutex.Lock()
		listenerStarted = false
		listenerConn = nil
		clientsMutex.Unlock()
		return
	}

	log.Println("✅ Successfully started LISTEN on models_changes channel")

	// Listen for notifications in a loop
	for {
		select {
		case <-stopListener:
			log.Println("🛑 Stopping database listener...")
			conn.Exec(ctx, "UNLISTEN models_changes")
			conn.Release()
			clientsMutex.Lock()
			listenerStarted = false
			listenerConn = nil
			clientsMutex.Unlock()
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
					case <-stopListener:
						log.Println("🛑 Stopping database listener...")
						conn.Exec(context.Background(), "UNLISTEN models_changes")
						conn.Release()
						clientsMutex.Lock()
						listenerStarted = false
						listenerConn = nil
						clientsMutex.Unlock()
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
	clientsMutex.Lock()
	if !listenerStarted {
		clientsMutex.Unlock()
		return
	}
	clientsMutex.Unlock()

	log.Println("Stopping database listener (no clients connected)...")
	if stopListener != nil {
		close(stopListener)
	}
}

func broadcastModelsToClients() {
	ctx := context.Background()

	// Broadcast to all connected clients - each gets only their own models
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	successCount := 0
	for conn, client := range clients {
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
			delete(clients, conn)
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
