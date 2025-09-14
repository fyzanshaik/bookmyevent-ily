-- Booking seats operations (future feature)

-- name: CreateBookingSeat :one
INSERT INTO booking_seats (
    booking_id, seat_number, seat_row, seat_section, status
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetBookingSeats :many
SELECT * FROM booking_seats WHERE booking_id = $1;

-- name: GetBookingSeat :one
SELECT * FROM booking_seats WHERE booking_seat_id = $1;

-- name: UpdateBookingSeatStatus :one
UPDATE booking_seats 
SET status = $2, updated_at = CURRENT_TIMESTAMP
WHERE booking_seat_id = $1 
RETURNING *;

-- name: GetSeatsBySection :many
SELECT * FROM booking_seats 
WHERE booking_id = $1 AND seat_section = $2;

-- name: DeleteBookingSeats :exec
DELETE FROM booking_seats WHERE booking_id = $1;

-- name: GetBookingSeatsCount :one
SELECT COUNT(*) FROM booking_seats WHERE booking_id = $1;