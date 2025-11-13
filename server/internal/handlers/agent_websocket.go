package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"server/aiAgent"
	"server/internal/middlewares"
	"server/internal/repository"
	"server/internal/ws"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now - restrict in production
	},
}

// AgentConnection represents a connected training agent
type AgentConnection struct {
	Conn       *websocket.Conn
	UserEmail  string
	ApiKey     string
	LastPing   time.Time
	IsTraining bool
	SystemInfo map[string]interface{}
	UserID     int
	mu         sync.Mutex
}

// AgentManager manages all connected agents
type AgentManager struct {
	agents map[string]*AgentConnection // key: user email
	mu     sync.RWMutex
}

var agentManager = &AgentManager{
	agents: make(map[string]*AgentConnection),
}

// Global trainer reference for storing remote training progress
var globalTrainer *aiAgent.Trainer

// SetGlobalTrainer sets the trainer instance for agent-based training
func SetGlobalTrainer(trainer *aiAgent.Trainer) {
	globalTrainer = trainer
}

// AgentWebSocketHandler handles WebSocket connections from training agents
func AgentWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔌 New agent connection attempt from %s", r.RemoteAddr)

	// Get API key from query params
	apiKey := r.URL.Query().Get("api_key")
	if apiKey == "" {
		log.Printf("❌ Connection rejected: No API key provided")
		http.Error(w, "API key required", http.StatusUnauthorized)
		return
	}

	// Log API key prefix for debugging (first 8 chars or less)
	apiKeyPrefix := apiKey
	if len(apiKey) > 8 {
		apiKeyPrefix = apiKey[:8] + "..."
	}
	log.Printf("🔑 Validating API key: %s", apiKeyPrefix)

	// Validate API key and get user
	user, err := repository.GetUserByApiKey(context.Background(), apiKey)
	if err != nil {
		log.Printf("❌ Database error while validating API key: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		log.Printf("❌ Invalid API key - no user found")
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	userEmail, ok := (*user)["email"].(string)
	if !ok {
		log.Printf("❌ User email not found in database result")
		http.Error(w, "Invalid user data", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ API key valid for user: %s", userEmail)

	// Get user ID for broadcasting
	userID, ok := (*user)["id"].(int)
	if !ok {
		// Try converting from int32 or other types
		if id32, ok := (*user)["id"].(int32); ok {
			userID = int(id32)
		} else {
			log.Printf("❌ Could not convert user ID to int")
			http.Error(w, "Invalid user ID", http.StatusInternalServerError)
			return
		}
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ Failed to upgrade connection: %v", err)
		return
	}

	// Create agent connection
	agent := &AgentConnection{
		Conn:       conn,
		UserEmail:  userEmail,
		ApiKey:     apiKey,
		LastPing:   time.Now(),
		IsTraining: false,
		SystemInfo: nil,
		UserID:     userID,
	}

	// Register agent
	agentManager.mu.Lock()
	agentManager.agents[userEmail] = agent
	agentManager.mu.Unlock()

	log.Printf("✅ Agent connected: %s", userEmail)

	// Broadcast agent connected status to all WebSocket clients for this user
	ws.BroadcastAgentStatus(userID, map[string]interface{}{
		"connected":   true,
		"status":      "connected",
		"system_info": nil, // Will be updated when system_info arrives
	})

	// Send welcome message
	if err := agent.SendMessage(map[string]interface{}{
		"type":    "connected",
		"message": "Welcome! Agent connected successfully",
	}); err != nil {
		log.Printf("⚠️  Failed to send welcome message: %v", err)
	} else {
		log.Printf("📤 Welcome message sent to %s", userEmail)
	}

	// Request system info
	if err := agent.SendMessage(map[string]interface{}{
		"type": "system_info_request",
	}); err != nil {
		log.Printf("⚠️  Failed to request system info: %v", err)
	} else {
		log.Printf("📤 System info requested from %s", userEmail)
	}

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

		// Broadcast agent disconnected status
		ws.BroadcastAgentStatus(ac.UserID, map[string]interface{}{
			"connected":   false,
			"status":      "disconnected",
			"system_info": nil,
		})
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
			// Store system info
			ac.mu.Lock()
			if dataMap, ok := data.(map[string]interface{}); ok {
				ac.SystemInfo = dataMap
			}
			ac.mu.Unlock()

			// Broadcast updated agent status with system info
			ws.BroadcastAgentStatus(ac.UserID, map[string]interface{}{
				"connected":   true,
				"status":      "connected",
				"system_info": data,
			})

		case "training_started":
			ac.mu.Lock()
			ac.IsTraining = true
			ac.mu.Unlock()
			trainingIDInterface := msg["training_id"]
			trainingID, _ := trainingIDInterface.(string)
			log.Printf("🚀 Training started: %v", trainingID)

			// Create training progress entry in trainer
			if globalTrainer != nil && trainingID != "" {
				createRemoteTrainingProgress(trainingID)
			}

			// Broadcast training started to frontend
			ws.BroadcastToUser(ac.UserID, map[string]interface{}{
				"type": "training_update",
				"data": map[string]interface{}{
					"training_id": trainingID,
					"status":      "running",
					"message":     "Training started on local agent",
				},
			})

		case "training_output":
			trainingIDInterface := msg["training_id"]
			trainingID, _ := trainingIDInterface.(string)
			outputInterface := msg["output"]
			output, _ := outputInterface.(string)
			log.Printf("📝 Training output: %v", output)

			// Update training progress with parsed output
			if globalTrainer != nil && trainingID != "" {
				updateRemoteTrainingProgress(trainingID, output)
			}

			// Broadcast training output to frontend
			ws.BroadcastToUser(ac.UserID, map[string]interface{}{
				"type": "training_output",
				"data": map[string]interface{}{
					"training_id": trainingID,
					"output":      output,
				},
			})

		case "training_completed":
			ac.mu.Lock()
			ac.IsTraining = false
			ac.mu.Unlock()
			trainingIDInterface := msg["training_id"]
			trainingID, _ := trainingIDInterface.(string)
			modelPathInterface := msg["model_path"]
			modelPath, _ := modelPathInterface.(string)
			log.Printf("✅ Training completed: %v", trainingID)
			if modelPath != "" {
				log.Printf("💾 Trained model path: %v", modelPath)
			}

			// Mark training as completed and update database with model path
			if globalTrainer != nil && trainingID != "" {
				markRemoteTrainingCompleted(trainingID, modelPath)
			}

			// Broadcast training completed to frontend
			ws.BroadcastToUser(ac.UserID, map[string]interface{}{
				"type": "training_update",
				"data": map[string]interface{}{
					"training_id": trainingID,
					"status":      "completed",
					"message":     "Training completed successfully!",
					"model_path":  modelPath,
				},
			})

		case "training_failed":
			ac.mu.Lock()
			ac.IsTraining = false
			ac.mu.Unlock()
			trainingIDInterface := msg["training_id"]
			trainingID, _ := trainingIDInterface.(string)
			errorInterface := msg["error"]
			error, _ := errorInterface.(string)
			log.Printf("❌ Training failed: %v - %v", trainingID, error)

			// Mark training as failed
			if globalTrainer != nil && trainingID != "" {
				markRemoteTrainingFailed(trainingID, error)
			}

			// Broadcast training failed to frontend
			ws.BroadcastToUser(ac.UserID, map[string]interface{}{
				"type": "training_update",
				"data": map[string]interface{}{
					"training_id":   trainingID,
					"status":        "failed",
					"error_message": error,
				},
			})

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
		systemInfo = agent.SystemInfo
		agent.mu.Unlock()
	} else {
		status = "disconnected"
	}

	log.Printf("📊 Agent status for %s: connected=%v, status=%s", userEmail, isConnected, status)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"status":      status,
		"connected":   isConnected,
		"system_info": systemInfo,
	})
}

