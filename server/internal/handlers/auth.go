package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"server/helpers"
	"server/internal/repository"
	"golang.org/x/crypto/bcrypt"
)




func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var rq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&rq); err != nil {
		http.Error(w, "Couldn't decode request", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if rq.Email == "" || rq.Password == "" || rq.Username == "" {
		http.Error(w, "Email, password, and username are required", http.StatusBadRequest)
		return
	}

	// Check if email already exists
	existing, err := repository.GetUserByEmail(r.Context(), rq.Email)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	if existing != nil {
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	// Check if username already exists
	existingUsername, err := repository.GetUserByUsername(r.Context(), rq.Username)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	if existingUsername != nil {
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(rq.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Couldn't hash password", http.StatusInternalServerError)
		return
	}

	// Insert user
	_, err = repository.InsertUser(r.Context(), rq.Email, string(hashed), rq.Username)
	if err != nil {
		http.Error(w, "Couldn't insert user into DB", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered"})
}



func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var rq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&rq); err != nil {
		http.Error(w, "Couldn't decode request", http.StatusBadRequest)
		return
	}

	// Fetch user by email
	user, err := repository.GetUserByEmail(r.Context(), rq.Email)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Get password from user map
	passwordHash, ok := (*user)["password"].(string)
	if !ok {
		http.Error(w, "Invalid user data", http.StatusInternalServerError)
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(rq.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Get user ID
	userID, ok := (*user)["id"].(int32)
	if !ok {
		http.Error(w, "Invalid user data", http.StatusInternalServerError)
		return
	}

	// Generate JWT token with email and userID
	token, err := helpers.GenerateJWT(rq.Email, int(userID))
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

	// Save session to DB
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	_, err = repository.InsertSession(r.Context(), int(userID), rq.Email, refreshToken, expiresAt)
	if err != nil {
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
