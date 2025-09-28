package middlewares
import (
	"net/http"
	"strings"
	"context"
	"server/helpers"
)

type contextKey string

const UserEmailKey contextKey = "userEmail"

func JWTGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := helpers.ValidateJWT(tokenStr)

		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserEmailKey, claims.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
