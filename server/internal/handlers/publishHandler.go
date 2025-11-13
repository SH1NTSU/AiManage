package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"server/internal/middlewares"
	"server/internal/repository"
)

type UnPublishModelRequest struct {
	// Required fields
	ModelID     int    `json:"model_id"`
}
type PublishModelRequest struct {
	// Required fields
	ModelID     int    `json:"model_id"`
	Description string `json:"description"`
	Price       int    `json:"price"` // Price in cents (0 = free)
	LicenseType string `json:"license_type"`

	// Optional fields
	Category  string   `json:"category,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	ModelType string   `json:"model_type,omitempty"`
	Framework string   `json:"framework,omitempty"`
}

func PubHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("📤 PublishModelHandler called")

	// Parse request body
	var req PublishModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("❌ Invalid request body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ModelID == 0 {
		http.Error(w, "model_id is required", http.StatusBadRequest)
		return
	}
	if req.Description == "" {
		http.Error(w, "description is required", http.StatusBadRequest)
		return
	}
	if req.LicenseType == "" {
		http.Error(w, "license_type is required", http.StatusBadRequest)
		return
	}
	if req.Price < 0 {
		http.Error(w, "price must be non-negative", http.StatusBadRequest)
		return
	}

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

	userID, ok := (*user)["id"].(int32)
	if !ok {
		log.Println("❌ Failed to get user ID")
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	// Get model from database
	model, err := repository.GetModelByID(r.Context(), req.ModelID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Println("❌ Model not found")
			http.Error(w, "Model not found", http.StatusNotFound)
			return
		}
		log.Println("❌ Failed to get model:", err)
		http.Error(w, "Failed to get model", http.StatusInternalServerError)
		return
	}

	// Verify model belongs to the user
	modelUserID, ok := (*model)["user_id"].(int32)
	if !ok || modelUserID != userID {
		log.Println("❌ User does not own this model")
		http.Error(w, "You don't have permission to publish this model", http.StatusForbidden)
		return
	}

	// Verify model has been trained
	trainedModelPath, _ := (*model)["trained_model_path"].(string)
	if trainedModelPath == "" {
		log.Println("❌ Model has not been trained yet")
		http.Error(w, "Model must be trained before publishing", http.StatusBadRequest)
		return
	}

	// Get accuracy score from model if available
	var accuracyScore interface{} = nil
	if acc, ok := (*model)["accuracy_score"]; ok && acc != nil {
		accuracyScore = acc
	}

	// Prepare data for insertion
	publishData := map[string]interface{}{
		"model_id":           req.ModelID,
		"publisher_id":       int(userID),
		"name":               (*model)["name"],
		"picture":            (*model)["picture"],
		"trained_model_path": trainedModelPath,
		"training_script":    (*model)["training_script"],
		"description":        req.Description,
		"price":              req.Price,
		"license_type":       req.LicenseType,
		"category":           req.Category,
		"tags":               req.Tags,
		"model_type":         req.ModelType,
		"framework":          req.Framework,
		"accuracy_score":     accuracyScore,
	}

	// Insert published model
	publishedID, err := repository.InsertPublishedModel(r.Context(), publishData)
	if err != nil {
		log.Println("❌ Failed to publish model:", err)
		http.Error(w, "Failed to publish model: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Model published successfully with ID: %d", publishedID)

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Model published successfully",
		"published_id": publishedID,
	})
}

// GetPublishedModelsHandler retrieves all active published models for the community marketplace
func GetPublishedModelsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("📋 GetPublishedModelsHandler called")

	publishedModels, err := repository.GetPublishedModels(r.Context())
	if err != nil {
		log.Println("❌ Failed to get published models:", err)
		http.Error(w, "Failed to retrieve published models", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Retrieved %d published models", len(publishedModels))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(publishedModels)
}

// GetMyPublishedModelsHandler retrieves all published models by the authenticated user
func GetMyPublishedModelsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("📋 GetMyPublishedModelsHandler called")

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

	userID, ok := (*user)["id"].(int32)
	if !ok {
		log.Println("❌ Failed to get user ID")
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	publishedModels, err := repository.GetPublishedModelsByPublisher(r.Context(), int(userID))
	if err != nil {
		log.Println("❌ Failed to get published models:", err)
		http.Error(w, "Failed to retrieve published models", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Retrieved %d published models for user %d", len(publishedModels), userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(publishedModels)
}

// UnPublishModel unpublishes a published model by setting is_active to false
func UnPublishModel(w http.ResponseWriter, r *http.Request) {
	log.Println("🚫 UnPublishModel handler called")

	// Extract published model ID from URL path
	publishedModelID := r.PathValue("id")
	if publishedModelID == "" {
		log.Println("❌ Missing published model ID in URL")
		http.Error(w, "Published model ID is required", http.StatusBadRequest)
		return
	}

	// Convert string ID to integer
	var modelID int
	if _, err := fmt.Sscanf(publishedModelID, "%d", &modelID); err != nil {
		log.Printf("❌ Invalid model ID format: %s", publishedModelID)
		http.Error(w, "Invalid model ID format", http.StatusBadRequest)
		return
	}

	// Get authenticated user email from context
	email, ok := r.Context().Value(middlewares.UserEmailKey).(string)
	if !ok || email == "" {
		log.Println("❌ User email not found in request context")
		http.Error(w, "Unauthorized - authentication required", http.StatusUnauthorized)
		return
	}

	// Fetch user from database
	user, err := repository.GetUserByEmail(r.Context(), email)
	if err != nil {
		log.Printf("❌ Failed to fetch user by email %s: %v", email, err)
		http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
		return
	}
	if user == nil {
		log.Printf("❌ No user found with email: %s", email)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Extract user ID from user record
	userID, ok := (*user)["id"].(int32)
	if !ok {
		log.Println("❌ Failed to extract user ID from database record")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("📋 User %d attempting to unpublish model %d", userID, modelID)

	// Call repository to unpublish the model (includes ownership verification)
	err = repository.UnpublishModel(r.Context(), modelID, int(userID))
	if err != nil {
		log.Printf("❌ Failed to unpublish model %d: %v", modelID, err)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	log.Printf("✅ Successfully unpublished model %d by user %d", modelID, userID)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Model unpublished successfully",
		"model_id": modelID,
	})
}
