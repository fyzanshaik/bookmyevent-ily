package auth

import (
	"context"
	"net/http"

	"github.com/fyzanshaik/bookmyevent-ily/internal/utils"
	"github.com/google/uuid"
)

type UserContextKey string

type AdminContextKey string

const (
	UserIDKey  UserContextKey  = "user_id"
	AdminIDKey AdminContextKey = "admin_id"
)

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

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			r = r.WithContext(ctx)

			next(w, r)
		}
	}
}

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

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

func RequireAdminAuth(jwtSecret string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			token, err := GetBearerToken(r.Header)
			if err != nil {
				utils.RespondWithError(w, http.StatusUnauthorized, "Missing or invalid authorization header")
				return
			}

			claims, err := ValidateAdminJWT(token, jwtSecret)
			if err != nil {
				utils.RespondWithError(w, http.StatusUnauthorized, "Invalid admin token")
				return
			}

			ctx := context.WithValue(r.Context(), AdminIDKey, claims.AdminID)
			r = r.WithContext(ctx)

			next(w, r)
		}
	}
}

func GetAdminIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	adminID, ok := ctx.Value(AdminIDKey).(uuid.UUID)
	return adminID, ok
}