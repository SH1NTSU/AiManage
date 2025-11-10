package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"server/internal/middlewares"
	"server/internal/repository"
)

// GetPublishedModelByIDHandler retrieves a single published model by ID
// Also increments the view count when accessed
func GetPublishedModelByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Get model ID from URL parameter
	modelIDStr := chi.URLParam(r, "id")
	if modelIDStr == "" {
		http.Error(w, "model ID is required", http.StatusBadRequest)
		return
	}

	modelID, err := strconv.Atoi(modelIDStr)
	if err != nil {
		http.Error(w, "Invalid model ID", http.StatusBadRequest)
		return
	}

	log.Printf("[COMMUNITY] Fetching published model ID: %d", modelID)

	// Get model from database
	model, err := repository.GetPublishedModelByID(r.Context(), modelID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Printf("[COMMUNITY] Published model %d not found", modelID)
			http.Error(w, "Model not found", http.StatusNotFound)
			return
		}
		log.Printf("[COMMUNITY ERROR] Failed to fetch model %d: %v", modelID, err)
		http.Error(w, "Failed to retrieve model", http.StatusInternalServerError)
		return
	}

	// Increment view count
	if err := repository.IncrementModelViews(r.Context(), modelID); err != nil {
		// Log error but don't fail the request
		log.Printf("[COMMUNITY WARNING] Failed to increment views for model %d: %v", modelID, err)
	}

	log.Printf("[COMMUNITY] Successfully fetched model: %s (ID: %d)", model["name"], modelID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model)
}

