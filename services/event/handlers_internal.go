package event

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/fyzanshaik/bookmyevent-ily/internal/repository/events"
	"github.com/fyzanshaik/bookmyevent-ily/internal/utils"
	"github.com/google/uuid"
)

func (cfg *APIConfig) UpdateEventAvailability(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.PathValue("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	var requestBody UpdateAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		cfg.Logger.WithFields(map[string]any{"error": err.Error()}).Warn("Invalid JSON in availability update")
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if requestBody.Quantity == 0 {
		cfg.Logger.WithFields(map[string]any{"event_id": eventID}).Warn("Availability update with zero quantity")
		utils.RespondWithError(w, http.StatusBadRequest, "Quantity cannot be zero")
		return
	}

	if requestBody.Version <= 0 {
		cfg.Logger.WithFields(map[string]any{"event_id": eventID, "version": requestBody.Version}).Warn("Availability update with invalid version")
		utils.RespondWithError(w, http.StatusBadRequest, "Version must be positive")
		return
	}

	if requestBody.Quantity < 0 {
		seatsToReserve := -requestBody.Quantity

		params := events.UpdateEventAvailabilityParams{
			EventID:        eventID,
			AvailableSeats: seatsToReserve,
			Version:        requestBody.Version,
		}

		result, err := cfg.DB.UpdateEventAvailability(r.Context(), params)
		if err != nil {
			cfg.Logger.Error("Failed to update event availability (reserve)",
				"error", err,
				"event_id", eventID,
				"quantity", seatsToReserve,
				"version", requestBody.Version)

			if strings.Contains(err.Error(), "available_seats") {
				utils.RespondWithError(w, http.StatusConflict, "Not enough seats available")
				return
			}
			if strings.Contains(err.Error(), "version") || err == sql.ErrNoRows {
				utils.RespondWithError(w, http.StatusConflict, "Event was updated by another process. Please retry.")
				return
			}
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to reserve seats")
			return
		}

		response := UpdateAvailabilityResponse{
			EventID:        result.EventID,
			AvailableSeats: result.AvailableSeats,
			Status:         result.Status.String,
			Version:        result.Version,
		}

		cfg.Logger.Info("Successfully reserved seats",
			"event_id", eventID,
			"seats_reserved", seatsToReserve,
			"remaining_seats", result.AvailableSeats,
			"new_version", result.Version)

		fmt.Printf("Reserving %d seats for event %s, %d seats remaining\n", seatsToReserve, eventID.String(), result.AvailableSeats)
		cfg.updateSearchAvailability(r.Context(), eventID, result.AvailableSeats)

		utils.RespondWithJSON(w, http.StatusOK, response)

	} else {
		params := events.ReturnEventSeatsParams{
			EventID:        eventID,
			AvailableSeats: requestBody.Quantity,
			Version:        requestBody.Version,
		}

		result, err := cfg.DB.ReturnEventSeats(r.Context(), params)
		if err != nil {
			cfg.Logger.Error("Failed to return event seats",
				"error", err,
				"event_id", eventID,
				"quantity", requestBody.Quantity,
				"version", requestBody.Version)

			if strings.Contains(err.Error(), "version") || err == sql.ErrNoRows {
				utils.RespondWithError(w, http.StatusConflict, "Event was updated by another process. Please retry.")
				return
			}
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to return seats")
			return
		}

		response := UpdateAvailabilityResponse{
			EventID:        result.EventID,
			AvailableSeats: result.AvailableSeats,
			Status:         result.Status.String,
			Version:        result.Version,
		}

		cfg.Logger.Info("Successfully returned seats",
			"event_id", eventID,
			"seats_returned", requestBody.Quantity,
			"available_seats", result.AvailableSeats,
			"new_version", result.Version)

		fmt.Printf("Return here %d seats for event %s, %d seats available\n", requestBody.Quantity, eventID.String(), result.AvailableSeats)
		cfg.updateSearchAvailability(r.Context(), eventID, result.AvailableSeats)
		utils.RespondWithJSON(w, http.StatusOK, response)
	}
}

func (cfg *APIConfig) ReturnEventSeats(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.PathValue("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	var requestBody ReturnSeatsRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if requestBody.Quantity <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "Quantity must be positive")
		return
	}

	if requestBody.Version <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "Version must be positive")
		return
	}

	params := events.ReturnEventSeatsParams{
		EventID:        eventID,
		AvailableSeats: requestBody.Quantity,
		Version:        requestBody.Version,
	}

	result, err := cfg.DB.ReturnEventSeats(r.Context(), params)
	if err != nil {
		cfg.Logger.Error("Failed to return seats",
			"error", err,
			"event_id", eventID,
			"quantity", requestBody.Quantity,
			"version", requestBody.Version)

		if strings.Contains(err.Error(), "version") || err == sql.ErrNoRows {
			utils.RespondWithError(w, http.StatusConflict, "Event was updated by another process. Please retry.")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to return seats")
		return
	}

	response := ReturnSeatsResponse{
		EventID:        result.EventID,
		AvailableSeats: result.AvailableSeats,
		Status:         result.Status.String,
		Version:        result.Version,
	}

	cfg.Logger.Info("Successfully returned seats via dedicated endpoint",
		"event_id", eventID,
		"seats_returned", requestBody.Quantity,
		"available_seats", result.AvailableSeats,
		"new_version", result.Version)

	fmt.Printf("Returned %d seats/internal/events/{id}/return-seats endpoint for event %s, %d seats available\n", requestBody.Quantity, eventID.String(), result.AvailableSeats)
	cfg.updateSearchAvailability(r.Context(), eventID, result.AvailableSeats)
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) GetEventForBooking(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.PathValue("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	event, err := cfg.DB.GetEventForBooking(r.Context(), eventID)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondWithError(w, http.StatusNotFound, "Event not found or not available for booking")
			return
		}
		cfg.Logger.Error("Failed to fetch event for booking", "error", err, "event_id", eventID)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch event")
		return
	}

	response := EventForBookingResponse{
		EventID:              event.EventID,
		AvailableSeats:       event.AvailableSeats,
		MaxTicketsPerBooking: event.MaxTicketsPerBooking.Int32,
		BasePrice: func() float64 {
			var price float64
			if _, err := fmt.Sscanf(event.BasePrice, "%f", &price); err != nil {
				return 0.0
			}
			return price
		}(),
		Version: event.Version,
		Status:  event.Status.String,
		Name:    event.Name,
	}

	cfg.Logger.Info("Event fetched for booking validation",
		"event_id", eventID,
		"available_seats", event.AvailableSeats,
		"status", event.Status,
		"version", event.Version)

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) updateSearchAvailability(ctx context.Context, eventID uuid.UUID, availableSeats int32) {
	if cfg.SearchClient == nil {
		return
	}

	go func() {
		event, err := cfg.DB.GetEventByID(ctx, eventID)
		if err != nil {
			cfg.Logger.Error("Failed to get event for search availability update", "error", err, "event_id", eventID)
			return
		}

		venue, err := cfg.DB.GetVenueByID(ctx, event.VenueID)
		if err != nil {
			cfg.Logger.Error("Failed to get venue for search availability update", "error", err, "venue_id", event.VenueID)
			return
		}

		response := EventResponse{
			EventID:        event.EventID,
			Name:           event.Name,
			Description:    utils.StringPtrFromNullString(event.Description),
			VenueID:        event.VenueID,
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

		if err := cfg.SearchClient.UpdateEvent(ctx, response, venueResp); err != nil {
			cfg.Logger.Error("Failed to update event availability in search service", "error", err, "event_id", eventID)
		} else {
			cfg.Logger.Debug("Updated event availability in search service", "event_id", eventID, "available_seats", availableSeats)
		}
	}()
}
