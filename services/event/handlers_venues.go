package event

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/fyzanshaik/bookmyevent-ily/internal/auth"
	"github.com/fyzanshaik/bookmyevent-ily/internal/repository/events"
	"github.com/fyzanshaik/bookmyevent-ily/internal/utils"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

func (cfg *APIConfig) CreateVenue(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetAdminIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Admin not authenticated")
		return
	}

	var requestBody CreateVenueRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		cfg.Logger.WithFields(map[string]any{"error": err.Error()}).Warn("Invalid JSON in venue creation")
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if requestBody.Name == "" || requestBody.Address == "" || requestBody.City == "" {
		cfg.Logger.Warn("Venue creation with missing required fields")
		utils.RespondWithError(w, http.StatusBadRequest, "Name, address, and city are required")
		return
	}

	if requestBody.Capacity <= 0 {
		cfg.Logger.WithFields(map[string]any{"capacity": requestBody.Capacity}).Warn("Venue creation with invalid capacity")
		utils.RespondWithError(w, http.StatusBadRequest, "Capacity must be positive")
		return
	}
	country := requestBody.Country
	if country == "" {
		country = "USA"
	}
	var layoutConfig pqtype.NullRawMessage
	if requestBody.LayoutConfig != nil {
		layoutConfig = pqtype.NullRawMessage{RawMessage: requestBody.LayoutConfig, Valid: true}
	} else {
		layoutConfig = pqtype.NullRawMessage{RawMessage: json.RawMessage("{}"), Valid: true}
	}

	params := events.CreateVenueParams{
		Name:         requestBody.Name,
		Address:      requestBody.Address,
		City:         requestBody.City,
		State:        sql.NullString{String: requestBody.State, Valid: requestBody.State != ""},
		Country:      country,
		PostalCode:   sql.NullString{String: requestBody.PostalCode, Valid: requestBody.PostalCode != ""},
		Capacity:     requestBody.Capacity,
		LayoutConfig: layoutConfig,
	}

	venue, err := cfg.DB.CreateVenue(r.Context(), params)
	if err != nil {
		cfg.Logger.WithFields(map[string]any{"name": requestBody.Name, "error": err.Error()}).Error("Venue creation failed")
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not create venue")
		return
	}

	cfg.Logger.WithFields(map[string]any{"venue_id": venue.VenueID, "name": venue.Name}).Info("Venue created successfully")

	response := VenueResponse{
		VenueID:      venue.VenueID,
		Name:         venue.Name,
		Address:      venue.Address,
		City:         venue.City,
		State:        utils.StringPtrFromNullString(venue.State),
		Country:      venue.Country,
		PostalCode:   utils.StringPtrFromNullString(venue.PostalCode),
		Capacity:     venue.Capacity,
		LayoutConfig: venue.LayoutConfig.RawMessage,
		CreatedAt:    venue.CreatedAt.Time,
		UpdatedAt:    venue.UpdatedAt.Time,
	}

	utils.RespondWithJSON(w, http.StatusCreated, response)
}