// DownloadPublishedModelHandler handles downloading a published model
// Requires authentication and increments download count
func DownloadPublishedModelHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (authentication required)
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int)
	if !ok {
		log.Println("[COMMUNITY ERROR] User ID not found in context")
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Get model ID from URL parameter
	modelIDStr := chi.URLParam(r, "id")
	if modelIDStr == "" {
		http.Error(w, "model ID is required", http.StatusBadRequest)
		return
	}

	modelID, err := strconv.Atoi(modelIDStr)
	if err != nil {
		http.Error(w, "Invalid model ID", http.StatusBadRequest)
		return
	}

	log.Printf("[COMMUNITY] User %d attempting to download published model %d", userID, modelID)

	// Get published model from database
	model, err := repository.GetPublishedModelByID(r.Context(), modelID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Printf("[COMMUNITY] Published model %d not found", modelID)
			http.Error(w, "Model not found", http.StatusNotFound)
			return
		}
		log.Printf("[COMMUNITY ERROR] Failed to fetch model %d: %v", modelID, err)
		http.Error(w, "Failed to retrieve model", http.StatusInternalServerError)
		return
	}

	// Check if model is active
	isActive, ok := model["is_active"].(bool)
	if !ok || !isActive {
		log.Printf("[COMMUNITY] Attempted to download inactive model %d", modelID)
		http.Error(w, "This model is not available for download", http.StatusForbidden)
		return
	}

	// Get trained model path
	trainedModelPath, ok := model["trained_model_path"].(string)
	if !ok || trainedModelPath == "" {
		log.Printf("[COMMUNITY] Model %d has no trained model path", modelID)
		http.Error(w, "No trained model file available", http.StatusNotFound)
		return
	}

	// Check if it's a paid model
	price, ok := model["price"].(int32)
	if !ok {
		price = 0
	}

	if price > 0 {
		// TODO: In the future, check if user has purchased this model
		// For now, we'll allow downloads (you can add payment logic later)
		log.Printf("[COMMUNITY] Model %d is a paid model ($%.2f), but purchase check not implemented yet", modelID, float64(price)/100.0)
	}

	// Construct full file path
	uploadsDir := os.Getenv("UPLOADS_PATH")
	if uploadsDir == "" {
		uploadsDir = "./uploads"
	}
	fullPath := filepath.Join(uploadsDir, trainedModelPath)

	// Security: ensure the path doesn't escape uploads directory
	absUploadsDir, err := filepath.Abs(uploadsDir)
	if err != nil {
		log.Printf("[COMMUNITY ERROR] Error resolving uploads directory: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		log.Printf("[COMMUNITY ERROR] Error resolving file path: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Ensure the path is within uploads directory (prevent directory traversal)
	if !filepath.HasPrefix(absFullPath, absUploadsDir) {
		log.Printf("[COMMUNITY SECURITY] Attempted path traversal: %s", trainedModelPath)
		http.Error(w, "Invalid file path", http.StatusForbidden)
		return
	}

	// Check if file exists
	fileInfo, err := os.Stat(absFullPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[COMMUNITY] Model file not found: %s", absFullPath)
			http.Error(w, "Model file not found on server", http.StatusNotFound)
			return
		}
		log.Printf("[COMMUNITY ERROR] Error accessing file: %v", err)
		http.Error(w, "Error accessing file", http.StatusInternalServerError)
		return
	}

	// Increment download count (do this before serving to ensure it's counted)
	if err := repository.IncrementModelDownloads(r.Context(), modelID); err != nil {
		// Log error but don't fail the request
		log.Printf("[COMMUNITY WARNING] Failed to increment downloads for model %d: %v", modelID, err)
	}

	// Record download in purchase/download history (optional)
	if err := repository.RecordModelDownload(r.Context(), userID, modelID); err != nil {
		// Log error but don't fail the request
		log.Printf("[COMMUNITY WARNING] Failed to record download for user %d, model %d: %v", userID, modelID, err)
	}

	// Set headers for download
	filename := filepath.Base(trainedModelPath)
	modelName, _ := model["name"].(string)
	if modelName != "" {
		// Use model name for better UX
		ext := filepath.Ext(filename)
		filename = fmt.Sprintf("%s%s", modelName, ext)
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Serve the file
	log.Printf("[COMMUNITY] Serving published model %s (ID: %d) to user %d", filename, modelID, userID)
	http.ServeFile(w, r, absFullPath)
}

// ===== LIKES =====

// LikeModelHandler handles liking a model
func LikeModelHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	modelIDStr := chi.URLParam(r, "id")
	if modelIDStr == "" {
		http.Error(w, "model ID is required", http.StatusBadRequest)
		return
	}

	modelID, err := strconv.Atoi(modelIDStr)
	if err != nil {
		http.Error(w, "Invalid model ID", http.StatusBadRequest)
		return
	}

	log.Printf("[COMMUNITY] User %d liking model %d", userID, modelID)

	if err := repository.LikeModel(r.Context(), userID, modelID); err != nil {
		log.Printf("[COMMUNITY ERROR] Failed to like model: %v", err)
		http.Error(w, "Failed to like model", http.StatusInternalServerError)
		return
	}

	// Get updated likes count
	likesCount, err := repository.GetModelLikesCount(r.Context(), modelID)
	if err != nil {
		log.Printf("[COMMUNITY ERROR] Failed to get likes count: %v", err)
		likesCount = 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Model liked successfully",
		"likes_count": likesCount,
	})
}

// UnlikeModelHandler handles unliking a model
func UnlikeModelHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	modelIDStr := chi.URLParam(r, "id")
	if modelIDStr == "" {
		http.Error(w, "model ID is required", http.StatusBadRequest)
		return
	}

	modelID, err := strconv.Atoi(modelIDStr)
	if err != nil {
		http.Error(w, "Invalid model ID", http.StatusBadRequest)
		return
	}

	log.Printf("[COMMUNITY] User %d unliking model %d", userID, modelID)

	if err := repository.UnlikeModel(r.Context(), userID, modelID); err != nil {
		log.Printf("[COMMUNITY ERROR] Failed to unlike model: %v", err)
		http.Error(w, "Failed to unlike model", http.StatusInternalServerError)
		return
	}

	// Get updated likes count
	likesCount, err := repository.GetModelLikesCount(r.Context(), modelID)
	if err != nil {
		log.Printf("[COMMUNITY ERROR] Failed to get likes count: %v", err)
		likesCount = 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Model unliked successfully",
		"likes_count": likesCount,
	})
}

