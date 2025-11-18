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

	// Ensure user has an API key (generate if missing)
	userID, ok := (*user)["id"].(int32)
	if !ok {
		// Try other integer types
		switch v := (*user)["id"].(type) {
		case int:
			userID = int32(v)
		case int64:
			userID = int32(v)
		default:
			log.Println("❌ Invalid user ID type")
			http.Error(w, "Invalid user ID", http.StatusInternalServerError)
			return
		}
	}

	apiKey, ok := (*user)["api_key"].(string)
	if !ok || apiKey == "" {
		// Generate API key if missing
		log.Printf("⚠️  User %s doesn't have an API key, generating one...", email)
		newKey, err := repository.EnsureUserHasAPIKey(r.Context(), int(userID))
		if err != nil {
			log.Printf("❌ Failed to generate API key: %v", err)
			// Continue with empty key rather than failing the request
			apiKey = ""
		} else {
			apiKey = newKey
			log.Printf("✅ Generated API key for user: %s", email)
		}
	}

	// Return user info (without password)
	userInfo := map[string]interface{}{
		"id":       (*user)["id"],
		"email":    (*user)["email"],
		"username": (*user)["username"],
		"api_key":  apiKey,
	}

	log.Printf("✅ Retrieved user info for: %s", email)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

// RegenerateAPIKeyHandler handles API key regeneration requests
func RegenerateAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("🔑 RegenerateAPIKeyHandler called")

	// Get authenticated user
	email, ok := r.Context().Value(middlewares.UserEmailKey).(string)
	if !ok || email == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user from database
	user, err := repository.GetUserByEmail(r.Context(), email)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	userID, ok := (*user)["id"].(int32)
	if !ok {
		// Try other integer types
		switch v := (*user)["id"].(type) {
		case int:
			userID = int32(v)
		case int64:
			userID = int32(v)
		default:
			log.Println("❌ Invalid user ID type")
			http.Error(w, "Invalid user ID", http.StatusInternalServerError)
			return
		}
	}

	// Regenerate API key
	newAPIKey, err := repository.RegenerateAPIKey(r.Context(), int(userID))
	if err != nil {
		log.Printf("❌ Failed to regenerate API key: %v", err)
		http.Error(w, "Failed to regenerate API key", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Regenerated API key for user: %s", email)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"api_key": newAPIKey,
		"message": "API key regenerated successfully",
	})
}
