package user

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/fyzanshaik/bookmyevent-ily/internal/auth"
	"github.com/fyzanshaik/bookmyevent-ily/internal/constants"
	"github.com/fyzanshaik/bookmyevent-ily/internal/repository/users"
	"github.com/fyzanshaik/bookmyevent-ily/internal/utils"
)

func (cfg *APIConfig) AddUser(w http.ResponseWriter, r *http.Request) {
	var requestBody CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if requestBody.Email == "" || requestBody.Password == "" || requestBody.Name == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Email, password, and name are required")
		return
	}

	hashedPassword, err := auth.HashedPassword(requestBody.Password)
	if err != nil {
		cfg.Logger.Error("Failed to hash password", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	var phoneNumber sql.NullString
	if requestBody.PhoneNumber != "" {
		phoneNumber = sql.NullString{String: requestBody.PhoneNumber, Valid: true}
	}

	params := users.CreateUserParams{
		Email:        requestBody.Email,
		PhoneNumber:  phoneNumber,
		Name:         requestBody.Name,
		PasswordHash: hashedPassword,
	}

	dbUser, err := cfg.DB.CreateUser(r.Context(), cfg.DB_Conn, params)
	if err != nil {
		cfg.Logger.Error("Failed to create user", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not create user")
		return
	}

	accessToken, err := auth.MakeJWT(dbUser.UserID, cfg.Config.JWTSecret, cfg.Config.JWTAccessDuration)
	if err != nil {
		cfg.Logger.Error("Failed to create access token", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create access token")
		return
	}

	refreshTokenString, err := auth.MakeRefreshToken()
	if err != nil {
		cfg.Logger.Error("Failed to create refresh token", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create refresh token")
		return
	}

	refreshTokenParams := users.CreateRefreshTokenParams{
		Token:     refreshTokenString,
		UserID:    dbUser.UserID,
		ExpiresAt: time.Now().UTC().Add(cfg.Config.JWTRefreshDuration),
	}

	_, err = cfg.DB.CreateRefreshToken(r.Context(), cfg.DB_Conn, refreshTokenParams)
	if err != nil {
		cfg.Logger.Error("Failed to store refresh token", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to store refresh token")
		return
	}

	response := AuthResponse{
		UserID:       dbUser.UserID,
		Email:        dbUser.Email,
		Name:         dbUser.Name,
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
	}

	utils.RespondWithJSON(w, http.StatusCreated, response)
}

func (cfg *APIConfig) LoginUser(w http.ResponseWriter, r *http.Request) {
	var requestBody UserLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if requestBody.Email == "" || requestBody.Password == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	currentUser, err := cfg.DB.GetUserByEmail(r.Context(), cfg.DB_Conn, requestBody.Email)
	if err != nil {
		cfg.Logger.Error("Failed to fetch user", "error", err)
		utils.RespondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	if err := auth.CheckPasswordHash(currentUser.PasswordHash, requestBody.Password); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	accessToken, err := auth.MakeJWT(currentUser.UserID, cfg.Config.JWTSecret, cfg.Config.JWTAccessDuration)
	if err != nil {
		cfg.Logger.Error("Failed to create access token", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create access token")
		return
	}

	refreshTokenString, err := auth.MakeRefreshToken()
	if err != nil {
		cfg.Logger.Error("Failed to create refresh token", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create refresh token")
		return
	}

	refreshTokenParams := users.CreateRefreshTokenParams{
		Token:     refreshTokenString,
		UserID:    currentUser.UserID,
		ExpiresAt: time.Now().UTC().Add(cfg.Config.JWTRefreshDuration),
	}

	_, err = cfg.DB.CreateRefreshToken(r.Context(), cfg.DB_Conn, refreshTokenParams)
	if err != nil {
		cfg.Logger.Error("Failed to store refresh token", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to store refresh token")
		return
	}

	response := AuthResponse{
		UserID:       currentUser.UserID,
		Email:        currentUser.Email,
		Name:         currentUser.Name,
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var requestBody RefreshTokenRequest

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if requestBody.RefreshToken == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	refreshTokenString := requestBody.RefreshToken

	refreshToken, err := cfg.DB.GetRefreshToken(r.Context(), cfg.DB_Conn, refreshTokenString)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	user, err := cfg.DB.GetUserByID(r.Context(), cfg.DB_Conn, refreshToken.UserID)
	if err != nil {
		cfg.Logger.Error("Failed to fetch user", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch user information")
		return
	}

	newAccessToken, err := auth.MakeJWT(user.UserID, cfg.Config.JWTSecret, cfg.Config.JWTAccessDuration)
	if err != nil {
		cfg.Logger.Error("Failed to create access token", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create access token")
		return
	}

	newRefreshTokenString, err := auth.MakeRefreshToken()
	if err != nil {
		cfg.Logger.Error("Failed to create refresh token", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create refresh token")
		return
	}

	err = cfg.DB.RevokeRefreshToken(r.Context(), cfg.DB_Conn, refreshTokenString)
	if err != nil {
		cfg.Logger.Error("Failed to revoke old refresh token", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to revoke old token")
		return
	}

	refreshTokenParams := users.CreateRefreshTokenParams{
		Token:     newRefreshTokenString,
		UserID:    user.UserID,
		ExpiresAt: time.Now().UTC().Add(cfg.Config.JWTRefreshDuration),
	}

	_, err = cfg.DB.CreateRefreshToken(r.Context(), cfg.DB_Conn, refreshTokenParams)
	if err != nil {
		cfg.Logger.Error("Failed to store new refresh token", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to store new refresh token")
		return
	}

	response := RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshTokenString,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) RevokeToken(w http.ResponseWriter, r *http.Request) {
	var requestBody LogoutRequest

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if requestBody.RefreshToken == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	refreshTokenString := requestBody.RefreshToken

	err := cfg.DB.RevokeRefreshToken(r.Context(), cfg.DB_Conn, refreshTokenString)
	if err != nil {
		cfg.Logger.Error("Failed to revoke token", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to revoke token")
		return
	}

	response := LogoutResponse{
		Message: constants.MessageSuccess,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, err := cfg.DB.GetUserByID(r.Context(), cfg.DB_Conn, userID)
	if err != nil {
		cfg.Logger.Error("Failed to fetch user", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch user information")
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

func (cfg *APIConfig) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var requestBody UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	currentUser, err := cfg.DB.GetUserByID(r.Context(), cfg.DB_Conn, userID)
	if err != nil {
		cfg.Logger.Error("Failed to get current user", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get current user")
		return
	}

	params := users.UpdateUserParams{
		UserID:      userID,
		Name:        currentUser.Name,
		PhoneNumber: currentUser.PhoneNumber,
	}

	if requestBody.Name != nil {
		params.Name = *requestBody.Name
	}

	if requestBody.PhoneNumber != nil {
		if *requestBody.PhoneNumber == "" {
			params.PhoneNumber = sql.NullString{Valid: false}
		} else {
			params.PhoneNumber = sql.NullString{String: *requestBody.PhoneNumber, Valid: true}
		}
	}

	updatedUser, err := cfg.DB.UpdateUser(r.Context(), cfg.DB_Conn, params)
	if err != nil {
		cfg.Logger.Error("Failed to update user", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	var phoneNumber *string
	if updatedUser.PhoneNumber.Valid {
		phoneNumber = &updatedUser.PhoneNumber.String
	}

	response := UserResponse{
		UserID:      updatedUser.UserID,
		Email:       updatedUser.Email,
		Name:        updatedUser.Name,
		PhoneNumber: phoneNumber,
		CreatedAt:   updatedUser.CreatedAt.Time,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) GetUserBookings(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	cfg.Logger.Info("User bookings requested", "user_id", userID.String())

	utils.RespondWithJSON(w, http.StatusOK, []any{})
}
