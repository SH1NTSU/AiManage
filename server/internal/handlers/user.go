package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"server/internal/middlewares"
	"server/internal/repository"
)

// GetCurrentUserHandler returns the current authenticated user's info
func GetCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("👤 GetCurrentUserHandler called")

	// Get user email from context
	email, ok := r.Context().Value(middlewares.UserEmailKey).(string)
	if !ok || email == "" {
		log.Println("❌ User email not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user from database
	user, err := repository.GetUserByEmail(r.Context(), email)
	if err != nil {
		log.Println("❌ Failed to get user:", err)
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}
	if user == nil {
		log.Println("❌ User not found")
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Return user info (without password)
	userInfo := map[string]interface{}{
		"id":       (*user)["id"],
		"email":    (*user)["email"],
		"username": (*user)["username"],
		"api_key":  (*user)["api_key"],
	}

	log.Printf("✅ Retrieved user info for: %s", email)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}
