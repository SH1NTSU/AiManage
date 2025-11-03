package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"server/helpers"
	"server/internal/repository"
)

func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Couldn't get the cookie", http.StatusBadRequest)
		return
	}

	session, err := repository.GetSessionByRefreshToken(r.Context(), cookie.Value)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	// Get email and user_id from session
	email, ok := (*session)["email"].(string)
	if !ok {
		http.Error(w, "Invalid session data", http.StatusInternalServerError)
		return
	}

	userID, ok := (*session)["user_id"].(int32)
	if !ok {
		http.Error(w, "Invalid session data", http.StatusInternalServerError)
		return
	}

	newAccessToken, err := helpers.GenerateJWT(email, int(userID))
	if err != nil {
		http.Error(w, "Couldn't generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": newAccessToken,
	})
	log.Println("Refresh token sent successfully")
}
