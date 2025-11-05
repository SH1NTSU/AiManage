package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"server/internal/middlewares"
	"server/internal/repository"
)

func ReadHandler(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value(middlewares.UserIDKey).(int)
	if !ok {
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}

	modelsData, err := repository.GetModelsByUserID(r.Context(), userID)
	if err != nil {
		log.Println("problem with getting response from db function", err)
		http.Error(w, "failed to fetch models", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(modelsData); err != nil {
		log.Println("error encoding response:", err)
	}
}

// DownloadTrainedModelHandler serves the trained model file for download
func DownloadTrainedModelHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context for security
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int)
	if !ok {
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}

	// Get model ID from query parameter
	modelIDStr := r.URL.Query().Get("model_id")
	if modelIDStr == "" {
		http.Error(w, "model_id parameter is required", http.StatusBadRequest)
		return
	}

	modelID, err := strconv.Atoi(modelIDStr)
	if err != nil {
		http.Error(w, "Invalid model_id", http.StatusBadRequest)
		return
	}

	// Get model from database
	model, err := repository.QueryRow(r.Context(), "SELECT id, user_id, name, trained_model_path FROM models WHERE id = $1", modelID)
	if err != nil {
		log.Printf("Error fetching model %d: %v", modelID, err)
		http.Error(w, "Model not found", http.StatusNotFound)
		return
	}

	// Security check: ensure the model belongs to this user
	modelUserID, ok := model["user_id"].(int32)
	if !ok || int(modelUserID) != userID {
		log.Printf("Security: User %d attempted to download model %d owned by user %d", userID, modelID, modelUserID)
		http.Error(w, "You don't have permission to download this model", http.StatusForbidden)
		return
	}

	// Check if trained model exists
	trainedModelPath, ok := model["trained_model_path"].(string)
	if !ok || trainedModelPath == "" {
		http.Error(w, "This model hasn't been trained yet", http.StatusNotFound)
		return
	}

	// Construct full file path (assuming uploads directory)
	uploadsDir := os.Getenv("UPLOADS_PATH")
	if uploadsDir == "" {
		uploadsDir = "./uploads"
	}
	fullPath := filepath.Join(uploadsDir, trainedModelPath)

	// Security: ensure the path doesn't escape uploads directory
	absUploadsDir, err := filepath.Abs(uploadsDir)
	if err != nil {
		log.Printf("Error resolving uploads directory: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		log.Printf("Error resolving file path: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Ensure the path is within uploads directory (prevent directory traversal)
	if !filepath.HasPrefix(absFullPath, absUploadsDir) {
		log.Printf("Security: Attempted path traversal: %s", trainedModelPath)
		http.Error(w, "Invalid file path", http.StatusForbidden)
		return
	}

	// Check if file exists
	fileInfo, err := os.Stat(absFullPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Trained model file not found: %s", absFullPath)
			http.Error(w, "Trained model file not found", http.StatusNotFound)
			return
		}
		log.Printf("Error accessing file: %v", err)
		http.Error(w, "Error accessing file", http.StatusInternalServerError)
		return
	}

	// Set headers for download
	filename := filepath.Base(trainedModelPath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Serve the file
	log.Printf("Serving trained model %s to user %d", filename, userID)
	http.ServeFile(w, r, absFullPath)
}
