package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"server/internal/models"

	"go.mongodb.org/mongo-driver/bson"
)



func ReadHandler(w http.ResponseWriter, r *http.Request) {
	modelsData, err := models.GetModels(bson.M{}) // pass empty filter to get all
	if err != nil {
		log.Println("problem with getting response from db function", err)
		http.Error(w, "failed to fetch models", http.StatusInternalServerError)
		return
	}
	
	// send as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(modelsData); err != nil {
		log.Println("error encoding response:", err)
	}
}

