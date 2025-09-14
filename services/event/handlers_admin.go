package event

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fyzanshaik/bookmyevent-ily/internal/auth"
	"github.com/fyzanshaik/bookmyevent-ily/internal/repository/events"
	"github.com/fyzanshaik/bookmyevent-ily/internal/utils"
	"github.com/google/uuid"
)

func (cfg *APIConfig) CreateEvent(w http.ResponseWriter, r *http.Request) {
	adminID, ok := auth.GetAdminIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Admin not authenticated")
		return
	}

	var requestBody CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if requestBody.Name == "" || requestBody.EventType == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Name and event_type are required")
		return
	}

	if requestBody.VenueID == uuid.Nil {
		utils.RespondWithError(w, http.StatusBadRequest, "venue_id is required")
		return
	}

	now := time.Now()
	if requestBody.StartDatetime.Before(now) {
		utils.RespondWithError(w, http.StatusBadRequest, "Start datetime must be in the future")
		return
	}

	if requestBody.EndDatetime.Before(now) {
		utils.RespondWithError(w, http.StatusBadRequest, "End datetime must be in the future")
		return
	}

	if requestBody.StartDatetime.After(requestBody.EndDatetime) || requestBody.StartDatetime.Equal(requestBody.EndDatetime) {
		utils.RespondWithError(w, http.StatusBadRequest, "Start datetime must be before end datetime")
		return
	}

	if requestBody.TotalCapacity <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "Total capacity must be positive")
		return
	}

	if requestBody.BasePrice < 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "Base price cannot be negative")
		return
	}

	maxTickets := requestBody.MaxTicketsPerBooking
	if maxTickets <= 0 {
		maxTickets = 10
	}

	params := events.CreateEventParams{
		Name:                 requestBody.Name,
		Description:          sql.NullString{String: requestBody.Description, Valid: requestBody.Description != ""},
		VenueID:              requestBody.VenueID,
		EventType:            requestBody.EventType,
		StartDatetime:        requestBody.StartDatetime,
		EndDatetime:          requestBody.EndDatetime,
		TotalCapacity:        requestBody.TotalCapacity,
		AvailableSeats:       requestBody.TotalCapacity,
		BasePrice:            fmt.Sprintf("%.2f", requestBody.BasePrice),
		MaxTicketsPerBooking: sql.NullInt32{Int32: maxTickets, Valid: true},
		Status:               sql.NullString{String: "draft", Valid: true},
		CreatedBy:            adminID,
	}

	event, err := cfg.DB.CreateEvent(r.Context(), params)
	if err != nil {
		cfg.Logger.Error("Failed to create event", "error", err)
		if strings.Contains(err.Error(), "foreign key") ||
			strings.Contains(err.Error(), "violates foreign key constraint") ||
			strings.Contains(err.Error(), "fk_venue") {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid venue ID - venue does not exist")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not create event")
		return
	}

	fmt.Printf("Created event in db: %s (ID: %s)\n", event.Name, event.EventID.String())
	cfg.Logger.Info("Event created successfully", "event_id", event.EventID, "event_name", event.Name, "admin_id", adminID)

	response := EventResponse{
		EventID:              event.EventID,
		Name:                 event.Name,
		Description:          utils.StringPtrFromNullString(event.Description),
		VenueID:              event.VenueID,
		EventType:            event.EventType,
		StartDatetime:        event.StartDatetime,
		EndDatetime:          event.EndDatetime,
		TotalCapacity:        event.TotalCapacity,
		AvailableSeats:       event.AvailableSeats,
		BasePrice:            requestBody.BasePrice,
		MaxTicketsPerBooking: event.MaxTicketsPerBooking.Int32,
		Status:               event.Status.String,
		Version:              event.Version,
		CreatedBy:            event.CreatedBy,
		CreatedAt:            event.CreatedAt.Time,
		UpdatedAt:            event.UpdatedAt.Time,
	}

	venue, err := cfg.DB.GetVenueByID(r.Context(), event.VenueID)
	if err != nil {
		cfg.Logger.Error("Failed to get venue for search indexing", "error", err, "venue_id", event.VenueID)
	} else {
		fmt.Printf("DEBUG: SearchClient status: %v (nil=%t)\n", cfg.SearchClient, cfg.SearchClient == nil)
		if cfg.SearchClient != nil {
			fmt.Printf("DEBUG: Attempting to index event %s in search service\n", event.Name)
			go func() {
				venueResp := VenueResponse{
					VenueID:      venue.VenueID,
					Name:         venue.Name,
					Address:      venue.Address,
					City:         venue.City,
					State:        utils.StringPtrFromNullString(venue.State),
					Country:      venue.Country,
					PostalCode:   utils.StringPtrFromNullString(venue.PostalCode),
					Capacity:     venue.Capacity,
					LayoutConfig: utils.NullRawMessageToJSONRawMessage(venue.LayoutConfig),
					CreatedAt:    venue.CreatedAt.Time,
					UpdatedAt:    venue.UpdatedAt.Time,
				}

				fmt.Printf("DEBUG: Calling SearchClient.IndexEvent for event %s\n", response.Name)
				ctx := context.Background()
				if err := cfg.SearchClient.IndexEvent(ctx, response, venueResp); err != nil {
					cfg.Logger.Error("Failed to index event in search service", "error", err, "event_id", event.EventID)
					fmt.Printf("DEBUG: IndexEvent failed: %v\n", err)
				} else {
					fmt.Printf("DEBUG: IndexEvent succeeded for event %s\n", response.Name)
				}
			}()
		} else {
			fmt.Printf("DEBUG: SearchClient is nil, search indexing disabled\n")
		}
	}

	utils.RespondWithJSON(w, http.StatusCreated, response)
}

