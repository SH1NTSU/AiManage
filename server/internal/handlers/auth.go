package handlers

import (
	"encoding/json"
	"net/http"
	"server/helpers"
	"server/internal/models"
	"server/internal/types"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)




const collectionName = "User"

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	const collectionName = "Users"

	var rq types.User
	if err := json.NewDecoder(r.Body).Decode(&rq); err != nil {
		http.Error(w, "Couldn't decode request", http.StatusBadRequest)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(rq.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Couldn't hash password", http.StatusInternalServerError)
		return
	}

	u := types.User{
		Email:    rq.Email,
		Password: string(hashed),
	}

	// Optional: Check if user already exists
	existing, err := models.GetDocuments[types.User](collectionName, bson.M{"email": rq.Email})
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	if len(existing) > 0 {
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	if err := models.Insert(collectionName, u); err != nil {
		http.Error(w, "Couldn't insert user into DB", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered"})
}



func LoginHandler(w http.ResponseWriter, r *http.Request) {
	const collectionName = "Users"

	var rq types.User
	if err := json.NewDecoder(r.Body).Decode(&rq); err != nil {
		http.Error(w, "Couldn't decode request", http.StatusBadRequest)
		return
	}

	// Fetch user by email
	users, err := models.GetDocuments[types.User](collectionName, bson.M{"email": rq.Email})
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	if len(users) == 0 {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	user := users[0]

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(rq.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := helpers.GenerateJWT(user.Email)
	if err != nil {
		http.Error(w, "Couldn't generate token", http.StatusInternalServerError)
		return
	}

	// Generate refresh token
	refreshToken, err := helpers.GenerateRandomString(64)
	if err != nil {
		http.Error(w, "Couldn't generate refresh token", http.StatusInternalServerError)
		return
	}

	// Create session struct
	session := types.Session{
		Email:        user.Email,
		Refresh_token: refreshToken,
		Expires_at: time.Now().Add(30 * 24 * time.Hour),
	}

	// Save session to DB
	if err := models.Insert("Sessions", session); err != nil {
		http.Error(w, "Couldn't save session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   30 * 24 * 60 * 60, 
	})
	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"token":         token,
		"refresh_token": refreshToken,
	})
}
