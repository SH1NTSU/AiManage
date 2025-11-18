package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)



var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type Claims struct {
	Email  string `json:"email"`
	UserID string `json:"userID"`
	jwt.RegisteredClaims
}

func GenerateJWT(email string, userID int) (string, error) {
	claims := Claims{
		Email:  email,
		UserID: strconv.Itoa(userID),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // valid for 24h
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateJWT(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")

}

func GenerateRandomString(n int) (string , error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return base64.URLEncoding.EncodeToString(b), err
}

// GenerateAPIKey generates a new API key in the format: sk_live_<random_string>
// The format matches the database migration pattern: sk_live_ + 24 random characters
func GenerateAPIKey(email string) (string, error) {
	// Generate random bytes
	randomBytes := make([]byte, 18) // 18 bytes = 24 base64 chars (after encoding)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	
	// Encode to base64 URL-safe string and take first 24 characters
	// This matches the SQL pattern: substr(md5(random()::text || email), 1, 24)
	randomStr := base64.URLEncoding.EncodeToString(randomBytes)
	if len(randomStr) > 24 {
		randomStr = randomStr[:24]
	}
	
	// Ensure we have exactly 24 characters (pad if needed, though unlikely)
	for len(randomStr) < 24 {
		extraBytes := make([]byte, 1)
		rand.Read(extraBytes)
		randomStr += base64.URLEncoding.EncodeToString(extraBytes)[:1]
		if len(randomStr) > 24 {
			randomStr = randomStr[:24]
		}
	}
	
	return "sk_live_" + randomStr, nil
}
