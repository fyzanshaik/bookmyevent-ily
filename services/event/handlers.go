package event

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fyzanshaik/bookmyevent-ily/internal/auth"
	"github.com/fyzanshaik/bookmyevent-ily/internal/repository/events"
	"github.com/fyzanshaik/bookmyevent-ily/internal/utils"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

func (cfg *APIConfig) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	if err := cfg.DB_Conn.PingContext(r.Context()); err != nil {
		cfg.Logger.Error("Database health check failed", "error", err)
		utils.RespondWithError(w, http.StatusServiceUnavailable, "Database not ready")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"status":   "ready",
		"database": "connected",
		"service":  "event-service",
	})
}

func (cfg *APIConfig) AdminRegister(w http.ResponseWriter, r *http.Request) {
	var requestBody AdminRegisterRequest

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

	role := requestBody.Role
	if role == "" {
		role = "event_manager"
	}

	var phoneNumber sql.NullString
	if requestBody.PhoneNumber != "" {
		phoneNumber = sql.NullString{String: requestBody.PhoneNumber, Valid: true}
	}

	params := events.CreateAdminParams{
		Email:        requestBody.Email,
		Name:         requestBody.Name,
		PhoneNumber:  phoneNumber,
		PasswordHash: hashedPassword,
		Role:         sql.NullString{String: role, Valid: true},
		Permissions:  pqtype.NullRawMessage{RawMessage: json.RawMessage("{}"), Valid: true},
	}

	admin, err := cfg.DB.CreateAdmin(r.Context(), params)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			utils.RespondWithError(w, http.StatusConflict, "Email already exists")
			return
		}
		cfg.Logger.Error("Failed to create admin", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not create admin")
		return
	}

	jwtSecret := cfg.Config.JWTSecret
	accessToken, err := auth.MakeAdminJWT(admin.AdminID, role, "{}", jwtSecret, cfg.Config.JWTAccessDuration)
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

	refreshTokenParams := events.CreateAdminRefreshTokenParams{
		Token:     refreshTokenString,
		AdminID:   admin.AdminID,
		ExpiresAt: time.Now().UTC().Add(cfg.Config.JWTRefreshDuration),
	}

	_, err = cfg.DB.CreateAdminRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		cfg.Logger.Error("Failed to store admin refresh token", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to store refresh token")
		return
	}

	response := AdminAuthResponse{
		AdminID:      admin.AdminID,
		Email:        admin.Email,
		Name:         admin.Name,
		Role:         role,
		Permissions:  "{}",
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
	}

	utils.RespondWithJSON(w, http.StatusCreated, response)
}