func (cfg *APIConfig) ListVenues(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetAdminIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Admin not authenticated")
		return
	}

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

	city := r.URL.Query().Get("city")
	state := r.URL.Query().Get("state")
	search := r.URL.Query().Get("search")

	offset := (page - 1) * limit

	if search != "" {
		venues, err := cfg.DB.SearchVenues(r.Context(), sql.NullString{String: search, Valid: search != ""})
		if err != nil {
			cfg.Logger.Error("Failed to search venues", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to search venues")
			return
		}

		venueResponses := make([]VenueResponse, len(venues))
		for i, venue := range venues {
			venueResponses[i] = VenueResponse{
				VenueID:      venue.VenueID,
				Name:         venue.Name,
				Address:      venue.Address,
				City:         venue.City,
				State:        utils.StringPtrFromNullString(venue.State),
				Country:      venue.Country,
				PostalCode:   utils.StringPtrFromNullString(venue.PostalCode),
				Capacity:     venue.Capacity,
				LayoutConfig: venue.LayoutConfig.RawMessage,
				CreatedAt:    venue.CreatedAt.Time,
				UpdatedAt:    venue.UpdatedAt.Time,
			}
		}

		response := VenueListResponse{
			Venues:  venueResponses,
			Total:   int64(len(venueResponses)),
			Page:    page,
			Limit:   limit,
			HasMore: false,
		}

		utils.RespondWithJSON(w, http.StatusOK, response)
		return
	}

	countParams := events.CountVenuesParams{
		Column1: city,
		Column2: state,
	}

	total, err := cfg.DB.CountVenues(r.Context(), countParams)
	if err != nil {
		cfg.Logger.Error("Failed to count venues", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch venues")
		return
	}

	listParams := events.ListVenuesParams{
		Limit:   int32(limit),
		Offset:  int32(offset),
		Column3: city,
		Column4: state,
	}

	venues, err := cfg.DB.ListVenues(r.Context(), listParams)
	if err != nil {
		cfg.Logger.Error("Failed to fetch venues", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch venues")
		return
	}

	venueResponses := make([]VenueResponse, len(venues))
	for i, venue := range venues {
		venueResponses[i] = VenueResponse{
			VenueID:      venue.VenueID,
			Name:         venue.Name,
			Address:      venue.Address,
			City:         venue.City,
			State:        utils.StringPtrFromNullString(venue.State),
			Country:      venue.Country,
			PostalCode:   utils.StringPtrFromNullString(venue.PostalCode),
			Capacity:     venue.Capacity,
			LayoutConfig: venue.LayoutConfig.RawMessage,
			CreatedAt:    venue.CreatedAt.Time,
			UpdatedAt:    venue.UpdatedAt.Time,
		}
	}

	response := VenueListResponse{
		Venues:  venueResponses,
		Total:   total,
		Page:    page,
		Limit:   limit,
		HasMore: int64((page-1)*limit+len(venueResponses)) < total,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) UpdateVenue(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetAdminIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Admin not authenticated")
		return
	}

	venueIDStr := r.PathValue("id")
	venueID, err := uuid.Parse(venueIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid venue ID")
		return
	}

	var requestBody UpdateVenueRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	currentVenue, err := cfg.DB.GetVenueByID(r.Context(), venueID)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondWithError(w, http.StatusNotFound, "Venue not found")
			return
		}
		cfg.Logger.Error("Failed to get current venue", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get current venue")
		return
	}

	params := events.UpdateVenueParams{
		VenueID: venueID,
		Name: func() string {
			if requestBody.Name != nil {
				return *requestBody.Name
			}
			return currentVenue.Name
		}(),
		Address: func() string {
			if requestBody.Address != nil {
				return *requestBody.Address
			}
			return currentVenue.Address
		}(),
		City: func() string {
			if requestBody.City != nil {
				return *requestBody.City
			}
			return currentVenue.City
		}(),
		State: func() sql.NullString {
			if requestBody.State != nil {
				return sql.NullString{String: *requestBody.State, Valid: true}
			}
			return currentVenue.State
		}(),
		Country: func() string {
			if requestBody.Country != nil {
				return *requestBody.Country
			}
			return currentVenue.Country
		}(),
		PostalCode: func() sql.NullString {
			if requestBody.PostalCode != nil {
				return sql.NullString{String: *requestBody.PostalCode, Valid: true}
			}
			return currentVenue.PostalCode
		}(),
		Capacity: func() int32 {
			if requestBody.Capacity != nil {
				return *requestBody.Capacity
			}
			return currentVenue.Capacity
		}(),
		LayoutConfig: func() pqtype.NullRawMessage {
			if requestBody.LayoutConfig != nil {
				return pqtype.NullRawMessage{RawMessage: *requestBody.LayoutConfig, Valid: true}
			}
			return currentVenue.LayoutConfig
		}(),
	}

	updatedVenue, err := cfg.DB.UpdateVenue(r.Context(), params)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondWithError(w, http.StatusNotFound, "Venue not found")
			return
		}
		cfg.Logger.Error("Failed to update venue", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update venue")
		return
	}

	response := VenueResponse{
		VenueID:      updatedVenue.VenueID,
		Name:         updatedVenue.Name,
		Address:      updatedVenue.Address,
		City:         updatedVenue.City,
		State:        utils.StringPtrFromNullString(updatedVenue.State),
		Country:      updatedVenue.Country,
		PostalCode:   utils.StringPtrFromNullString(updatedVenue.PostalCode),
		Capacity:     updatedVenue.Capacity,
		LayoutConfig: updatedVenue.LayoutConfig.RawMessage,
		CreatedAt:    updatedVenue.CreatedAt.Time,
		UpdatedAt:    updatedVenue.UpdatedAt.Time,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) DeleteVenue(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetAdminIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Admin not authenticated")
		return
	}

	venueIDStr := r.PathValue("id")
	venueID, err := uuid.Parse(venueIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid venue ID")
		return
	}

	err = cfg.DB.DeleteVenue(r.Context(), venueID)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondWithError(w, http.StatusNotFound, "Venue not found")
			return
		}
		if strings.Contains(err.Error(), "foreign key") {
			utils.RespondWithError(w, http.StatusConflict, "Cannot delete venue with associated events")
			return
		}
		cfg.Logger.Error("Failed to delete venue", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete venue")
		return
	}

	response := SuccessResponse{
		Message: "Venue deleted successfully",
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}