// GetModelLikesHandler returns likes info for a model (count + whether current user liked it)
func GetModelLikesHandler(w http.ResponseWriter, r *http.Request) {
	modelIDStr := chi.URLParam(r, "id")
	if modelIDStr == "" {
		http.Error(w, "model ID is required", http.StatusBadRequest)
		return
	}

	modelID, err := strconv.Atoi(modelIDStr)
	if err != nil {
		http.Error(w, "Invalid model ID", http.StatusBadRequest)
		return
	}

	likesCount, err := repository.GetModelLikesCount(r.Context(), modelID)
	if err != nil {
		log.Printf("[COMMUNITY ERROR] Failed to get likes count: %v", err)
		http.Error(w, "Failed to get likes", http.StatusInternalServerError)
		return
	}

	// Check if current user liked it (optional, requires auth)
	userLiked := false
	if userID, ok := r.Context().Value(middlewares.UserIDKey).(int); ok {
		userLiked, _ = repository.HasUserLikedModel(r.Context(), userID, modelID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"likes_count": likesCount,
		"user_liked":  userLiked,
	})
}

// ===== COMMENTS =====

// GetModelCommentsHandler retrieves all comments for a model
func GetModelCommentsHandler(w http.ResponseWriter, r *http.Request) {
	modelIDStr := chi.URLParam(r, "id")
	if modelIDStr == "" {
		http.Error(w, "model ID is required", http.StatusBadRequest)
		return
	}

	modelID, err := strconv.Atoi(modelIDStr)
	if err != nil {
		http.Error(w, "Invalid model ID", http.StatusBadRequest)
		return
	}

	log.Printf("[COMMUNITY] Fetching comments for model %d", modelID)

	comments, err := repository.GetModelComments(r.Context(), modelID)
	if err != nil {
		log.Printf("[COMMUNITY ERROR] Failed to get comments: %v", err)
		http.Error(w, "Failed to retrieve comments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

// AddModelCommentHandler adds a new comment to a model
func AddModelCommentHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	modelIDStr := chi.URLParam(r, "id")
	if modelIDStr == "" {
		http.Error(w, "model ID is required", http.StatusBadRequest)
		return
	}

	modelID, err := strconv.Atoi(modelIDStr)
	if err != nil {
		http.Error(w, "Invalid model ID", http.StatusBadRequest)
		return
	}

	var req struct {
		CommentText     string `json:"comment_text"`
		ParentCommentID *int   `json:"parent_comment_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CommentText == "" {
		http.Error(w, "comment_text is required", http.StatusBadRequest)
		return
	}

	log.Printf("[COMMUNITY] User %d adding comment to model %d", userID, modelID)

	commentID, err := repository.AddComment(r.Context(), userID, modelID, req.CommentText, req.ParentCommentID)
	if err != nil {
		log.Printf("[COMMUNITY ERROR] Failed to add comment: %v", err)
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Comment added successfully",
		"comment_id": commentID,
	})
}

// DeleteModelCommentHandler deletes a comment (only by comment author)
func DeleteModelCommentHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	commentIDStr := chi.URLParam(r, "commentId")
	if commentIDStr == "" {
		http.Error(w, "comment ID is required", http.StatusBadRequest)
		return
	}

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	log.Printf("[COMMUNITY] User %d deleting comment %d", userID, commentID)

	if err := repository.DeleteComment(r.Context(), commentID, userID); err != nil {
		log.Printf("[COMMUNITY ERROR] Failed to delete comment: %v", err)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Comment deleted successfully",
	})
}
