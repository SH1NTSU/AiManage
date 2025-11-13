package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"server/internal/repository"
)

// UploadTrainedModelHandler handles uploading trained model files from agents
func UploadTrainedModelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("📤 [UPLOAD] Received trained model upload request")

	// Validate API key from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Println("❌ [UPLOAD] No Authorization header")
		http.Error(w, "API key required", http.StatusUnauthorized)
		return
	}

	// Extract API key (format: "Bearer <api_key>")
	apiKey := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		apiKey = authHeader[7:]
	}

	// Validate API key
	user, err := repository.GetUserByApiKey(r.Context(), apiKey)
	if err != nil || user == nil {
		log.Printf("❌ [UPLOAD] Invalid API key")
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	userEmail, _ := (*user)["email"].(string)
	log.Printf("✅ [UPLOAD] Authenticated user: %s", userEmail)

	// Parse multipart form (max 500MB for model files)
	err = r.ParseMultipartForm(500 << 20)
	if err != nil {
		log.Printf("❌ [UPLOAD] Failed to parse form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get model name
	modelName := r.FormValue("model_name")
	if modelName == "" {
		log.Println("❌ [UPLOAD] Model name is required")
		http.Error(w, "model_name is required", http.StatusBadRequest)
		return
	}

	// Get original file path (for reference)
	originalPath := r.FormValue("original_path")
	log.Printf("📋 [UPLOAD] Model: %s, Original path: %s", modelName, originalPath)

	// Get the uploaded file
	file, header, err := r.FormFile("model_file")
	if err != nil {
		log.Printf("❌ [UPLOAD] No file uploaded: %v", err)
		http.Error(w, "model_file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("📦 [UPLOAD] File: %s (%.2f MB)", header.Filename, float64(header.Size)/(1024*1024))

	// Create uploads directory for this model
	modelDir := filepath.Join("./uploads", modelName)
	if err := os.MkdirAll(modelDir, os.ModePerm); err != nil {
		log.Printf("❌ [UPLOAD] Failed to create directory: %v", err)
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		return
	}

	// Save file with original filename
	destPath := filepath.Join(modelDir, header.Filename)
	destFile, err := os.Create(destPath)
	if err != nil {
		log.Printf("❌ [UPLOAD] Failed to create file: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	// Copy file contents
	bytesWritten, err := io.Copy(destFile, file)
	if err != nil {
		log.Printf("❌ [UPLOAD] Failed to write file: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ [UPLOAD] Saved %d bytes to: %s", bytesWritten, destPath)

	// Create relative path for database (remove ./ prefix)
	relativePath := filepath.Join(modelName, header.Filename)
	log.Printf("💾 [UPLOAD] Relative path: %s", relativePath)

	// Update database with trained model path
	ctx := context.Background()
	if err := repository.UpdateTrainedModelPath(ctx, modelName, relativePath); err != nil {
		log.Printf("⚠️  [UPLOAD] Failed to update database: %v", err)
		// Don't fail the request - file is already uploaded
	} else {
		log.Printf("✅ [UPLOAD] Database updated for model: %s", modelName)
	}

	// Return success with the server path
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"success":true,"message":"Model uploaded successfully","server_path":"%s"}`, relativePath)
}