// Helper functions for remote training progress

func createRemoteTrainingProgress(trainingID string) {
	progress := &aiAgent.TrainingProgress{
		Status:      aiAgent.StatusRunning,
		StartTime:   time.Now(),
		Logs:        []string{},
		Metrics:     []aiAgent.TrainingMetrics{},
		TotalEpochs: 0,
	}

	globalTrainer.StoreTrainingProgress(trainingID, progress)
	log.Printf("📊 Created remote training progress: %s", trainingID)
}

func updateRemoteTrainingProgress(trainingID string, output string) {
	progress, err := globalTrainer.GetProgress(trainingID)
	if err != nil {
		log.Printf("⚠️  Failed to get progress for %s: %v", trainingID, err)
		return
	}

	// Add log
	progress.AddLog(output)

	// Try to parse metrics from output
	if metrics := parseMetricsFromOutput(output); metrics != nil {
		progress.AddMetrics(*metrics)
		log.Printf("📈 Parsed metrics: Epoch %d/%d, Loss: %.4f",
			metrics.Epoch, metrics.TotalEpochs, metrics.TrainLoss)
	}
}

func markRemoteTrainingCompleted(trainingID string, modelPath string) {
	progress, err := globalTrainer.GetProgress(trainingID)
	if err != nil {
		log.Printf("⚠️  Failed to get progress for %s: %v", trainingID, err)
		return
	}

	progress.MarkCompleted()

	// Set model path if provided
	if modelPath != "" {
		progress.SetModelPath(modelPath)
		log.Printf("💾 Set model path: %s", modelPath)

		// Extract model name from training ID (format: "ModelName_timestamp")
		modelName := extractModelName(trainingID)
		if modelName != "" {
			log.Printf("📝 Updating database for model: %s", modelName)

			// Update database with trained model path
			ctx := context.Background()
			if err := repository.UpdateTrainedModelPath(ctx, modelName, modelPath); err != nil {
				log.Printf("⚠️  Failed to update database with model path: %v", err)
			} else {
				log.Printf("✅ Database updated with trained model path for model: %s", modelName)
			}
		}
	}

	log.Printf("✅ Marked training as completed: %s", trainingID)
}

