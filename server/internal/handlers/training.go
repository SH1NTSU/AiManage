package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"server/aiAgent"
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

	// Check if user has permission to train on server
	canTrain, message := CanUserTrainOnServer(r)
	if !canTrain {
		println("❌ [TRAINING] Permission denied:", message)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   message,
			"message": "Consider training locally or upgrading your subscription",
		})
		return
	}

	var req aiAgent.TrainingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		println("❌ [TRAINING] Failed to decode request:", err.Error())
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	println("📋 [TRAINING] Request details:")
	println("   - Folder:", req.FolderName)
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

	// Start training
	println("🔄 [TRAINING] Starting training process...")
	ctx := context.Background()
	trainer := h.agent.GetTrainer()
	progress, err := trainer.StartTraining(ctx, req)
	if err != nil {
		println("❌ [TRAINING] Failed to start:", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	println("✅ [TRAINING] Training started successfully!")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Training started successfully",
		"progress": progress,
	})
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
