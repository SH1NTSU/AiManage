package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"server/aiAgent"
	"server/internal/middlewares"
	"server/internal/repository"
	"time"
)

// TrainingHandler handles training-related requests
type TrainingHandler struct {
	agent *aiAgent.Agent
}

// NewTrainingHandler creates a new training handler
func NewTrainingHandler(agent *aiAgent.Agent) *TrainingHandler {
	return &TrainingHandler{
		agent: agent,
	}
}

// StartTraining handles requests to start model training
func (h *TrainingHandler) StartTraining(w http.ResponseWriter, r *http.Request) {
	println("🚀 [TRAINING] Received start training request")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user email for agent check
	userEmail, ok := r.Context().Value(middlewares.UserEmailKey).(string)
	if !ok {
		println("❌ [TRAINING] No user email in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	println("👤 [TRAINING] User email:", userEmail)

	// Check if user has an agent connected (free local training)
	hasAgent := IsAgentConnected(userEmail)
	println("🔍 [TRAINING] Agent connected for", userEmail, ":", hasAgent)

	// If no agent, check if user can train on server (paid)
	if !hasAgent {
		canTrain, message := CanUserTrainOnServer(r)
		if !canTrain {
			println("❌ [TRAINING] Permission denied:", message)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   message,
				"message": "Connect your training agent or upgrade to a paid subscription",
			})
			return
		}
		println("✅ [TRAINING] User has paid subscription, training on server")
	} else {
		println("✅ [TRAINING] User has agent connected, training locally")
	}

	var req aiAgent.TrainingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		println("❌ [TRAINING] Failed to decode request:", err.Error())
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	println("📋 [TRAINING] Request details:")
	println("   - Model Name:", req.FolderName)
	println("   - Script:", req.ScriptName)
	println("   - Python:", req.PythonCommand)

	// Validate required fields
	if req.FolderName == "" {
		println("❌ [TRAINING] Missing folder_name")
		http.Error(w, "folder_name is required", http.StatusBadRequest)
		return
	}
	if req.ScriptName == "" {
		req.ScriptName = "train.py" // Default to train.py
		println("   - Using default script: train.py")
	}
	if req.PythonCommand == "" {
		req.PythonCommand = "python3" // Default to python3
		println("   - Using default Python: python3")
	}

	// Get the actual folder path from the database
	println("🔍 [TRAINING] Looking up model in database...")
	user, err := repository.GetUserByEmail(r.Context(), userEmail)
	if err != nil || user == nil {
		println("❌ [TRAINING] Failed to get user")
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	userID, ok := (*user)["id"].(int32)
	if !ok {
		println("❌ [TRAINING] Failed to get user ID")
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	models, err := repository.GetModelsByUserID(r.Context(), int(userID))
	if err != nil {
		println("❌ [TRAINING] Failed to get models:", err.Error())
		http.Error(w, "Failed to get models", http.StatusInternalServerError)
		return
	}

	// Find the model by name
	var modelFolder string
	modelName := req.FolderName // Save the original model name for training ID
	for _, model := range models {
		if name, ok := model["name"].(string); ok && name == req.FolderName {
			// Get the folder path from the model
			if folder, ok := model["folder"].([]interface{}); ok && len(folder) > 0 {
				if folderPath, ok := folder[0].(string); ok {
					modelFolder = folderPath
					println("✅ [TRAINING] Found model folder:", modelFolder)
					break
				}
			}
		}
	}

	if modelFolder == "" {
		println("❌ [TRAINING] Model not found or has no folder path")
		http.Error(w, "Model not found", http.StatusNotFound)
		return
	}

	// Update the request to use the actual folder path
	req.FolderName = modelFolder
	println("📂 [TRAINING] Using folder path:", req.FolderName)

	// Start training
	println("🔄 [TRAINING] Starting training process...")

	if hasAgent {
		// Local training: send to agent
		println("🌐 [TRAINING] Sending training request to agent...")

		// Generate training ID using model name (not folder path) so Statistics page can find it
		trainingID := fmt.Sprintf("%s_%d", modelName, time.Now().Unix())
		println("🆔 [TRAINING] Training ID:", trainingID)

		trainingData := map[string]interface{}{
			"training_id":    trainingID,
			"folder_path":    req.FolderName, // Agent expects folder_path, not folder_name
			"script_name":    req.ScriptName,
			"python_command": req.PythonCommand,
			"args":           req.Args,
			"env":            req.Env,
		}

		err := StartRemoteTraining(userEmail, trainingData)
		if err != nil {
			println("❌ [TRAINING] Failed to start remote training:", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		println("✅ [TRAINING] Training request sent to agent successfully!")
		println("🆔 [TRAINING] Training ID:", trainingID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     true,
			"message":     "Training started on your local agent",
			"remote":      true,
			"training_id": trainingID,
		})
	} else {
		// Server training: use server's trainer
		println("🖥️  [TRAINING] Starting training on server...")
		ctx := context.Background()
		trainer := h.agent.GetTrainer()
		progress, err := trainer.StartTraining(ctx, req)
		if err != nil {
			println("❌ [TRAINING] Failed to start:", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		println("✅ [TRAINING] Training started successfully on server!")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"message":  "Training started on server",
			"progress": progress,
			"remote":   false,
		})
	}
}

// GetTrainingProgress handles requests to get training progress
func (h *TrainingHandler) GetTrainingProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	trainingID := r.URL.Query().Get("id")
	if trainingID == "" {
		// Return all trainings if no ID specified
		// Only log if verbose mode or if there are active trainings
		h.getAllTrainings(w, r)
		return
	}

	trainer := h.agent.GetTrainer()
	progress, err := trainer.GetProgress(trainingID)
	if err != nil {
		println("❌ [PROGRESS] Training not found:", trainingID)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Only log significant events (not every poll)
	if progress.Status == "completed" || progress.Status == "failed" {
		println("✅ [PROGRESS] Training", trainingID, "finished with status:", progress.Status)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"progress": progress,
	})
}

// getAllTrainings returns all training jobs
func (h *TrainingHandler) getAllTrainings(w http.ResponseWriter, r *http.Request) {
	trainer := h.agent.GetTrainer()
	trainings := trainer.GetAllTrainings()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"trainings": trainings,
		"count":     len(trainings),
	})
}