// extractModelName extracts the model name from a training ID
// Training ID format: "ModelName_timestamp"
func extractModelName(trainingID string) string {
	// Split by underscore to remove timestamp
	parts := regexp.MustCompile(`_\d+$`).Split(trainingID, -1)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func markRemoteTrainingFailed(trainingID string, errorMsg string) {
	progress, err := globalTrainer.GetProgress(trainingID)
	if err != nil {
		log.Printf("⚠️  Failed to get progress for %s: %v", trainingID, err)
		return
	}

	progress.MarkFailed(errorMsg)
	log.Printf("❌ Marked training as failed: %s - %s", trainingID, errorMsg)
}

func parseMetricsFromOutput(line string) *aiAgent.TrainingMetrics {
	metrics := &aiAgent.TrainingMetrics{}

	// Pattern: Epoch 1/10, Train Loss: 0.5432
	epochPattern := regexp.MustCompile(`Epoch\s+(\d+)[/:](\d+)`)
	if matches := epochPattern.FindStringSubmatch(line); len(matches) == 3 {
		epoch, _ := strconv.Atoi(matches[1])
		total, _ := strconv.Atoi(matches[2])
		metrics.Epoch = epoch
		metrics.TotalEpochs = total
	}

	// Pattern: Train Loss: 0.5432 or loss: 0.5432
	lossPattern := regexp.MustCompile(`(?i)(train\s*)?loss[:\s]+([0-9.]+)`)
	if matches := lossPattern.FindStringSubmatch(line); len(matches) == 3 {
		loss, _ := strconv.ParseFloat(matches[2], 64)
		metrics.TrainLoss = loss
	}

	// Pattern: Val Loss: 0.4321 or validation loss: 0.4321
	valLossPattern := regexp.MustCompile(`(?i)(val|validation)\s*loss[:\s]+([0-9.]+)`)
	if matches := valLossPattern.FindStringSubmatch(line); len(matches) == 3 {
		valLoss, _ := strconv.ParseFloat(matches[2], 64)
		metrics.ValLoss = valLoss
	}

	// Pattern: Accuracy: 0.95 or Train Accuracy: 95%
	accPattern := regexp.MustCompile(`(?i)(train\s*)?acc(?:uracy)?[:\s]+([0-9.]+)%?`)
	if matches := accPattern.FindStringSubmatch(line); len(matches) == 3 {
		acc, _ := strconv.ParseFloat(matches[2], 64)
		// Convert to 0-1 range if it's a percentage
		if acc > 1 {
			acc = acc / 100
		}
		metrics.TrainAccuracy = acc
	}

	// Pattern: Val Accuracy: 0.93
	valAccPattern := regexp.MustCompile(`(?i)(val|validation)\s*acc(?:uracy)?[:\s]+([0-9.]+)%?`)
	if matches := valAccPattern.FindStringSubmatch(line); len(matches) == 3 {
		valAcc, _ := strconv.ParseFloat(matches[2], 64)
		if valAcc > 1 {
			valAcc = valAcc / 100
		}
		metrics.ValAccuracy = valAcc
	}

	// Only return metrics if we found something useful
	if metrics.Epoch > 0 || metrics.TrainLoss > 0 || metrics.TrainAccuracy > 0 {
		return metrics
	}

	return nil
}
