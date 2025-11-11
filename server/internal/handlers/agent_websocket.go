package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"server/internal/middlewares"
	"server/internal/repository"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now - restrict in production
	},
}

// AgentConnection represents a connected training agent
type AgentConnection struct {
	Conn      *websocket.Conn
	UserEmail string
	ApiKey    string
	LastPing  time.Time
	IsTraining bool
	mu        sync.Mutex
}

// AgentManager manages all connected agents
type AgentManager struct {
	agents map[string]*AgentConnection // key: user email
	mu     sync.RWMutex
}

var agentManager = &AgentManager{
	agents: make(map[string]*AgentConnection),
}

// AgentWebSocketHandler handles WebSocket connections from training agents
func AgentWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Get API key from query params
	apiKey := r.URL.Query().Get("api_key")
	if apiKey == "" {
		http.Error(w, "API key required", http.StatusUnauthorized)
		return
	}

	// Validate API key and get user
	user, err := repository.GetUserByApiKey(context.Background(), apiKey)
	if err != nil {
		log.Printf("❌ Database error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		log.Printf("❌ Invalid API key")
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	userEmail, ok := (*user)["email"].(string)
	if !ok {
		log.Printf("❌ User email not found")
		http.Error(w, "Invalid user data", http.StatusInternalServerError)
		return
	}

	log.Printf("🔌 Agent connection request from: %s", userEmail)

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ Failed to upgrade connection: %v", err)
		return
	}

	// Create agent connection
	agent := &AgentConnection{
		Conn:      conn,
		UserEmail: userEmail,
		ApiKey:    apiKey,
		LastPing:  time.Now(),
		IsTraining: false,
	}

	// Register agent
	agentManager.mu.Lock()
	agentManager.agents[userEmail] = agent
	agentManager.mu.Unlock()

	log.Printf("✅ Agent connected: %s", userEmail)

	// Send welcome message
	agent.SendMessage(map[string]interface{}{
		"type":    "connected",
		"message": "Welcome! Agent connected successfully",
	})

	// Request system info
	agent.SendMessage(map[string]interface{}{
		"type": "system_info_request",
	})

	// Handle messages
	go agent.HandleMessages()

	// Ping loop
	go agent.PingLoop()
}

// HandleMessages processes messages from the agent
func (ac *AgentConnection) HandleMessages() {
	defer func() {
		// Cleanup on disconnect
		agentManager.mu.Lock()
		delete(agentManager.agents, ac.UserEmail)
		agentManager.mu.Unlock()
		ac.Conn.Close()
		log.Printf("👋 Agent disconnected: %s", ac.UserEmail)
	}()

	for {
		_, message, err := ac.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ WebSocket error: %v", err)
			}
			break
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("❌ Failed to parse message: %v", err)
			continue
		}

		msgType, ok := msg["type"].(string)
		if !ok {
			continue
		}

		switch msgType {
		case "pong":
			ac.mu.Lock()
			ac.LastPing = time.Now()
			ac.mu.Unlock()

		case "system_info":
			data := msg["data"]
			log.Printf("📊 System info from %s: %v", ac.UserEmail, data)

		case "training_started":
			ac.mu.Lock()
			ac.IsTraining = true
			ac.mu.Unlock()
			trainingID := msg["training_id"]
			log.Printf("🚀 Training started: %v", trainingID)

		case "training_output":
			output := msg["output"]
			log.Printf("📝 Training output: %v", output)
			// TODO: Broadcast to web clients via WebSocket

		case "training_completed":
			ac.mu.Lock()
			ac.IsTraining = false
			ac.mu.Unlock()
			trainingID := msg["training_id"]
			log.Printf("✅ Training completed: %v", trainingID)

		case "training_failed":
			ac.mu.Lock()
			ac.IsTraining = false
			ac.mu.Unlock()
			trainingID := msg["training_id"]
			error := msg["error"]
			log.Printf("❌ Training failed: %v - %v", trainingID, error)

		case "error":
			error := msg["message"]
			log.Printf("❌ Agent error: %v", error)
		}
	}
}

// SendMessage sends a message to the agent
func (ac *AgentConnection) SendMessage(data map[string]interface{}) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	return ac.Conn.WriteJSON(data)
}

// PingLoop sends periodic pings to keep connection alive
func (ac *AgentConnection) PingLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := ac.SendMessage(map[string]interface{}{"type": "ping"}); err != nil {
			return
		}

		// Check if agent is still alive
		ac.mu.Lock()
		if time.Since(ac.LastPing) > 2*time.Minute {
			ac.mu.Unlock()
			log.Printf("⚠️  Agent timeout: %s", ac.UserEmail)
			ac.Conn.Close()
			return
		}
		ac.mu.Unlock()
	}
}

// StartRemoteTraining sends a training command to the user's agent
func StartRemoteTraining(userEmail string, trainingData map[string]interface{}) error {
	agentManager.mu.RLock()
	agent, exists := agentManager.agents[userEmail]
	agentManager.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no agent connected for user: %s", userEmail)
	}

	agent.mu.Lock()
	if agent.IsTraining {
		agent.mu.Unlock()
		return fmt.Errorf("agent is already training a model")
	}
	agent.mu.Unlock()

	return agent.SendMessage(map[string]interface{}{
		"type": "train",
		"data": trainingData,
	})
}

// IsAgentConnected checks if a user has an agent connected
func IsAgentConnected(userEmail string) bool {
	agentManager.mu.RLock()
	defer agentManager.mu.RUnlock()

	agent, exists := agentManager.agents[userEmail]
	if !exists {
		return false
	}

	// Check if agent is alive
	agent.mu.Lock()
	defer agent.mu.Unlock()
	return time.Since(agent.LastPing) < 2*time.Minute
}

// GetAgentStatus returns the status of a user's agent
func GetAgentStatusHandler(w http.ResponseWriter, r *http.Request) {
	userEmail, ok := r.Context().Value(middlewares.UserEmailKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	isConnected := IsAgentConnected(userEmail)

	var status string
	var systemInfo interface{}

	agentManager.mu.RLock()
	agent, exists := agentManager.agents[userEmail]
	agentManager.mu.RUnlock()

	if exists && isConnected {
		agent.mu.Lock()
		status = "connected"
		if agent.IsTraining {
			status = "training"
		}
		agent.mu.Unlock()
	} else {
		status = "disconnected"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"status":      status,
		"connected":   isConnected,
		"system_info": systemInfo,
	})
}
