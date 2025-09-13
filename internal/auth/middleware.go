package auth

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/fyzanshaik/bookmyevent-ily/internal/utils"
)

// UserContextKey is the key used to store user ID in context
type UserContextKey string

const (
	UserIDKey UserContextKey = "user_id"
)

// RequireAuth middleware that requires valid JWT authentication
func RequireAuth(jwtSecret string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			token, err := GetBearerToken(r.Header)
			if err != nil {
				utils.RespondWithError(w, http.StatusUnauthorized, "Missing or invalid authorization header")
				return
			}

			userID, err := ValidateJWT(token, jwtSecret)
			if err != nil {
				utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			r = r.WithContext(ctx)

			next(w, r)
		}
	}
}

// RequireInternalAuth middleware for inter-service communication
func RequireInternalAuth(apiKey string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			receivedKey, err := GetAPIKey(r.Header)
			if err != nil {
				utils.RespondWithError(w, http.StatusUnauthorized, "Missing or invalid API key")
				return
			}

			if receivedKey != apiKey {
				utils.RespondWithError(w, http.StatusForbidden, "Invalid API key")
				return
			}

			next(w, r)
		}
	}
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}