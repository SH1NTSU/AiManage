package handlers

import (
	"encoding/json"
	"log"
	"net/http"

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
