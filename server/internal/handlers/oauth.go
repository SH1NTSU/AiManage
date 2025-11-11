package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"server/helpers"
	"server/internal/repository"
)

// OAuth providers configuration
var (
	GoogleClientID     = os.Getenv("GOOGLE_CLIENT_ID")
	GoogleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	GoogleRedirectURI  = os.Getenv("GOOGLE_REDIRECT_URI")

	GithubClientID     = os.Getenv("GITHUB_CLIENT_ID")
	GithubClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")

	AppleClientID     = os.Getenv("APPLE_CLIENT_ID")
	AppleClientSecret = os.Getenv("APPLE_CLIENT_SECRET")
	AppleRedirectURI  = os.Getenv("APPLE_REDIRECT_URI")
)

// GoogleOAuthHandler handles Google OAuth callback
func GoogleOAuthHandler(w http.ResponseWriter, r *http.Request) {
	// Get the authorization code from request
	var req struct {
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Exchange code for access token
	tokenResp, err := http.PostForm("https://oauth2.googleapis.com/token", map[string][]string{
		"code":          {req.Code},
		"client_id":     {GoogleClientID},
		"client_secret": {GoogleClientSecret},
		"redirect_uri":  {GoogleRedirectURI},
		"grant_type":    {"authorization_code"},
	})

	if err != nil {
		log.Printf("Error exchanging code for token: %v", err)
		http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
		return
	}
	defer tokenResp.Body.Close()

	var tokenData struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}

	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
		log.Printf("Error decoding token response: %v", err)
		http.Error(w, "Failed to decode token", http.StatusInternalServerError)
		return
	}

	// Get user info from Google
	userResp, err := http.Get(fmt.Sprintf("https://www.googleapis.com/oauth2/v2/userinfo?access_token=%s", tokenData.AccessToken))
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer userResp.Body.Close()

	var userInfo struct {
		Email      string `json:"email"`
		Name       string `json:"name"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
	}

	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		log.Printf("Error decoding user info: %v", err)
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	// Check if user exists
	user, err := repository.GetUserByEmail(r.Context(), userInfo.Email)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	var userID int
	if user == nil {
		// Create new user with Google data
		username := strings.ToLower(strings.ReplaceAll(userInfo.Email, "@", "_"))
		if userInfo.GivenName != "" {
			username = strings.ToLower(userInfo.GivenName)
		}

		// Generate a random password (user won't use it for OAuth login)
		randomPassword, err := helpers.GenerateRandomString(32)
		if err != nil {
			http.Error(w, "Failed to generate password", http.StatusInternalServerError)
			return
		}

		userID, err = repository.InsertUser(r.Context(), userInfo.Email, randomPassword, username)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	} else {
		// Extract user ID
		switch v := (*user)["id"].(type) {
		case int:
			userID = v
		case int32:
			userID = int(v)
		case int64:
			userID = int(v)
		default:
			http.Error(w, "Invalid user data", http.StatusInternalServerError)
			return
		}
	}

	// Generate JWT token
	token, err := helpers.GenerateJWT(userInfo.Email, userID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Generate refresh token
	refreshToken, err := helpers.GenerateRandomString(64)
	if err != nil {
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	// Save session
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	_, err = repository.InsertSession(r.Context(), userID, userInfo.Email, refreshToken, expiresAt)
	if err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":         token,
		"refresh_token": refreshToken,
	})
}

// GitHubOAuthHandler handles GitHub OAuth callback
func GitHubOAuthHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Exchange code for access token
	tokenReq, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(fmt.Sprintf(
		"client_id=%s&client_secret=%s&code=%s",
		GithubClientID, GithubClientSecret, req.Code,
	)))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	tokenReq.Header.Set("Accept", "application/json")
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		log.Printf("Error exchanging code for token: %v", err)
		http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
		return
	}
	defer tokenResp.Body.Close()

	var tokenData struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
		log.Printf("Error decoding token response: %v", err)
		http.Error(w, "Failed to decode token", http.StatusInternalServerError)
		return
	}

	// Get user info from GitHub
	userReq, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		http.Error(w, "Failed to create user request", http.StatusInternalServerError)
		return
	}
	userReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenData.AccessToken))

	userResp, err := client.Do(userReq)
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer userResp.Body.Close()

	var userInfo struct {
		Login string `json:"login"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		log.Printf("Error decoding user info: %v", err)
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	// If email is not public, fetch from emails endpoint
	if userInfo.Email == "" {
		emailReq, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
		if err == nil {
			emailReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenData.AccessToken))
			emailResp, err := client.Do(emailReq)
			if err == nil {
				defer emailResp.Body.Close()
				var emails []struct {
					Email    string `json:"email"`
					Primary  bool   `json:"primary"`
					Verified bool   `json:"verified"`
				}
				if err := json.NewDecoder(emailResp.Body).Decode(&emails); err == nil {
					for _, email := range emails {
						if email.Primary && email.Verified {
							userInfo.Email = email.Email
							break
						}
					}
				}
			}
		}
	}

	if userInfo.Email == "" {
		http.Error(w, "Email not available from GitHub", http.StatusBadRequest)
		return
	}

	// Check if user exists
	user, err := repository.GetUserByEmail(r.Context(), userInfo.Email)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	var userID int
	if user == nil {
		// Create new user
		username := userInfo.Login
		if username == "" {
			username = strings.ToLower(strings.ReplaceAll(userInfo.Email, "@", "_"))
		}

		randomPassword, err := helpers.GenerateRandomString(32)
		if err != nil {
			http.Error(w, "Failed to generate password", http.StatusInternalServerError)
			return
		}

		userID, err = repository.InsertUser(r.Context(), userInfo.Email, randomPassword, username)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	} else {
		switch v := (*user)["id"].(type) {
		case int:
			userID = v
		case int32:
			userID = int(v)
		case int64:
			userID = int(v)
		default:
			http.Error(w, "Invalid user data", http.StatusInternalServerError)
			return
		}
	}

	// Generate tokens
	token, err := helpers.GenerateJWT(userInfo.Email, userID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := helpers.GenerateRandomString(64)
	if err != nil {
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	_, err = repository.InsertSession(r.Context(), userID, userInfo.Email, refreshToken, expiresAt)
	if err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":         token,
		"refresh_token": refreshToken,
	})
}

