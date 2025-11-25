package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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
	ReadBufferSize:  1024 * 1024, // 1MB read buffer for large training outputs
	WriteBufferSize: 1024 * 1024, // 1MB write buffer
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

// GetGlobalTrainer returns the global trainer instance
func GetGlobalTrainer() *aiAgent.Trainer {
	return globalTrainer
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

	// Set read deadline to detect dead connections
	ac.Conn.SetReadDeadline(time.Now().Add(2 * time.Minute))
	ac.Conn.SetPongHandler(func(string) error {
		ac.mu.Lock()
		ac.LastPing = time.Now()
		ac.mu.Unlock()
		ac.Conn.SetReadDeadline(time.Now().Add(2 * time.Minute))
		return nil
	})

	for {
		_, message, err := ac.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ WebSocket error: %v", err)
			} else if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Printf("✅ WebSocket closed normally: %s", ac.UserEmail)
			} else {
				log.Printf("⚠️  WebSocket read error: %v", err)
			}
			break
		}

		// Reset read deadline after successful read
		ac.Conn.SetReadDeadline(time.Now().Add(2 * time.Minute))

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
			// Legacy JSON pong message (WebSocket ping/pong frames are handled automatically via SetPongHandler)
			ac.mu.Lock()
			ac.LastPing = time.Now()
			ac.mu.Unlock()
			log.Printf("📡 JSON pong received from %s", ac.UserEmail)

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
				createRemoteTrainingProgress(trainingID, ac.UserID)
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

	// Set write deadline to prevent blocking indefinitely
	deadline := time.Now().Add(10 * time.Second)
	if err := ac.Conn.SetWriteDeadline(deadline); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	err := ac.Conn.WriteJSON(data)

	// Clear deadline after write
	ac.Conn.SetWriteDeadline(time.Time{})

	return err
}

// PingLoop sends periodic pings to keep connection alive
func (ac *AgentConnection) PingLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ac.mu.Lock()
		conn := ac.Conn
		ac.mu.Unlock()

		if conn == nil {
			return
		}

		// Use WriteControl for ping instead of JSON message (more efficient)
		deadline := time.Now().Add(5 * time.Second)
		if err := conn.SetWriteDeadline(deadline); err != nil {
			log.Printf("⚠️  Failed to set write deadline for ping: %v", err)
			return
		}

		if err := conn.WriteControl(websocket.PingMessage, []byte{}, deadline); err != nil {
			log.Printf("⚠️  Failed to send ping: %v", err)
			return
		}

		conn.SetWriteDeadline(time.Time{})

		// Check if agent is still alive (responds to pings)
		ac.mu.Lock()
		if time.Since(ac.LastPing) > 2*time.Minute {
			ac.mu.Unlock()
			log.Printf("⚠️  Agent timeout: %s (no pong received)", ac.UserEmail)
			// Send close frame before closing
			conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "timeout"), time.Now().Add(5*time.Second))
			conn.Close()
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

func createRemoteTrainingProgress(trainingID string, userID int) {
	progress := &aiAgent.TrainingProgress{
		UserID:      userID,
		Status:      aiAgent.StatusRunning,
		StartTime:   time.Now(),
		Logs:        []string{},
		Metrics:     []aiAgent.TrainingMetrics{},
		TotalEpochs: 0,
	}

	globalTrainer.StoreTrainingProgress(trainingID, progress)
	log.Printf("📊 Created remote training progress: %s for user %d", trainingID, userID)
}

