package booking

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
	"github.com/fyzanshaik/bookmyevent-ily/internal/repository/bookings"
	"github.com/fyzanshaik/bookmyevent-ily/internal/utils"
	"github.com/google/uuid"
)

func (cfg *APIConfig) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	if err := cfg.DB_Conn.Ping(); err != nil {
		utils.RespondWithError(w, http.StatusServiceUnavailable, "Database not ready")
		return
	}

	if err := cfg.RedisClient.client.Ping(r.Context()).Err(); err != nil {
		utils.RespondWithError(w, http.StatusServiceUnavailable, "Redis not ready")
		return
	}

	response := struct {
		Status   string `json:"status"`
		Database string `json:"database"`
		Redis    string `json:"redis"`
	}{
		Status:   "ready",
		Database: "connected",
		Redis:    "connected",
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.URL.Query().Get("event_id")
	quantityStr := r.URL.Query().Get("quantity")

	if eventIDStr == "" || quantityStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "event_id and quantity are required")
		return
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid event_id format")
		return
	}

	quantity, err := strconv.ParseInt(quantityStr, 10, 32)
	if err != nil || quantity <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid quantity")
		return
	}

	cachedSeats, err := cfg.RedisClient.GetCachedEventAvailability(r.Context(), eventID)
	if err == nil {
		response := CheckAvailabilityResponse{
			Available:      cachedSeats >= int32(quantity),
			AvailableSeats: cachedSeats,
			MaxPerBooking:  10,
			BasePrice:      0.0,
		}
		utils.RespondWithJSON(w, http.StatusOK, response)
		return
	}

	event, err := cfg.EventServiceClient.GetEventForBooking(r.Context(), eventID)
	if err != nil {
		cfg.Logger.Error("Failed to get event for availability check",
			"error", err, "event_id", eventID)

		if err.Error() == "event not found or not available for booking" {
			utils.RespondWithError(w, http.StatusNotFound, "Event not found or not available for booking")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check event availability")
		}
		return
	}

	cfg.RedisClient.CacheEventAvailability(r.Context(), eventID, event.AvailableSeats, 30*time.Second)
	response := CheckAvailabilityResponse{
		Available:      event.AvailableSeats >= int32(quantity),
		AvailableSeats: event.AvailableSeats,
		MaxPerBooking:  event.MaxTicketsPerBooking,
		BasePrice:      event.BasePrice,
	}

	cfg.Logger.Info("Availability check completed",
		"event_id", eventID,
		"requested_quantity", quantity,
		"available_seats", event.AvailableSeats,
		"available", response.Available)

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) ReserveSeats(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req ReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.EventID == uuid.Nil || req.Quantity <= 0 || req.IdempotencyKey == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "event_id, quantity, and idempotency_key are required")
		return
	}

	if req.Quantity > int32(cfg.Config.MaxTicketsPerUser) {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Maximum %d tickets allowed per booking", cfg.Config.MaxTicketsPerUser))
		return
	}

	allowed, err := cfg.RedisClient.CheckRateLimit(r.Context(), userID, int64(cfg.Config.RateLimitPerMinute))
	if err != nil {
		cfg.Logger.Error("Rate limit check failed", "error", err, "user_id", userID)
	} else if !allowed {
		utils.RespondWithError(w, http.StatusTooManyRequests, "Too many booking attempts. Please try again later.")
		return
	}

	cfg.RedisClient.IncrementRateLimit(r.Context(), userID, time.Minute)

	existingBooking, err := cfg.DB.GetBookingByIdempotencyKey(r.Context(), cfg.DB_Conn, sql.NullString{String: req.IdempotencyKey, Valid: true})
	if err == nil {
		cfg.Logger.Info("Idempotent booking request", "idempotency_key", req.IdempotencyKey, "existing_booking_id", existingBooking.BookingID)

		response := ReservationResponse{
			ReservationID:    existingBooking.BookingID,
			BookingReference: existingBooking.BookingReference,
			ExpiresAt:        existingBooking.ExpiresAt.Time,
			TotalAmount:      utils.ParseAmount(existingBooking.TotalAmount),
		}
		utils.RespondWithJSON(w, http.StatusOK, response)
		return
	}

	event, err := cfg.EventServiceClient.GetEventForBooking(r.Context(), req.EventID)
	if err != nil {
		cfg.Logger.Error("Failed to get event for reservation", "error", err, "event_id", req.EventID)
		utils.RespondWithError(w, http.StatusBadRequest, "Event not found or not available for booking")
		return
	}

	if req.Quantity > event.MaxTicketsPerBooking {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Maximum %d tickets allowed per booking for this event", event.MaxTicketsPerBooking))
		return
	}

	isWaitlistUser := false
	userWaitlistEntry, err := cfg.DB.GetWaitlistEntryByUserAndEvent(r.Context(), cfg.DB_Conn, bookings.GetWaitlistEntryByUserAndEventParams{
		UserID:  userID,
		EventID: req.EventID,
	})
	if err == nil && userWaitlistEntry.Status.String == "offered" {
		if userWaitlistEntry.ExpiresAt.Valid && time.Now().Before(userWaitlistEntry.ExpiresAt.Time) {
			isWaitlistUser = true
			cfg.Logger.Info("Waitlist user booking", "user_id", userID, "event_id", req.EventID)
		}
	}

	updateResp, err := cfg.EventServiceClient.UpdateAvailability(r.Context(), req.EventID, -req.Quantity, event.Version)
	if err != nil {
		cfg.Logger.Error("Failed to reserve seats", "error", err, "event_id", req.EventID, "quantity", req.Quantity, "version", event.Version)

		if strings.Contains(err.Error(), "Not enough seats available") {
			utils.RespondWithError(w, http.StatusConflict, "Not enough seats available")
		} else if strings.Contains(err.Error(), "updated by another process") {
			utils.RespondWithError(w, http.StatusConflict, "Event was updated by another process. Please retry.")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to reserve seats")
		}
		return
	}

	totalAmount := event.BasePrice * float64(req.Quantity)
	bookingRef := utils.GenerateBookingReference()

	var expiresAt time.Time
	var reservationExpiry time.Duration

	if isWaitlistUser {
		cfg.Logger.Info("Waitlist user booking reservation", "user_id", userID, "event_id", req.EventID)
		expiresAt = userWaitlistEntry.ExpiresAt.Time
		reservationExpiry = time.Until(expiresAt)
	} else {
		cfg.Logger.Info("Regular user booking reservation", "user_id", userID, "event_id", req.EventID)
		expiresAt = time.Now().Add(cfg.Config.ReservationExpiry)
		reservationExpiry = cfg.Config.ReservationExpiry
	}

	booking, err := cfg.DB.CreateBooking(r.Context(), cfg.DB_Conn, bookings.CreateBookingParams{
		UserID:           userID,
		EventID:          req.EventID,
		BookingReference: bookingRef,
		Quantity:         req.Quantity,
		TotalAmount:      fmt.Sprintf("%.2f", totalAmount),
		Status:           "pending",
		PaymentStatus:    "pending",
		IdempotencyKey:   sql.NullString{String: req.IdempotencyKey, Valid: true},
		ExpiresAt:        sql.NullTime{Time: expiresAt, Valid: true},
	})
	if err != nil {
		cfg.Logger.Error("Failed to create booking", "error", err)
		cfg.EventServiceClient.ReturnSeats(r.Context(), req.EventID, req.Quantity, updateResp.Version)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create booking")
		return
	}

	reservationData := &ReservationData{
		UserID:           userID,
		EventID:          req.EventID,
		Quantity:         req.Quantity,
		Amount:           totalAmount,
		BookingID:        booking.BookingID,
		BookingReference: bookingRef,
		ExpiresAt:        expiresAt,
	}

	if err := cfg.RedisClient.SetReservation(r.Context(), booking.BookingID, reservationData, reservationExpiry); err != nil {
		cfg.Logger.Error("Failed to store reservation in Redis", "error", err, "booking_id", booking.BookingID)
	}

	cfg.Logger.Info("Seats reserved successfully",
		"booking_id", booking.BookingID,
		"user_id", userID,
		"event_id", req.EventID,
		"quantity", req.Quantity,
		"expires_at", expiresAt)

	response := ReservationResponse{
		ReservationID:    booking.BookingID,
		BookingReference: bookingRef,
		ExpiresAt:        expiresAt,
		TotalAmount:      totalAmount,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) ConfirmBooking(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req ConfirmationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.ReservationID == uuid.Nil || req.PaymentToken == "" || req.PaymentMethod == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "reservation_id, payment_token, and payment_method are required")
		return
	}

	reservation, err := cfg.RedisClient.GetReservation(r.Context(), req.ReservationID)
	if err != nil {
		cfg.Logger.Error("Failed to get reservation", "error", err, "reservation_id", req.ReservationID)
		utils.RespondWithError(w, http.StatusBadRequest, "Reservation not found or expired")
		return
	}

	if reservation.UserID != userID {
		utils.RespondWithError(w, http.StatusForbidden, "Reservation does not belong to authenticated user")
		return
	}

	booking, err := cfg.DB.GetBookingByID(r.Context(), cfg.DB_Conn, req.ReservationID)
	if err != nil {
		cfg.Logger.Error("Failed to get booking", "error", err, "booking_id", req.ReservationID)
		utils.RespondWithError(w, http.StatusNotFound, "Booking not found")
		return
	}

	if booking.Status != "pending" {
		utils.RespondWithError(w, http.StatusConflict, "Booking is not in pending state")
		return
	}

	gatewayTxnID := utils.GenerateGatewayTransactionID()
	ticketURL := utils.GenerateTicketURL(booking.BookingReference)

	payment, err := cfg.DB.CreatePayment(r.Context(), cfg.DB_Conn, bookings.CreatePaymentParams{
		BookingID:            booking.BookingID,
		UserID:               userID,
		EventID:              booking.EventID,
		Amount:               booking.TotalAmount,
		Currency:             sql.NullString{String: "INR", Valid: true},
		PaymentMethod:        sql.NullString{String: req.PaymentMethod, Valid: true},
		PaymentGateway:       sql.NullString{String: "mock_gateway", Valid: true},
		GatewayTransactionID: sql.NullString{String: gatewayTxnID, Valid: true},
		Status:               "completed",
	})
	if err != nil {
		cfg.Logger.Error("Failed to create payment", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process payment")
		return
	}

	_, err = cfg.DB.UpdatePaymentTicketURL(r.Context(), cfg.DB_Conn, bookings.UpdatePaymentTicketURLParams{
		PaymentID: payment.PaymentID,
		TicketUrl: sql.NullString{String: ticketURL, Valid: true},
	})
	if err != nil {
		cfg.Logger.Error("Failed to update ticket URL", "error", err)
	}

	_, err = cfg.DB.UpdateBookingStatus(r.Context(), cfg.DB_Conn, bookings.UpdateBookingStatusParams{
		BookingID: booking.BookingID,
		Status:    "confirmed",
	})
	if err != nil {
		cfg.Logger.Error("Failed to update booking status", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to confirm booking")
		return
	}

	_, err = cfg.DB.UpdateBookingPaymentStatus(r.Context(), cfg.DB_Conn, bookings.UpdateBookingPaymentStatusParams{
		BookingID:     booking.BookingID,
		PaymentStatus: "completed",
	})
	if err != nil {
		cfg.Logger.Error("Failed to update payment status", "error", err)
	}

	cfg.RedisClient.DeleteReservation(r.Context(), req.ReservationID)
	cfg.RedisClient.InvalidateEventAvailabilityCache(r.Context(), booking.EventID)

	userWaitlistEntry, err := cfg.DB.GetWaitlistEntryByUserAndEvent(r.Context(), cfg.DB_Conn, bookings.GetWaitlistEntryByUserAndEventParams{
		UserID:  userID,
		EventID: booking.EventID,
	})
	if err == nil && userWaitlistEntry.Status.String == "offered" {
		_, err = cfg.DB.UpdateWaitlistStatus(r.Context(), cfg.DB_Conn, bookings.UpdateWaitlistStatusParams{
			WaitlistID: userWaitlistEntry.WaitlistID,
			Status:     sql.NullString{String: "converted", Valid: true},
			ExpiresAt:  sql.NullTime{Valid: false},
		})
		if err != nil {
			cfg.Logger.Error("Failed to convert waitlist status", "error", err, "user_id", userID)
		} else {
			cfg.Logger.Info("Waitlist user converted to booking", "user_id", userID, "event_id", booking.EventID)
		}
	}

	cfg.Logger.Info("Booking confirmed successfully",
		"booking_id", booking.BookingID,
		"user_id", userID,
		"event_id", booking.EventID,
		"payment_id", payment.PaymentID)

	response := ConfirmationResponse{
		BookingID:        booking.BookingID,
		BookingReference: booking.BookingReference,
		Status:           "confirmed",
		TicketURL:        ticketURL,
		Payment: PaymentInfo{
			TransactionID: gatewayTxnID,
			Status:        "completed",
			Amount:        utils.ParseAmount(booking.TotalAmount),
		},
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) GetBookingDetails(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	bookingIDStr := r.PathValue("id")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid booking ID format")
		return
	}

	bookingWithPayment, err := cfg.DB.GetBookingWithPayment(r.Context(), cfg.DB_Conn, bookingID)
	if err != nil {
		cfg.Logger.Error("Failed to get booking", "error", err, "booking_id", bookingID)
		utils.RespondWithError(w, http.StatusNotFound, "Booking not found")
		return
	}

	if bookingWithPayment.UserID != userID {
		utils.RespondWithError(w, http.StatusForbidden, "Access denied")
		return
	}

	event, err := cfg.EventServiceClient.GetEventForBooking(r.Context(), bookingWithPayment.EventID)
	if err != nil {
		cfg.Logger.Error("Failed to get event details", "error", err, "event_id", bookingWithPayment.EventID)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get event details")
		return
	}

	response := BookingDetailsResponse{
		BookingID:        bookingWithPayment.BookingID,
		BookingReference: bookingWithPayment.BookingReference,
		Event: EventInfo{
			Name:     event.Name,
			Venue:    "Event Venue",
			DateTime: time.Now().Add(24 * time.Hour),
		},
		Quantity:      bookingWithPayment.Quantity,
		TotalAmount:   utils.ParseAmount(bookingWithPayment.TotalAmount),
		Status:        bookingWithPayment.Status,
		PaymentStatus: bookingWithPayment.PaymentStatus,
		BookedAt:      bookingWithPayment.BookedAt.Time,
	}

	if bookingWithPayment.ConfirmedAt.Valid {
		response.ConfirmedAt = &bookingWithPayment.ConfirmedAt.Time
	}

	if bookingWithPayment.TicketUrl.Valid {
		response.TicketURL = bookingWithPayment.TicketUrl.String
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) CancelBooking(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	bookingIDStr := r.PathValue("id")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid booking ID format")
		return
	}

	booking, err := cfg.DB.GetBookingByID(r.Context(), cfg.DB_Conn, bookingID)
	if err != nil {
		cfg.Logger.Error("Failed to get booking", "error", err, "booking_id", bookingID)
		utils.RespondWithError(w, http.StatusNotFound, "Booking not found")
		return
	}

	if booking.UserID != userID {
		utils.RespondWithError(w, http.StatusForbidden, "Access denied")
		return
	}

	if booking.Status == "cancelled" {
		utils.RespondWithError(w, http.StatusConflict, "Booking is already cancelled")
		return
	}

	if booking.Status == "expired" {
		utils.RespondWithError(w, http.StatusConflict, "Cannot cancel expired booking")
		return
	}

	originalAmount := utils.ParseAmount(booking.TotalAmount)
	refundAmount := utils.CalculateRefundAmount(originalAmount, booking.BookedAt.Time, time.Now().Add(48*time.Hour))

	_, err = cfg.DB.UpdateBookingStatus(r.Context(), cfg.DB_Conn, bookings.UpdateBookingStatusParams{
		BookingID: bookingID,
		Status:    "cancelled",
	})
	if err != nil {
		cfg.Logger.Error("Failed to cancel booking", "error", err, "booking_id", bookingID)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to cancel booking")
		return
	}

	if booking.Status == "confirmed" {
		event, err := cfg.EventServiceClient.GetEventForBooking(r.Context(), booking.EventID)
		if err != nil {
			cfg.Logger.Error("Failed to get event for cancellation", "error", err, "event_id", booking.EventID)
		} else {
			_, err = cfg.EventServiceClient.ReturnSeats(r.Context(), booking.EventID, booking.Quantity, event.Version)
			if err != nil {
				cfg.Logger.Error("Failed to return seats", "error", err, "booking_id", bookingID)
			}
		}

		cfg.ProcessWaitlist(r.Context(), booking.EventID, booking.Quantity)
	}

	refundStatus := "none"
	if refundAmount > 0 {
		refundStatus = "processed"
	}

	cfg.RedisClient.DeleteReservation(r.Context(), bookingID)
	cfg.RedisClient.InvalidateEventAvailabilityCache(r.Context(), booking.EventID)

	cfg.Logger.Info("Booking cancelled",
		"booking_id", bookingID,
		"user_id", userID,
		"refund_amount", refundAmount)

	response := CancellationResponse{
		Message:      "Booking cancelled successfully",
		RefundStatus: refundStatus,
		RefundAmount: refundAmount,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) GetUserBookings(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	userBookings, err := cfg.DB.GetUserBookings(r.Context(), cfg.DB_Conn, bookings.GetUserBookingsParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		cfg.Logger.Error("Failed to get user bookings", "error", err, "user_id", userID)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get bookings")
		return
	}

	total, err := cfg.DB.GetUserBookingsCount(r.Context(), cfg.DB_Conn, userID)
	if err != nil {
		cfg.Logger.Error("Failed to get user bookings count", "error", err, "user_id", userID)
		total = 0
	}

	bookingList := make([]BookingDetailsResponse, len(userBookings))
	for i, booking := range userBookings {
		bookingList[i] = BookingDetailsResponse{
			BookingID:        booking.BookingID,
			BookingReference: booking.BookingReference,
			Quantity:         booking.Quantity,
			TotalAmount:      utils.ParseAmount(booking.TotalAmount),
			Status:           booking.Status,
			PaymentStatus:    booking.PaymentStatus,
			BookedAt:         booking.BookedAt.Time,
		}

		if booking.ConfirmedAt.Valid {
			bookingList[i].ConfirmedAt = &booking.ConfirmedAt.Time
		}
	}

	response := map[string]any{
		"bookings":    bookingList,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (int(total) + limit - 1) / limit,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) JoinWaitlist(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req JoinWaitlistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.EventID == uuid.Nil || req.Quantity <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "event_id and quantity are required")
		return
	}

	existingEntry, err := cfg.DB.GetWaitlistEntryByUserAndEvent(r.Context(), cfg.DB_Conn, bookings.GetWaitlistEntryByUserAndEventParams{
		UserID:  userID,
		EventID: req.EventID,
	})
	if err == nil {
		response := JoinWaitlistResponse{
			WaitlistID:    existingEntry.WaitlistID,
			Position:      existingEntry.Position,
			EstimatedWait: cfg.calculateEstimatedWait(existingEntry.Position),
			Status:        existingEntry.Status.String,
		}
		utils.RespondWithJSON(w, http.StatusOK, response)
		return
	}

	event, err := cfg.EventServiceClient.GetEventForBooking(r.Context(), req.EventID)
	if err != nil {
		cfg.Logger.Error("Failed to get event for waitlist", "error", err, "event_id", req.EventID)
		utils.RespondWithError(w, http.StatusBadRequest, "Event not found")
		return
	}

	if event.AvailableSeats >= req.Quantity {
		utils.RespondWithError(w, http.StatusBadRequest, "Seats are available, please book directly")
		return
	}

	waitlistEntry, err := cfg.DB.JoinWaitlist(r.Context(), cfg.DB_Conn, bookings.JoinWaitlistParams{
		EventID:           req.EventID,
		UserID:            userID,
		QuantityRequested: req.Quantity,
	})
	if err != nil {
		cfg.Logger.Error("Failed to join waitlist", "error", err, "user_id", userID, "event_id", req.EventID)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to join waitlist")
		return
	}

	cfg.Logger.Info("User joined waitlist",
		"user_id", userID,
		"event_id", req.EventID,
		"position", waitlistEntry.Position,
		"quantity", req.Quantity)

	response := JoinWaitlistResponse{
		WaitlistID:    waitlistEntry.WaitlistID,
		Position:      waitlistEntry.Position,
		EstimatedWait: cfg.calculateEstimatedWait(waitlistEntry.Position),
		Status:        "waiting",
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) GetWaitlistPosition(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	eventIDStr := r.URL.Query().Get("event_id")
	if eventIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "event_id is required")
		return
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid event_id format")
		return
	}

	waitlistEntry, err := cfg.DB.GetWaitlistEntryByUserAndEvent(r.Context(), cfg.DB_Conn, bookings.GetWaitlistEntryByUserAndEventParams{
		UserID:  userID,
		EventID: eventID,
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Not in waitlist for this event")
		return
	}

	stats, err := cfg.DB.GetWaitlistStats(r.Context(), cfg.DB_Conn, eventID)
	if err != nil {
		cfg.Logger.Error("Failed to get waitlist stats", "error", err, "event_id", eventID)
		stats.TotalWaiting = 0
	}

	response := WaitlistPositionResponse{
		Position:          waitlistEntry.Position,
		TotalWaiting:      int32(stats.TotalWaiting),
		Status:            waitlistEntry.Status.String,
		EstimatedWait:     cfg.calculateEstimatedWait(waitlistEntry.Position),
		QuantityRequested: waitlistEntry.QuantityRequested,
	}

	if waitlistEntry.Status.String == "offered" {
		if waitlistEntry.ExpiresAt.Valid {
			response.ExpiresAt = &waitlistEntry.ExpiresAt.Time
		}
		if waitlistEntry.OfferedAt.Valid {
			response.OfferedAt = &waitlistEntry.OfferedAt.Time
		}
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) LeaveWaitlist(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req LeaveWaitlistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.EventID == uuid.Nil {
		utils.RespondWithError(w, http.StatusBadRequest, "event_id is required")
		return
	}

	waitlistEntry, err := cfg.DB.GetWaitlistEntryByUserAndEvent(r.Context(), cfg.DB_Conn, bookings.GetWaitlistEntryByUserAndEventParams{
		UserID:  userID,
		EventID: req.EventID,
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Not in waitlist for this event")
		return
	}

	err = cfg.DB.RemoveFromWaitlist(r.Context(), cfg.DB_Conn, bookings.RemoveFromWaitlistParams{
		UserID:  userID,
		EventID: req.EventID,
	})
	if err != nil {
		cfg.Logger.Error("Failed to remove from waitlist", "error", err, "user_id", userID, "event_id", req.EventID)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to leave waitlist")
		return
	}

	err = cfg.DB.ReorderWaitlistAfterRemoval(r.Context(), cfg.DB_Conn, bookings.ReorderWaitlistAfterRemovalParams{
		EventID:  req.EventID,
		Position: waitlistEntry.Position,
	})
	if err != nil {
		cfg.Logger.Error("Failed to reorder waitlist", "error", err, "event_id", req.EventID)
	}

	cfg.Logger.Info("User left waitlist",
		"user_id", userID,
		"event_id", req.EventID,
		"position", waitlistEntry.Position)

	response := map[string]string{
		"message": "Successfully removed from waitlist",
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) GetBookingInternal(w http.ResponseWriter, r *http.Request) {
	bookingIDStr := r.PathValue("id")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid booking ID format")
		return
	}

	booking, err := cfg.DB.GetBookingByID(r.Context(), cfg.DB_Conn, bookingID)
	if err != nil {
		cfg.Logger.Error("Failed to get booking", "error", err, "booking_id", bookingID)
		utils.RespondWithError(w, http.StatusNotFound, "Booking not found")
		return
	}

	response := map[string]any{
		"booking_id":        booking.BookingID,
		"user_id":           booking.UserID,
		"event_id":          booking.EventID,
		"booking_reference": booking.BookingReference,
		"quantity":          booking.Quantity,
		"total_amount":      utils.ParseAmount(booking.TotalAmount),
		"status":            booking.Status,
		"payment_status":    booking.PaymentStatus,
		"booked_at":         booking.BookedAt,
	}

	if booking.ConfirmedAt.Valid {
		response["confirmed_at"] = booking.ConfirmedAt.Time
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) ExpireReservations(w http.ResponseWriter, r *http.Request) {
	expiredBookings, err := cfg.DB.GetExpiredBookings(r.Context(), cfg.DB_Conn, 100)
	if err != nil {
		cfg.Logger.Error("Failed to get expired bookings", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process expired reservations")
		return
	}

	processed := 0
	for _, booking := range expiredBookings {
		_, err := cfg.DB.UpdateBookingStatus(r.Context(), cfg.DB_Conn, bookings.UpdateBookingStatusParams{
			BookingID: booking.BookingID,
			Status:    "expired",
		})
		if err != nil {
			cfg.Logger.Error("Failed to expire booking", "error", err, "booking_id", booking.BookingID)
			continue
		}

		event, err := cfg.EventServiceClient.GetEventForBooking(r.Context(), booking.EventID)
		if err != nil {
			cfg.Logger.Error("Failed to get event for seat return", "error", err, "event_id", booking.EventID)
		} else {
			_, err = cfg.EventServiceClient.ReturnSeats(r.Context(), booking.EventID, booking.Quantity, event.Version)
			if err != nil {
				cfg.Logger.Error("Failed to return seats", "error", err, "booking_id", booking.BookingID, "event_id", booking.EventID)
			}
		}

		cfg.RedisClient.DeleteReservation(r.Context(), booking.BookingID)
		cfg.RedisClient.InvalidateEventAvailabilityCache(r.Context(), booking.EventID)
		processed++

		cfg.Logger.Info("Expired booking processed",
			"booking_id", booking.BookingID,
			"event_id", booking.EventID,
			"quantity", booking.Quantity)

		cfg.ProcessWaitlist(r.Context(), booking.EventID, booking.Quantity)
	}

	response := map[string]any{
		"processed": processed,
		"total":     len(expiredBookings),
	}

	cfg.ExpireWaitlistOffers(r.Context())

	cfg.Logger.Info("Reservation expiry job completed", "processed", processed, "total", len(expiredBookings))
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) ForceExpireAll(w http.ResponseWriter, r *http.Request) {
	allPendingBookings, err := cfg.DB.GetPendingBookings(r.Context(), cfg.DB_Conn, 1000)
	if err != nil {
		cfg.Logger.Error("Failed to get pending bookings for force expiry", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get pending bookings")
		return
	}

	processed := 0
	for _, booking := range allPendingBookings {
		_, err := cfg.DB.UpdateBookingStatus(r.Context(), cfg.DB_Conn, bookings.UpdateBookingStatusParams{
			BookingID: booking.BookingID,
			Status:    "expired",
		})
		if err != nil {
			cfg.Logger.Error("Failed to force expire booking", "error", err, "booking_id", booking.BookingID)
			continue
		}

		event, err := cfg.EventServiceClient.GetEventForBooking(r.Context(), booking.EventID)
		if err != nil {
			cfg.Logger.Error("Failed to get event for seat return in force expiry", "error", err, "event_id", booking.EventID)
		} else {
			_, err = cfg.EventServiceClient.ReturnSeats(r.Context(), booking.EventID, booking.Quantity, event.Version)
			if err != nil {
				cfg.Logger.Error("Failed to return seats in force expiry", "error", err, "booking_id", booking.BookingID, "event_id", booking.EventID)
			}
		}

		cfg.RedisClient.DeleteReservation(r.Context(), booking.BookingID)
		cfg.RedisClient.InvalidateEventAvailabilityCache(r.Context(), booking.EventID)
		processed++

		cfg.Logger.Info("Force expired booking",
			"booking_id", booking.BookingID,
			"event_id", booking.EventID,
			"quantity", booking.Quantity)

		cfg.ProcessWaitlist(r.Context(), booking.EventID, booking.Quantity)
	}

	response := map[string]any{
		"message":   "Force expired all pending reservations",
		"processed": processed,
		"total":     len(allPendingBookings),
	}

	cfg.ExpireWaitlistOffers(r.Context())

	cfg.Logger.Info("Force expiry job completed", "processed", processed, "total", len(allPendingBookings))
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) ManualExpireReservation(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	bookingIDStr := r.PathValue("id")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid booking ID format")
		return
	}

	booking, err := cfg.DB.GetBookingByID(r.Context(), cfg.DB_Conn, bookingID)
	if err != nil {
		cfg.Logger.Error("Failed to get booking for manual expiry", "error", err, "booking_id", bookingID)
		utils.RespondWithError(w, http.StatusNotFound, "Booking not found")
		return
	}

	if booking.UserID != userID {
		utils.RespondWithError(w, http.StatusForbidden, "Access denied")
		return
	}

	if booking.Status != "pending" {
		utils.RespondWithError(w, http.StatusConflict, "Only pending bookings can be manually expired")
		return
	}

	_, err = cfg.DB.UpdateBookingStatus(r.Context(), cfg.DB_Conn, bookings.UpdateBookingStatusParams{
		BookingID: bookingID,
		Status:    "expired",
	})
	if err != nil {
		cfg.Logger.Error("Failed to expire booking manually", "error", err, "booking_id", bookingID)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to expire booking")
		return
	}

	event, err := cfg.EventServiceClient.GetEventForBooking(r.Context(), booking.EventID)
	if err != nil {
		cfg.Logger.Error("Failed to get event for manual seat return", "error", err, "event_id", booking.EventID)
	} else {
		_, err = cfg.EventServiceClient.ReturnSeats(r.Context(), booking.EventID, booking.Quantity, event.Version)
		if err != nil {
			cfg.Logger.Error("Failed to return seats manually", "error", err, "booking_id", bookingID, "event_id", booking.EventID)
		}
	}

	cfg.RedisClient.DeleteReservation(r.Context(), bookingID)
	cfg.RedisClient.InvalidateEventAvailabilityCache(r.Context(), booking.EventID)

	cfg.ProcessWaitlist(r.Context(), booking.EventID, booking.Quantity)

	cfg.Logger.Info("Booking manually expired",
		"booking_id", bookingID,
		"user_id", userID,
		"event_id", booking.EventID,
		"quantity", booking.Quantity)

	response := map[string]string{
		"message": "Reservation expired successfully",
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) calculateEstimatedWait(position int32) string {
	if position == 1 {
		return "Next in line"
	} else if position <= 5 {
		return "5-15 minutes"
	} else if position <= 20 {
		return "15-60 minutes"
	}
	return "More than 1 hour"
}

func (cfg *APIConfig) ProcessWaitlist(ctx context.Context, eventID uuid.UUID, availableSeats int32) {
	if availableSeats <= 0 {
		return
	}

	nextEntries, err := cfg.DB.GetNextWaitlistEntries(ctx, cfg.DB_Conn, bookings.GetNextWaitlistEntriesParams{
		EventID: eventID,
		Limit:   availableSeats * 2,
	})
	if err != nil {
		cfg.Logger.Error("Failed to get waitlist entries", "error", err, "event_id", eventID)
		return
	}

	if len(nextEntries) == 0 {
		return
	}

	seatsToOffer := availableSeats
	for _, entry := range nextEntries {
		if seatsToOffer <= 0 {
			break
		}

		if entry.QuantityRequested <= seatsToOffer || seatsToOffer == availableSeats {
			expiresAt := time.Now().Add(2 * time.Minute)

			_, err := cfg.DB.SetWaitlistOffered(ctx, cfg.DB_Conn, bookings.SetWaitlistOfferedParams{
				WaitlistID: entry.WaitlistID,
				ExpiresAt:  sql.NullTime{Time: expiresAt, Valid: true},
			})
			if err != nil {
				cfg.Logger.Error("Failed to update waitlist status", "error", err, "waitlist_id", entry.WaitlistID)
				continue
			}

			err = cfg.DB.ReorderWaitlistAfterRemoval(ctx, cfg.DB_Conn, bookings.ReorderWaitlistAfterRemovalParams{
				EventID:  eventID,
				Position: entry.Position,
			})
			if err != nil {
				cfg.Logger.Error("Failed to reorder waitlist after offer", "error", err, "event_id", eventID, "position", entry.Position)
			}

			cfg.Logger.Info("Waitlist offer created",
				"user_id", entry.UserID,
				"event_id", eventID,
				"position", entry.Position,
				"seats_offered", min(entry.QuantityRequested, seatsToOffer),
				"expires_at", expiresAt)

			seatsToOffer -= min(entry.QuantityRequested, seatsToOffer)
		}
	}
}

func (cfg *APIConfig) ExpireWaitlistOffers(ctx context.Context) error {
	expiredOffers, err := cfg.DB.GetExpiredWaitlistOffers(ctx, cfg.DB_Conn)
	if err != nil {
		return err
	}

	for _, offer := range expiredOffers {
		stats, err := cfg.DB.GetWaitlistStats(ctx, cfg.DB_Conn, offer.EventID)
		if err != nil {
			cfg.Logger.Error("Failed to get waitlist stats for expired offer", "error", err, "event_id", offer.EventID)
			continue
		}

		newPosition := int32(1)
		if stats.TotalWaiting > 0 {
			if lastPos, ok := stats.LastPosition.(int32); ok {
				newPosition = lastPos + 1
			} else if lastPos, ok := stats.LastPosition.(int64); ok {
				newPosition = int32(lastPos) + 1
			}
		}

		err = cfg.DB.ReassignWaitlistPosition(ctx, cfg.DB_Conn, bookings.ReassignWaitlistPositionParams{
			WaitlistID: offer.WaitlistID,
			Position:   newPosition,
		})
		if err != nil {
			cfg.Logger.Error("Failed to reassign position for expired offer", "error", err, "waitlist_id", offer.WaitlistID)
			continue
		}

		_, err = cfg.DB.SetWaitlistWaiting(ctx, cfg.DB_Conn, offer.WaitlistID)
		if err != nil {
			cfg.Logger.Error("Failed to expire waitlist offer", "error", err, "waitlist_id", offer.WaitlistID)
			continue
		}

		cfg.Logger.Info("Waitlist offer expired - user moved to end of queue",
			"user_id", offer.UserID,
			"event_id", offer.EventID,
			"old_position", offer.Position,
			"new_position", newPosition)
	}

	return nil
}

func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
