package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"server/helpers"
	"server/internal/email"
	"server/internal/repository"
	"golang.org/x/crypto/bcrypt"
)




func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var rq struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
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

	// Generate verification token
	token, err := helpers.GenerateRandomString(32)
	if err != nil {
		log.Printf("[REGISTER ERROR] Failed to generate verification token: %v", err)
		http.Error(w, "Failed to generate verification token", http.StatusInternalServerError)
		return
	}

	// Set token expiry (24 hours)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Save token to database
	err = repository.SetVerificationToken(r.Context(), rq.Email, token, expiresAt)
	if err != nil {
		log.Printf("[REGISTER ERROR] Failed to save verification token: %v", err)
		// Continue without verification - user can request resend
	}

	// Send verification email (non-blocking)
	emailService := email.NewEmailService()
	go func() {
		err := emailService.SendVerificationEmail(rq.Email, rq.Username, token)
		if err != nil {
			log.Printf("[REGISTER ERROR] Failed to send verification email: %v", err)
		}
	}()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered. Please check your email to verify your account.",
	})
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

	log.Printf("[LOGIN] Attempting login for email: %s", rq.Email)

	// Fetch user by email
	user, err := repository.GetUserByEmail(r.Context(), rq.Email)
	if err != nil {
		log.Printf("[LOGIN ERROR] DB error fetching user: %v", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		log.Printf("[LOGIN ERROR] User not found for email: %s", rq.Email)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	log.Printf("[LOGIN] User found: %+v", *user)

	// Check if email is verified
	emailVerified, ok := (*user)["email_verified"].(bool)
	if !ok {
		log.Printf("[LOGIN ERROR] email_verified field type assertion failed")
		emailVerified = false
	}

	if !emailVerified {
		log.Printf("[LOGIN ERROR] Email not verified for: %s", rq.Email)
		http.Error(w, "Email not verified. Please check your email for verification link.", http.StatusUnauthorized)
		return
	}

	// Get password from user map
	passwordHash, ok := (*user)["password"].(string)
	if !ok {
		log.Printf("[LOGIN ERROR] Password field type assertion failed. User data: %+v", *user)
		http.Error(w, "Invalid user data", http.StatusInternalServerError)
		return
	}

	log.Printf("[LOGIN] Password hash retrieved, length: %d", len(passwordHash))

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(rq.Password)); err != nil {
		log.Printf("[LOGIN ERROR] Password comparison failed for email: %s, error: %v", rq.Email, err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	log.Printf("[LOGIN] Password verified successfully for email: %s", rq.Email)

	// Get user ID - handle multiple integer types from PostgreSQL
	var userID int
	switch v := (*user)["id"].(type) {
	case int:
		userID = v
		log.Printf("[LOGIN] User ID extracted as int: %d", userID)
	case int32:
		userID = int(v)
		log.Printf("[LOGIN] User ID extracted as int32, converted to int: %d", userID)
	case int64:
		userID = int(v)
		log.Printf("[LOGIN] User ID extracted as int64, converted to int: %d", userID)
	default:
		log.Printf("[LOGIN ERROR] User ID type assertion failed. Type: %T, Value: %v", (*user)["id"], (*user)["id"])
		http.Error(w, "Invalid user data", http.StatusInternalServerError)
		return
	}

	// Generate JWT token with email and userID
	log.Printf("[LOGIN] Generating JWT for userID: %d, email: %s", userID, rq.Email)
	token, err := helpers.GenerateJWT(rq.Email, userID)
	if err != nil {
		log.Printf("[LOGIN ERROR] JWT generation failed: %v", err)
		http.Error(w, "Couldn't generate token", http.StatusInternalServerError)
		return
	}

	log.Printf("[LOGIN] JWT generated successfully")

	// Generate refresh token
	refreshToken, err := helpers.GenerateRandomString(64)
	if err != nil {
		log.Printf("[LOGIN ERROR] Refresh token generation failed: %v", err)
		http.Error(w, "Couldn't generate refresh token", http.StatusInternalServerError)
		return
	}

	log.Printf("[LOGIN] Refresh token generated")

	// Save session to DB
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	sessionID, err := repository.InsertSession(r.Context(), userID, rq.Email, refreshToken, expiresAt)
	if err != nil {
		log.Printf("[LOGIN ERROR] Session save failed: %v", err)
		http.Error(w, "Couldn't save session", http.StatusInternalServerError)
		return
	}

	log.Printf("[LOGIN] Session saved with ID: %d", sessionID)

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

	log.Printf("[LOGIN] Login successful for email: %s, userID: %d", rq.Email, userID)
}

// VerifyEmailHandler handles email verification via token
func VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Verification token is required", http.StatusBadRequest)
		return
	}

	log.Printf("[EMAIL VERIFICATION] Attempting to verify email with token")

	// Verify the email using the token
	user, err := repository.VerifyEmailByToken(r.Context(), token)
	if err != nil {
		log.Printf("[EMAIL VERIFICATION ERROR] %v", err)
		http.Error(w, "Invalid or expired verification token", http.StatusBadRequest)
		return
	}

	log.Printf("[EMAIL VERIFICATION] Email verified successfully for user: %v", (*user)["email"])

	// Send welcome email (optional, non-blocking)
	userEmail := (*user)["email"].(string)
	username := (*user)["username"].(string)
	emailService := email.NewEmailService()
	go emailService.SendWelcomeEmail(userEmail, username)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Email verified successfully",
		"email":   userEmail,
	})
}

// ResendVerificationEmailHandler resends the verification email
func ResendVerificationEmailHandler(w http.ResponseWriter, r *http.Request) {
	var rq struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&rq); err != nil {
		http.Error(w, "Couldn't decode request", http.StatusBadRequest)
		return
	}

	if rq.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	log.Printf("[RESEND VERIFICATION] Resending verification email to: %s", rq.Email)

	// Check if user exists
	user, err := repository.GetUserByEmail(r.Context(), rq.Email)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		// Don't reveal whether email exists
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "If the email exists, a verification link has been sent"})
		return
	}

	// Check if already verified
	emailVerified, ok := (*user)["email_verified"].(bool)
	if ok && emailVerified {
		http.Error(w, "Email is already verified", http.StatusBadRequest)
		return
	}

	// Generate new verification token
	token, err := helpers.GenerateRandomString(32)
	if err != nil {
		log.Printf("[RESEND VERIFICATION ERROR] Failed to generate token: %v", err)
		http.Error(w, "Failed to generate verification token", http.StatusInternalServerError)
		return
	}

	// Set token expiry (24 hours)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Save token to database
	err = repository.SetVerificationToken(r.Context(), rq.Email, token, expiresAt)
	if err != nil {
		log.Printf("[RESEND VERIFICATION ERROR] Failed to save token: %v", err)
		http.Error(w, "Failed to save verification token", http.StatusInternalServerError)
		return
	}

	// Send verification email
	username, _ := (*user)["username"].(string)
	if username == "" {
		username = rq.Email
	}

	emailService := email.NewEmailService()
	err = emailService.SendVerificationEmail(rq.Email, username, token)
	if err != nil {
		log.Printf("[RESEND VERIFICATION ERROR] Failed to send email: %v", err)
		http.Error(w, "Failed to send verification email", http.StatusInternalServerError)
		return
	}

	log.Printf("[RESEND VERIFICATION] Verification email sent to: %s", rq.Email)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Verification email sent"})
}
