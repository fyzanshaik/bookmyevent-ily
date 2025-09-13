package user

import (
	"encoding/json"
	"net/http"

	"github.com/fyzanshaik/bookmyevent-ily/internal/auth"
	"github.com/fyzanshaik/bookmyevent-ily/internal/utils"
	"github.com/google/uuid"
)

func (cfg *APIConfig) HandleInternalVerify(w http.ResponseWriter, r *http.Request) {
	var requestBody VerifyTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if requestBody.Token == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Token is required")
		return
	}

	userID, err := auth.ValidateJWT(requestBody.Token, cfg.Config.JWTSecret)
	if err != nil {
		cfg.Logger.Debug("Token validation failed", "error", err)
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	user, err := cfg.DB.GetUserByID(r.Context(), cfg.DB_Conn, userID)
	if err != nil {
		cfg.Logger.Error("Failed to fetch user", "error", err, "user_id", userID.String())
		utils.RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	response := TokenVerificationResponse{
		UserID: user.UserID,
		Email:  user.Email,
		Valid:  true,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) HandleInternalGetUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("userId")
	if userIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	user, err := cfg.DB.GetUserByID(r.Context(), cfg.DB_Conn, userID)
	if err != nil {
		cfg.Logger.Error("Failed to fetch user", "error", err, "user_id", userID.String())
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	var phoneNumber *string
	if user.PhoneNumber.Valid {
		phoneNumber = &user.PhoneNumber.String
	}

	response := UserResponse{
		UserID:      user.UserID,
		Email:       user.Email,
		Name:        user.Name,
		PhoneNumber: phoneNumber,
		CreatedAt:   user.CreatedAt.Time,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}
