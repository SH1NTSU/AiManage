package handlers

import (
	"encoding/json"
	"net/http"
	"server/helpers"
	"server/internal/models"
	"server/internal/types"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)




func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	// Get cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Couldn't get the cookie", http.StatusBadRequest)
		return
	}
			
	// Query MongoDB
	sessions, err := models.GetDocuments[types.Session]("Sessions", bson.M{
	    "refresh_token": cookie.Value,
	    "expires_at":    bson.M{"$gt": time.Now()},
	})
	if err != nil {
	    http.Error(w, "DB error", http.StatusInternalServerError)
	    return
	}
	if len(sessions) == 0 {
	    http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
	    return
	}

	session := sessions[0]

	// Generate new access token
	newAccessToken, err := helpers.GenerateJWT(session.Email)
	if err != nil {
		http.Error(w, "Couldn't generate token", http.StatusInternalServerError)
		return
	}
	
	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": newAccessToken,
	})
}