// AppleOAuthHandler handles Apple Sign In callback
func AppleOAuthHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code     string `json:"code"`
		IDToken  string `json:"id_token"`
		User     string `json:"user"` // Apple sends user info on first sign-in only
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// For Apple Sign In, we typically decode the ID token
	// This is a simplified version - in production, you'd want to verify the JWT signature
	// For now, we'll exchange the code for tokens

	tokenReq, err := http.NewRequest("POST", "https://appleid.apple.com/auth/token", strings.NewReader(fmt.Sprintf(
		"client_id=%s&client_secret=%s&code=%s&grant_type=authorization_code&redirect_uri=%s",
		AppleClientID, AppleClientSecret, req.Code, AppleRedirectURI,
	)))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		log.Printf("Error exchanging code for token: %v", err)
		http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
		return
	}
	defer tokenResp.Body.Close()

	bodyBytes, _ := io.ReadAll(tokenResp.Body)
	log.Printf("Apple token response: %s", string(bodyBytes))

	// Parse the ID token to get user email
	// In production, use a JWT library to properly verify and decode
	// For this example, we'll use the id_token from the request

	var tokenData struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
	}

	if err := json.Unmarshal(bodyBytes, &tokenData); err != nil {
		log.Printf("Error decoding token response: %v", err)
		http.Error(w, "Failed to decode token", http.StatusInternalServerError)
		return
	}

	// Decode ID token (simplified - in production use proper JWT validation)
	// For now, return an error message that Apple OAuth requires additional setup
	http.Error(w, "Apple OAuth requires additional JWT validation setup", http.StatusNotImplemented)
}
