
-- name: CreatePayment :one
INSERT INTO payments (
    booking_id, user_id, event_id, amount, currency, 
    payment_method, payment_gateway, gateway_transaction_id, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetPaymentByID :one
SELECT * FROM payments WHERE payment_id = $1;

-- name: GetPaymentByBookingID :one
SELECT * FROM payments WHERE booking_id = $1;

-- name: GetPaymentByGatewayTransactionID :one
SELECT * FROM payments WHERE gateway_transaction_id = $1;

-- name: UpdatePaymentStatus :one
UPDATE payments 
SET status = $2, updated_at = CURRENT_TIMESTAMP,
    completed_at = CASE WHEN $2 = 'completed' THEN CURRENT_TIMESTAMP ELSE completed_at END,
    failed_at = CASE WHEN $2 = 'failed' THEN CURRENT_TIMESTAMP ELSE failed_at END,
    error_message = $3
WHERE payment_id = $1 
RETURNING *;

-- name: UpdatePaymentTicketURL :one
UPDATE payments 
SET ticket_url = $2, updated_at = CURRENT_TIMESTAMP
WHERE payment_id = $1 
RETURNING *;

-- name: GetUserPayments :many
SELECT * FROM payments 
WHERE user_id = $1 AND status = 'completed'
ORDER BY completed_at DESC 
LIMIT $2 OFFSET $3;

-- name: GetEventPayments :many
SELECT * FROM payments 
WHERE event_id = $1 AND status = 'completed'
ORDER BY completed_at DESC 
LIMIT $2 OFFSET $3;

-- name: GetPaymentsForAnalytics :many
SELECT 
    event_id,
    COUNT(*) as total_payments,
    SUM(amount) as total_revenue,
    AVG(amount) as average_amount
FROM payments 
WHERE status = 'completed' 
    AND completed_at >= $1 
    AND completed_at <= $2
GROUP BY event_id
ORDER BY total_revenue DESC;

-- name: GetUserPaymentHistory :many
SELECT 
    p.*,
    b.booking_reference,
    b.quantity,
    b.event_id
FROM payments p
JOIN bookings b ON p.booking_id = b.booking_id
WHERE p.user_id = $1
ORDER BY p.completed_at DESC 
LIMIT $2 OFFSET $3;