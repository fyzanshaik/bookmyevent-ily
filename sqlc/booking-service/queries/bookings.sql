
-- name: CreateBooking :one
INSERT INTO bookings (
    user_id, event_id, booking_reference, quantity, total_amount, 
    status, payment_status, idempotency_key, expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetBookingByID :one
SELECT * FROM bookings WHERE booking_id = $1;

-- name: GetBookingByReference :one
SELECT * FROM bookings WHERE booking_reference = $1;

-- name: GetBookingByIdempotencyKey :one
SELECT * FROM bookings WHERE idempotency_key = $1;

-- name: UpdateBookingStatus :one
UPDATE bookings 
SET status = @status::VARCHAR, updated_at = CURRENT_TIMESTAMP,
    confirmed_at = CASE WHEN @status::VARCHAR = 'confirmed' THEN CURRENT_TIMESTAMP ELSE confirmed_at END,
    cancelled_at = CASE WHEN @status::VARCHAR = 'cancelled' THEN CURRENT_TIMESTAMP ELSE cancelled_at END
WHERE booking_id = @booking_id 
RETURNING *;

-- name: UpdateBookingPaymentStatus :one
UPDATE bookings 
SET payment_status = $2, updated_at = CURRENT_TIMESTAMP
WHERE booking_id = $1 
RETURNING *;

-- name: GetUserBookings :many
SELECT * FROM bookings 
WHERE user_id = $1 
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;

-- name: GetUserBookingsCount :one
SELECT COUNT(*) FROM bookings WHERE user_id = $1;

-- name: GetEventBookings :many
SELECT * FROM bookings 
WHERE event_id = $1 AND status = $2
ORDER BY created_at DESC 
LIMIT $3 OFFSET $4;

-- name: GetExpiredBookings :many
SELECT * FROM bookings 
WHERE status = 'pending' AND expires_at < CURRENT_TIMESTAMP
LIMIT $1;

-- name: GetPendingBookings :many
SELECT * FROM bookings 
WHERE status = 'pending'
LIMIT $1;

-- name: GetBookingsForCleanup :many
SELECT * FROM bookings 
WHERE status = 'pending' AND expires_at < CURRENT_TIMESTAMP;

-- name: DeleteBooking :exec
DELETE FROM bookings WHERE booking_id = $1;

-- name: GetBookingWithPayment :one
SELECT 
    b.*,
    p.payment_id,
    p.amount as payment_amount,
    p.gateway_transaction_id,
    p.payment_method,
    p.ticket_url,
    p.status as payment_status,
    p.completed_at as payment_completed_at
FROM bookings b
LEFT JOIN payments p ON b.booking_id = p.booking_id
WHERE b.booking_id = $1;