func updateRemoteTrainingProgress(trainingID string, output string) {
	progress, err := globalTrainer.GetProgress(trainingID)
	if err != nil {
		log.Printf("⚠️  Failed to get progress for %s: %v", trainingID, err)
		return
	}

	// Add log
	progress.AddLog(output)

	// Try to parse PROGRESS JSON lines first (more reliable)
	if strings.HasPrefix(output, "PROGRESS:") {
		jsonStr := strings.TrimPrefix(output, "PROGRESS:")
		jsonStr = strings.TrimSpace(jsonStr)
		if metrics := parseProgressJSONFromOutput(jsonStr); metrics != nil {
			progress.AddMetrics(*metrics)
			log.Printf("📈 Parsed metrics from JSON: Epoch %d/%d, Loss: %.4f, Train Acc: %.2f%%, Test Acc: %.2f%%",
				metrics.Epoch, metrics.TotalEpochs, metrics.TrainLoss, metrics.TrainAccuracy*100, metrics.TestAccuracy*100)
			// Store final metrics if:
			// 1. Status is "completed"
			// 2. This is the last epoch
			// 3. Has any accuracy
			isCompleted := false
			if metrics.CustomMetrics != nil {
				if status, ok := metrics.CustomMetrics["status"].(string); ok && status == "completed" {
					isCompleted = true
				}
			}
			if isCompleted || metrics.TestAccuracy > 0 || metrics.ValAccuracy > 0 || metrics.TrainAccuracy > 0 ||
				(metrics.Epoch == metrics.TotalEpochs && metrics.TotalEpochs > 0) {
				progress.SetFinalMetrics(metrics)
				if isCompleted {
					log.Printf("📊 Set FinalMetrics (status=completed) with accuracy: Test=%.2f%%, Val=%.2f%%, Train=%.2f%%",
						metrics.TestAccuracy*100, metrics.ValAccuracy*100, metrics.TrainAccuracy*100)
				}
			}
			return
		}
	}

	// Try to parse metrics from output using regex patterns
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

	// Extract model name from training ID (format: "ModelName_timestamp")
	modelName := extractModelName(trainingID)
	if modelName == "" {
		log.Printf("⚠️  Could not extract model name from training ID: %s", trainingID)
		return
	}
	log.Printf("🔍 Extracted model name '%s' from training ID '%s'", modelName, trainingID)

	// Extract final accuracy from training progress
	// Note: Database expects percentage format (e.g., 95.50), but metrics are in 0-1 range
	var finalAccuracy *float64
	// Prefer FinalMetrics if available, then last metric from Metrics array
	if progress.FinalMetrics != nil {
		// Use FinalMetrics (prefer test > val > train)
		if progress.FinalMetrics.TestAccuracy > 0 {
			acc := progress.FinalMetrics.TestAccuracy * 100 // Convert 0-1 range to percentage
			finalAccuracy = &acc
			log.Printf("📊 Using FinalMetrics test accuracy: %.2f%%", acc)
		} else if progress.FinalMetrics.ValAccuracy > 0 {
			acc := progress.FinalMetrics.ValAccuracy * 100 // Convert 0-1 range to percentage
			finalAccuracy = &acc
			log.Printf("📊 Using FinalMetrics validation accuracy: %.2f%%", acc)
		} else if progress.FinalMetrics.TrainAccuracy > 0 {
			acc := progress.FinalMetrics.TrainAccuracy * 100 // Convert 0-1 range to percentage
			finalAccuracy = &acc
			log.Printf("📊 Using FinalMetrics train accuracy: %.2f%%", acc)
		}
	}
	// Fallback: search through all metrics (reverse order) to find the most recent accuracy
	if finalAccuracy == nil && len(progress.Metrics) > 0 {
		// Search from end to beginning to find the most recent metric with accuracy
		for i := len(progress.Metrics) - 1; i >= 0; i-- {
			metric := progress.Metrics[i]
			if metric.TestAccuracy > 0 {
				acc := metric.TestAccuracy * 100 // Convert 0-1 range to percentage
				finalAccuracy = &acc
				log.Printf("📊 Using metric[%d] test accuracy: %.2f%%", i, acc)
				break
			} else if metric.ValAccuracy > 0 {
				acc := metric.ValAccuracy * 100 // Convert 0-1 range to percentage
				finalAccuracy = &acc
				log.Printf("📊 Using metric[%d] validation accuracy: %.2f%%", i, acc)
				break
			} else if metric.TrainAccuracy > 0 {
				acc := metric.TrainAccuracy * 100 // Convert 0-1 range to percentage
				finalAccuracy = &acc
				log.Printf("📊 Using metric[%d] train accuracy: %.2f%%", i, acc)
				break
			}
		}
	}
	if finalAccuracy == nil {
		log.Printf("⚠️  No accuracy found in training progress")
		log.Printf("   FinalMetrics: %v", progress.FinalMetrics != nil)
		if progress.FinalMetrics != nil {
			log.Printf("   FinalMetrics.TestAccuracy: %.4f", progress.FinalMetrics.TestAccuracy)
			log.Printf("   FinalMetrics.ValAccuracy: %.4f", progress.FinalMetrics.ValAccuracy)
			log.Printf("   FinalMetrics.TrainAccuracy: %.4f", progress.FinalMetrics.TrainAccuracy)
		}
		log.Printf("   Total metrics: %d", len(progress.Metrics))
	}

	// Set model path if provided
	if modelPath != "" {
		progress.SetModelPath(modelPath)
		log.Printf("💾 Set model path: %s", modelPath)

		// Update database with trained model path and accuracy
		ctx := context.Background()
		if err := repository.UpdateTrainedModelPathAndAccuracy(ctx, modelName, modelPath, finalAccuracy); err != nil {
			log.Printf("⚠️  Failed to update database: %v", err)
		} else {
			if finalAccuracy != nil {
				log.Printf("✅ Database updated with trained model path and accuracy (%.2f%%) for model: %s", *finalAccuracy, modelName)
			} else {
				log.Printf("✅ Database updated with trained model path for model: %s", modelName)
			}
		}
	} else if finalAccuracy != nil {
		// Update accuracy even if no model path
		ctx := context.Background()
		if err := repository.UpdateModelAccuracy(ctx, modelName, *finalAccuracy); err != nil {
			log.Printf("⚠️  Failed to update accuracy: %v", err)
		} else {
			log.Printf("✅ Database updated with accuracy (%.2f%%) for model: %s", *finalAccuracy, modelName)
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

func parseProgressJSONFromOutput(jsonStr string) *aiAgent.TrainingMetrics {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil
	}

	metrics := &aiAgent.TrainingMetrics{}

	// Extract epoch
	if epoch, ok := data["epoch"].(float64); ok {
		metrics.Epoch = int(epoch)
	}
	if totalEpochs, ok := data["total_epochs"].(float64); ok {
		metrics.TotalEpochs = int(totalEpochs)
	}

	// Extract losses
	if trainLoss, ok := data["train_loss"].(float64); ok {
		metrics.TrainLoss = trainLoss
	}
	if valLoss, ok := data["val_loss"].(float64); ok {
		metrics.ValLoss = valLoss
	}
	if testLoss, ok := data["test_loss"].(float64); ok {
		metrics.ValLoss = testLoss // Use ValLoss field for test loss
	}

	// Extract accuracies (convert from percentage to 0-1 range if needed)
	if trainAcc, ok := data["train_accuracy"].(float64); ok {
		if trainAcc > 1 {
			metrics.TrainAccuracy = trainAcc / 100
		} else {
			metrics.TrainAccuracy = trainAcc
		}
	}
	if valAcc, ok := data["val_accuracy"].(float64); ok {
		if valAcc > 1 {
			metrics.ValAccuracy = valAcc / 100
		} else {
			metrics.ValAccuracy = valAcc
		}
	}
	if testAcc, ok := data["test_accuracy"].(float64); ok {
		if testAcc > 1 {
			metrics.TestAccuracy = testAcc / 100
		} else {
			metrics.TestAccuracy = testAcc
		}
	}
	// Handle generic "accuracy" field (typically used for final/test accuracy)
	if acc, ok := data["accuracy"].(float64); ok {
		// Convert from percentage to 0-1 range if needed
		if acc > 1 {
			acc = acc / 100
		}
		// Generic accuracy typically represents test/final accuracy
		// Prefer TestAccuracy, but fall back to TrainAccuracy if TestAccuracy already set from test_accuracy field
		if metrics.TestAccuracy == 0 {
			metrics.TestAccuracy = acc
		} else if metrics.TrainAccuracy == 0 {
			// If TestAccuracy is already set, use TrainAccuracy as fallback
			metrics.TrainAccuracy = acc
		} else {
			// If both are set, prefer TestAccuracy for generic accuracy (overwrite)
			metrics.TestAccuracy = acc
		}
	}

	// Extract generic "loss" field if specific loss fields are not present
	if metrics.TrainLoss == 0 {
		if loss, ok := data["loss"].(float64); ok {
			metrics.TrainLoss = loss
		}
	}

	// Check for "status" field to identify final/completed metrics
	// Store it in CustomMetrics for later use
	if status, ok := data["status"].(string); ok {
		if metrics.CustomMetrics == nil {
			metrics.CustomMetrics = make(map[string]interface{})
		}
		metrics.CustomMetrics["status"] = status
	}

	// Only return if we found useful data
	if metrics.Epoch > 0 || metrics.TrainLoss > 0 || metrics.TrainAccuracy > 0 || metrics.TestAccuracy > 0 || metrics.ValAccuracy > 0 {
		return metrics
	}

	return nil
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
