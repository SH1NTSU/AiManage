package handlers

import (
	"encoding/json"
	"server/internal/types"
	"log"
	"net/http"
	"server/internal/models"

	"go.mongodb.org/mongo-driver/bson"
)



func ReadHandler(w http.ResponseWriter, r *http.Request) {
    modelsData, err := models.GetDocuments[types.Model]("Models", bson.M{}) // generic
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
