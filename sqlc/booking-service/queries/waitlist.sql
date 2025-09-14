-- Waitlist operations

-- name: JoinWaitlist :one
INSERT INTO waitlist (
    event_id, user_id, quantity_requested, position
) VALUES (
    $1, $2, $3, 
    COALESCE((SELECT MAX(position) FROM waitlist WHERE event_id = $1 AND status = 'waiting'), 0) + 1
) RETURNING *;

-- name: GetWaitlistEntry :one
SELECT * FROM waitlist WHERE waitlist_id = $1;

-- name: GetUserWaitlistEntry :one
SELECT * FROM waitlist WHERE user_id = $1 AND event_id = $2;

-- name: GetWaitlistPosition :one
SELECT position, status FROM waitlist 
WHERE user_id = $1 AND event_id = $2;

-- name: GetEventWaitlist :many
SELECT * FROM waitlist 
WHERE event_id = $1 AND status = 'waiting'
ORDER BY position ASC
LIMIT $2;

-- name: GetNextWaitlistEntries :many
SELECT * FROM waitlist 
WHERE event_id = $1 AND status = 'waiting'
ORDER BY position ASC
LIMIT $2;

-- name: UpdateWaitlistStatus :one
UPDATE waitlist 
SET status = $2, updated_at = CURRENT_TIMESTAMP,
    offered_at = CASE WHEN $2 = 'offered' THEN CURRENT_TIMESTAMP ELSE offered_at END,
    converted_at = CASE WHEN $2 = 'converted' THEN CURRENT_TIMESTAMP ELSE converted_at END,
    expires_at = $3
WHERE waitlist_id = $1 
RETURNING *;

-- name: GetOfferedWaitlistEntries :many
SELECT * FROM waitlist 
WHERE status = 'offered' AND expires_at < CURRENT_TIMESTAMP;

-- name: RemoveFromWaitlist :exec
DELETE FROM waitlist WHERE user_id = $1 AND event_id = $2;

-- name: GetWaitlistStats :one
SELECT 
    COUNT(*) as total_waiting,
    MIN(position) as first_position,
    MAX(position) as last_position,
    AVG(quantity_requested) as avg_quantity_requested
FROM waitlist 
WHERE event_id = $1 AND status = 'waiting';

-- name: ReorderWaitlistAfterRemoval :exec
UPDATE waitlist 
SET position = position - 1, updated_at = CURRENT_TIMESTAMP
WHERE event_id = $1 AND position > $2 AND status = 'waiting';

-- name: GetExpiredWaitlistOffers :many
SELECT * FROM waitlist 
WHERE status = 'offered' 
    AND expires_at IS NOT NULL 
    AND expires_at < CURRENT_TIMESTAMP;

-- name: GetWaitlistEntryByUserAndEvent :one
SELECT * FROM waitlist 
WHERE user_id = $1 AND event_id = $2;