func (cfg *APIConfig) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	adminID, ok := auth.GetAdminIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Admin not authenticated")
		return
	}

	eventIDStr := r.PathValue("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	var requestBody UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	_, err = cfg.DB.CheckEventOwnership(r.Context(), events.CheckEventOwnershipParams{
		EventID:   eventID,
		CreatedBy: adminID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondWithError(w, http.StatusNotFound, "Event not found or you don't have permission")
			return
		}
		cfg.Logger.Error("Failed to check event ownership", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to verify event ownership")
		return
	}

	currentEvent, err := cfg.DB.GetEventByID(r.Context(), eventID)
	if err != nil {
		cfg.Logger.Error("Failed to get current event", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get current event")
		return
	}

	params := events.UpdateEventParams{
		EventID: eventID,
		Name: func() string {
			if requestBody.Name != nil {
				return *requestBody.Name
			}
			return currentEvent.Name
		}(),
		Description: func() sql.NullString {
			if requestBody.Description != nil {
				return sql.NullString{String: *requestBody.Description, Valid: true}
			}
			return currentEvent.Description
		}(),
		VenueID: func() uuid.UUID {
			if requestBody.VenueID != nil {
				return *requestBody.VenueID
			}
			return currentEvent.VenueID
		}(),
		EventType: func() string {
			if requestBody.EventType != nil {
				return *requestBody.EventType
			}
			return currentEvent.EventType
		}(),
		StartDatetime: func() time.Time {
			if requestBody.StartDatetime != nil {
				return *requestBody.StartDatetime
			}
			return currentEvent.StartDatetime
		}(),
		EndDatetime: func() time.Time {
			if requestBody.EndDatetime != nil {
				return *requestBody.EndDatetime
			}
			return currentEvent.EndDatetime
		}(),
		TotalCapacity: func() int32 {
			if requestBody.TotalCapacity != nil {
				return *requestBody.TotalCapacity
			}
			return currentEvent.TotalCapacity
		}(),
		AvailableSeats: func() int32 {
			if requestBody.AvailableSeats != nil {
				return *requestBody.AvailableSeats
			}
			return currentEvent.AvailableSeats
		}(),
		BasePrice: func() string {
			if requestBody.BasePrice != nil {
				return fmt.Sprintf("%.2f", *requestBody.BasePrice)
			}
			return currentEvent.BasePrice
		}(),
		MaxTicketsPerBooking: func() sql.NullInt32 {
			if requestBody.MaxTicketsPerBooking != nil {
				return sql.NullInt32{Int32: *requestBody.MaxTicketsPerBooking, Valid: true}
			}
			return currentEvent.MaxTicketsPerBooking
		}(),
		Status: func() sql.NullString {
			if requestBody.Status != nil {
				return sql.NullString{String: *requestBody.Status, Valid: true}
			}
			return currentEvent.Status
		}(),
		Version: requestBody.Version,
	}

	updatedEvent, err := cfg.DB.UpdateEvent(r.Context(), params)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondWithError(w, http.StatusConflict, "Event was updated by another admin. Please refresh and try again.")
			return
		}
		if strings.Contains(err.Error(), "version") {
			utils.RespondWithError(w, http.StatusConflict, "Event was updated by another admin. Please refresh and try again.")
			return
		}
		cfg.Logger.Error("Failed to update event", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update event")
		return
	}

	fmt.Printf("Update event in db: %s (ID: %s)\n", updatedEvent.Name, updatedEvent.EventID.String())
	cfg.Logger.Info("Event updated successfully", "event_id", updatedEvent.EventID, "event_name", updatedEvent.Name, "admin_id", adminID)

	response := EventResponse{
		EventID:        updatedEvent.EventID,
		Name:           updatedEvent.Name,
		Description:    utils.StringPtrFromNullString(updatedEvent.Description),
		VenueID:        updatedEvent.VenueID,
		EventType:      updatedEvent.EventType,
		StartDatetime:  updatedEvent.StartDatetime,
		EndDatetime:    updatedEvent.EndDatetime,
		TotalCapacity:  updatedEvent.TotalCapacity,
		AvailableSeats: updatedEvent.AvailableSeats,
		BasePrice: func() float64 {
			var price float64
			if _, err := fmt.Sscanf(updatedEvent.BasePrice, "%f", &price); err != nil {
				return 0.0
			}
			return price
		}(),
		MaxTicketsPerBooking: updatedEvent.MaxTicketsPerBooking.Int32,
		Status:               updatedEvent.Status.String,
		Version:              updatedEvent.Version,
		CreatedBy:            updatedEvent.CreatedBy,
		CreatedAt:            updatedEvent.CreatedAt.Time,
		UpdatedAt:            updatedEvent.UpdatedAt.Time,
	}

	venue, err := cfg.DB.GetVenueByID(r.Context(), updatedEvent.VenueID)
	if err != nil {
		cfg.Logger.Error("Failed to get venue for search indexing", "error", err, "venue_id", updatedEvent.VenueID)
	} else if cfg.SearchClient != nil {
		go func() {
			venueResp := VenueResponse{
				VenueID:      venue.VenueID,
				Name:         venue.Name,
				Address:      venue.Address,
				City:         venue.City,
				State:        utils.StringPtrFromNullString(venue.State),
				Country:      venue.Country,
				PostalCode:   utils.StringPtrFromNullString(venue.PostalCode),
				Capacity:     venue.Capacity,
				LayoutConfig: utils.NullRawMessageToJSONRawMessage(venue.LayoutConfig),
				CreatedAt:    venue.CreatedAt.Time,
				UpdatedAt:    venue.UpdatedAt.Time,
			}

			ctx := context.Background()
			if err := cfg.SearchClient.UpdateEvent(ctx, response, venueResp); err != nil {
				cfg.Logger.Error("Failed to update event in search service", "error", err, "event_id", updatedEvent.EventID)
			}
		}()
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	adminID, ok := auth.GetAdminIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Admin not authenticated")
		return
	}

	eventIDStr := r.PathValue("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	var requestBody struct {
		Version int32 `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	_, err = cfg.DB.CheckEventOwnership(r.Context(), events.CheckEventOwnershipParams{
		EventID:   eventID,
		CreatedBy: adminID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondWithError(w, http.StatusNotFound, "Event not found or you don't have permission")
			return
		}
		cfg.Logger.Error("Failed to check event ownership", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to verify event ownership")
		return
	}

	err = cfg.DB.DeleteEvent(r.Context(), events.DeleteEventParams{
		EventID: eventID,
		Version: requestBody.Version,
	})
	if err != nil {
		if strings.Contains(err.Error(), "version") {
			utils.RespondWithError(w, http.StatusConflict, "Event was updated by another admin. Please refresh and try again.")
			return
		}
		cfg.Logger.Error("Failed to delete event", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete event")
		return
	}

	fmt.Printf("deleted event eventservice: %s\n", eventID.String())
	cfg.Logger.Info("Event deleted successfully", "event_id", eventID, "admin_id", adminID)

	if cfg.SearchClient != nil {
		go func() {
			if err := cfg.SearchClient.DeleteEvent(r.Context(), eventID); err != nil {
				cfg.Logger.Error("Failed to delete event from search service", "error", err, "event_id", eventID)
			}
		}()
	}

	response := SuccessResponse{
		Message: "Event cancelled successfully",
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) ListAdminEvents(w http.ResponseWriter, r *http.Request) {
	adminID, ok := auth.GetAdminIDFromContext(r.Context())
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

	offset := (page - 1) * limit

	params := events.ListEventsByAdminParams{
		Limit:     int32(limit),
		Offset:    int32(offset),
		CreatedBy: adminID,
	}

	eventsList, err := cfg.DB.ListEventsByAdmin(r.Context(), params)
	if err != nil {
		cfg.Logger.Error("Failed to fetch admin events", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch events")
		return
	}

	eventResponses := make([]EventResponse, len(eventsList))
	for i, event := range eventsList {
		eventResponses[i] = EventResponse{
			EventID:        event.EventID,
			Name:           event.Name,
			Description:    utils.StringPtrFromNullString(event.Description),
			VenueID:        event.VenueID,
			VenueName:      &event.VenueName,
			VenueCity:      &event.City,
			EventType:      event.EventType,
			StartDatetime:  event.StartDatetime,
			EndDatetime:    event.EndDatetime,
			TotalCapacity:  event.TotalCapacity,
			AvailableSeats: event.AvailableSeats,
			BasePrice: func() float64 {
				var price float64
				if _, err := fmt.Sscanf(event.BasePrice, "%f", &price); err != nil {
					return 0.0
				}
				return price
			}(),
			MaxTicketsPerBooking: event.MaxTicketsPerBooking.Int32,
			Status:               event.Status.String,
			Version:              event.Version,
			CreatedBy:            event.CreatedBy,
			CreatedAt:            event.CreatedAt.Time,
			UpdatedAt:            event.UpdatedAt.Time,
		}
	}

	response := EventListResponse{
		Events:  eventResponses,
		Total:   int64(len(eventResponses)),
		Page:    page,
		Limit:   limit,
		HasMore: len(eventResponses) == limit,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) GetEventAnalytics(w http.ResponseWriter, r *http.Request) {
	adminID, ok := auth.GetAdminIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Admin not authenticated")
		return
	}

	eventIDStr := r.PathValue("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	_, err = cfg.DB.CheckEventOwnership(r.Context(), events.CheckEventOwnershipParams{
		EventID:   eventID,
		CreatedBy: adminID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondWithError(w, http.StatusNotFound, "Event not found or you don't have permission")
			return
		}
		cfg.Logger.Error("Failed to check event ownership", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to verify event ownership")
		return
	}

	analytics, err := cfg.DB.GetEventAnalytics(r.Context(), eventID)
	if err != nil {
		cfg.Logger.Error("Failed to fetch event analytics", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch analytics")
		return
	}

	response := EventAnalyticsResponse{
		EventID:        analytics.EventID,
		Name:           analytics.Name,
		TotalCapacity:  analytics.TotalCapacity,
		AvailableSeats: analytics.AvailableSeats,
		TicketsSold:    analytics.TicketsSold,
		CapacityUtilization: func() float64 {
			var util float64
			if _, err := fmt.Sscanf(analytics.CapacityUtilization, "%f", &util); err != nil {
				return 0.0
			}
			return util
		}(),
		BasePrice: func() float64 {
			var price float64
			if _, err := fmt.Sscanf(analytics.BasePrice, "%f", &price); err != nil {
				return 0.0
			}
			return price
		}(),
		EstimatedRevenue: func() float64 {
			var revenue float64
			if _, err := fmt.Sscanf(analytics.EstimatedRevenue, "%f", &revenue); err != nil {
				return 0.0
			}
			return revenue
		}(),
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}