func (cfg *APIConfig) AdminLogin(w http.ResponseWriter, r *http.Request) {
	var requestBody AdminLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if requestBody.Email == "" || requestBody.Password == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	admin, err := cfg.DB.GetAdminByEmail(r.Context(), requestBody.Email)
	if err != nil {
		cfg.Logger.Error("Failed to fetch admin", "error", err)
		utils.RespondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	if err := auth.CheckPasswordHash(admin.PasswordHash, requestBody.Password); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	role := admin.Role.String
	permissions := "{}"
	if admin.Permissions.Valid {
		permissions = string(admin.Permissions.RawMessage)
	}

	jwtSecret := cfg.Config.JWTSecret
	accessToken, err := auth.MakeAdminJWT(admin.AdminID, role, permissions, jwtSecret, cfg.Config.JWTAccessDuration)
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

	refreshTokenParams := events.CreateAdminRefreshTokenParams{
		Token:     refreshTokenString,
		AdminID:   admin.AdminID,
		ExpiresAt: time.Now().UTC().Add(cfg.Config.JWTRefreshDuration),
	}

	_, err = cfg.DB.CreateAdminRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		cfg.Logger.Error("Failed to store admin refresh token", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to store refresh token")
		return
	}

	response := AdminAuthResponse{
		AdminID:      admin.AdminID,
		Email:        admin.Email,
		Name:         admin.Name,
		Role:         role,
		Permissions:  permissions,
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) AdminRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing or invalid authorization header")
		return
	}

	refreshToken, err := cfg.DB.GetAdminRefreshToken(r.Context(), refreshTokenString)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	admin, err := cfg.DB.GetAdminByID(r.Context(), refreshToken.AdminID)
	if err != nil {
		cfg.Logger.Error("Failed to fetch admin", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch admin information")
		return
	}

	role := admin.Role.String
	permissions := "{}"
	if admin.Permissions.Valid {
		permissions = string(admin.Permissions.RawMessage)
	}

	newAccessToken, err := auth.MakeAdminJWT(admin.AdminID, role, permissions, cfg.Config.JWTSecret, cfg.Config.JWTAccessDuration)
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

	err = cfg.DB.RevokeAdminRefreshToken(r.Context(), refreshTokenString)
	if err != nil {
		cfg.Logger.Error("Failed to revoke old admin refresh token", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to revoke old token")
		return
	}

	refreshTokenParams := events.CreateAdminRefreshTokenParams{
		Token:     newRefreshTokenString,
		AdminID:   admin.AdminID,
		ExpiresAt: time.Now().UTC().Add(cfg.Config.JWTRefreshDuration),
	}

	_, err = cfg.DB.CreateAdminRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		cfg.Logger.Error("Failed to store new admin refresh token", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to store new refresh token")
		return
	}

	response := AdminRefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshTokenString,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) AdminLogout(w http.ResponseWriter, r *http.Request) {
	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing or invalid authorization header")
		return
	}

	err = cfg.DB.RevokeAdminRefreshToken(r.Context(), refreshTokenString)
	if err != nil {
		cfg.Logger.Error("Failed to revoke admin refresh token", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to revoke token")
		return
	}

	response := map[string]string{
		"message": "Admin logged out successfully",
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) ListPublishedEvents(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	eventType := r.URL.Query().Get("type")
	city := r.URL.Query().Get("city")
	dateFromStr := r.URL.Query().Get("date_from")
	dateToStr := r.URL.Query().Get("date_to")

	// Use zero time as "no limit" values - the SQL query will handle these properly
	var dateFrom, dateTo time.Time

	if dateFromStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			dateFrom = parsed
		}
	}
	if dateToStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateToStr); err == nil {
			dateTo = parsed
		}
	}

	offset := (page - 1) * limit

	countParams := events.CountPublishedEventsParams{
		Column1: eventType,
		Column2: city,
		Column3: dateFrom,
		Column4: dateTo,
	}

	total, err := cfg.DB.CountPublishedEvents(r.Context(), countParams)
	if err != nil {
		cfg.Logger.Error("Failed to count published events", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch events")
		return
	}

	listParams := events.ListPublishedEventsParams{
		Limit:   int32(limit),
		Offset:  int32(offset),
		Column3: eventType,
		Column4: city,
		Column5: dateFrom,
		Column6: dateTo,
	}

	eventsList, err := cfg.DB.ListPublishedEvents(r.Context(), listParams)
	if err != nil {
		cfg.Logger.Error("Failed to fetch published events", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch events")
		return
	}

	eventResponses := make([]EventResponse, len(eventsList))
	for i, event := range eventsList {
		eventResponses[i] = EventResponse{
			EventID:              event.EventID,
			Name:                 event.Name,
			Description:          utils.StringPtrFromNullString(event.Description),
			VenueID:              event.VenueID,
			VenueName:            &event.VenueName,
			VenueCity:            &event.City,
			VenueState:           utils.StringPtrFromNullString(event.State),
			EventType:            event.EventType,
			StartDatetime:        event.StartDatetime,
			EndDatetime:          event.EndDatetime,
			TotalCapacity:        event.TotalCapacity,
			AvailableSeats:       event.AvailableSeats,
			BasePrice:            utils.ParsePrice(sql.NullString{String: event.BasePrice, Valid: event.BasePrice != ""}),
			MaxTicketsPerBooking: event.MaxTicketsPerBooking.Int32,
			Status:               event.Status.String,
			CreatedAt:            event.CreatedAt.Time,
		}
	}

	response := EventListResponse{
		Events:  eventResponses,
		Total:   total,
		Page:    page,
		Limit:   limit,
		HasMore: int64((page-1)*limit+len(eventResponses)) < total,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) GetEventByID(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.PathValue("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	event, err := cfg.DB.GetEventByID(r.Context(), eventID)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondWithError(w, http.StatusNotFound, "Event not found")
			return
		}
		cfg.Logger.Error("Failed to fetch event", "error", err, "event_id", eventID)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch event")
		return
	}

	response := EventResponse{
		EventID:              event.EventID,
		Name:                 event.Name,
		Description:          utils.StringPtrFromNullString(event.Description),
		VenueID:              event.VenueID,
		VenueName:            &event.VenueName,
		VenueAddress:         &event.Address,
		VenueCity:            &event.City,
		VenueState:           utils.StringPtrFromNullString(event.State),
		VenueCountry:         &event.Country,
		EventType:            event.EventType,
		StartDatetime:        event.StartDatetime,
		EndDatetime:          event.EndDatetime,
		TotalCapacity:        event.TotalCapacity,
		AvailableSeats:       event.AvailableSeats,
		BasePrice:            utils.ParsePrice(sql.NullString{String: event.BasePrice, Valid: event.BasePrice != ""}),
		MaxTicketsPerBooking: event.MaxTicketsPerBooking.Int32,
		Status:               event.Status.String,
		Version:              event.Version,
		CreatedBy:            event.CreatedBy,
		CreatedAt:            event.CreatedAt.Time,
		UpdatedAt:            event.UpdatedAt.Time,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) GetEventAvailability(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.PathValue("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	event, err := cfg.DB.GetEventByID(r.Context(), eventID)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondWithError(w, http.StatusNotFound, "Event not found")
			return
		}
		cfg.Logger.Error("Failed to fetch event availability", "error", err, "event_id", eventID)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch event availability")
		return
	}

	response := EventAvailabilityResponse{
		AvailableSeats: event.AvailableSeats,
		Status:         event.Status.String,
		LastUpdated:    event.UpdatedAt.Time,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}
