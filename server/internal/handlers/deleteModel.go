package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"server/internal/middlewares"
	"server/internal/repository"
)

func DeleteModel(w http.ResponseWriter, r *http.Request) {
	// 1. Get userID from context (set by JWT middleware)
	//    This is WHO is making the request
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int)
	if !ok {
		log.Println("❌ User ID not found in context")
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}

	// 2. Get modelID from request body
	//    This is WHAT they want to delete
	var req struct {
		ModelID int `json:"model_id"`
		Name string `json:"name"`

	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("❌ Failed to decode request:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ModelID == 0 {
		http.Error(w, "model_id is required", http.StatusBadRequest)
		return
	}

	log.Printf("🗑️  User %d deleting model %d", userID, req.ModelID)

	// 3. Call repository with context from request
	//    r.Context() is the ctx you were missing!
	deletedID, err := repository.DeleteModel(r.Context(), req.ModelID, userID)
	if err != nil {
		log.Println("❌ Delete failed:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	modelDir := "./uploads/" + req.Name
	if err := os.RemoveAll(modelDir); err != nil {
		log.Println("❌ Failed to create model directory:", err)
		http.Error(w, "Could not create model directory: "+err.Error(), http.StatusInternalServerError)
		return
	}


	log.Printf("✅ Deleted model ID: %d", deletedID)

	// 4. Send success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Model deleted successfully",
		"deleted_id": deletedID,
	})
}