// AnalyzeResults handles requests to analyze training results
func (h *TrainingHandler) AnalyzeResults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestBody struct {
		TrainingID string `json:"training_id"`
		UseAI      bool   `json:"use_ai"` // Whether to use Claude AI or quick analysis
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if requestBody.TrainingID == "" {
		http.Error(w, "training_id is required", http.StatusBadRequest)
		return
	}

	// Get training progress
	trainer := h.agent.GetTrainer()
	progress, err := trainer.GetProgress(requestBody.TrainingID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Check if training is complete
	if progress.Status != aiAgent.StatusCompleted && progress.Status != aiAgent.StatusFailed {
		http.Error(w, "Training is still in progress", http.StatusBadRequest)
		return
	}

	// Generate detailed metrics (no AI needed!)
	detailedMetrics := aiAgent.GenerateDetailedMetrics(progress)

	var analysis interface{}

	if requestBody.UseAI {
		// Use Gemini AI for detailed analysis (if available)
		aiAnalysis, err := h.agent.AnalyzeTrainingResults(progress)
		if err != nil {
			// Return detailed metrics without AI
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":  true,
				"metrics":  detailedMetrics,
				"warning":  "AI analysis not available, showing detailed metrics instead",
				"error":    err.Error(),
			})
			return
		}
		// Combine AI analysis with metrics
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"metrics":  detailedMetrics,
			"ai_analysis": aiAnalysis,
		})
		return
	} else {
		// Use detailed metrics analysis (no AI)
		analysis = detailedMetrics
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"analysis": analysis,
	})
}

// CleanupOldTrainings handles cleanup of old training records
func (h *TrainingHandler) CleanupOldTrainings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Default to cleaning up trainings older than 24 hours
	olderThan := 24 * time.Hour

	trainer := h.agent.GetTrainer()
	trainer.CleanupOldTrainings(olderThan)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Cleanup completed",
	})
}
