
-- name: CreateEvent :one
INSERT INTO events (
    name, description, venue_id, event_type, start_datetime, end_datetime,
    total_capacity, available_seats, base_price, max_tickets_per_booking,
    status, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING *;

-- name: GetEventByID :one
SELECT e.*, v.name as venue_name, v.address, v.city, v.state, v.country
FROM events e
JOIN venues v ON e.venue_id = v.venue_id
WHERE e.event_id = $1;

-- name: ListPublishedEvents :many
SELECT e.*, v.name as venue_name, v.city, v.state
FROM events e
JOIN venues v ON e.venue_id = v.venue_id
WHERE e.status = 'published'
  AND e.start_datetime > CURRENT_TIMESTAMP
  AND ($3::text = '' OR e.event_type = $3)
  AND ($4::text = '' OR v.city ILIKE '%' || $4 || '%')
  AND ($5::timestamp = '0001-01-01'::timestamp OR e.start_datetime >= $5)
  AND ($6::timestamp = '0001-01-01'::timestamp OR e.start_datetime <= $6)
ORDER BY e.start_datetime ASC
LIMIT $1 OFFSET $2;

-- name: CountPublishedEvents :one
SELECT COUNT(*)
FROM events e
JOIN venues v ON e.venue_id = v.venue_id
WHERE e.status = 'published'
  AND e.start_datetime > CURRENT_TIMESTAMP
  AND ($1::text = '' OR e.event_type = $1)
  AND ($2::text = '' OR v.city ILIKE '%' || $2 || '%')
  AND ($3::timestamp = '0001-01-01'::timestamp OR e.start_datetime >= $3)
  AND ($4::timestamp = '0001-01-01'::timestamp OR e.start_datetime <= $4);

-- name: GetEventForBooking :one
SELECT event_id, available_seats, total_capacity, max_tickets_per_booking,
       status, version, base_price, name
FROM events
WHERE event_id = $1
  AND (status = 'published' OR status = 'sold_out')
FOR UPDATE;

-- name: UpdateEventAvailability :one
UPDATE events
SET available_seats = available_seats - $2,
    version = version + 1,
    updated_at = CURRENT_TIMESTAMP,
    status = CASE
        WHEN (available_seats - $2) = 0 THEN 'sold_out'::text
        ELSE status
    END
WHERE event_id = $1
  AND version = $3
  AND available_seats >= $2
RETURNING event_id, available_seats, status, version;


-- name: ReturnEventSeats :one
UPDATE events
SET available_seats = available_seats + $2,
    version = version + 1,
    updated_at = CURRENT_TIMESTAMP,
    status = CASE
        WHEN status = 'sold_out' AND (available_seats + $2) > 0 THEN 'published'::text
        ELSE status
    END
WHERE event_id = $1
  AND version = $3
RETURNING event_id, available_seats, status, version;

-- name: UpdateEvent :one
UPDATE events
SET name = COALESCE($2, name),
    description = COALESCE($3, description),
    venue_id = COALESCE($4, venue_id),
    event_type = COALESCE($5, event_type),
    start_datetime = COALESCE($6, start_datetime),
    end_datetime = COALESCE($7, end_datetime),
    total_capacity = COALESCE($8, total_capacity),
    available_seats = COALESCE($9, available_seats),
    base_price = COALESCE($10, base_price),
    max_tickets_per_booking = COALESCE($11, max_tickets_per_booking),
    status = COALESCE($12, status),
    updated_at = CURRENT_TIMESTAMP,
    version = version + 1
WHERE event_id = $1
  AND version = $13
RETURNING *;

-- name: DeleteEvent :exec
UPDATE events
SET status = 'cancelled',
    updated_at = CURRENT_TIMESTAMP,
    version = version + 1
WHERE event_id = $1
  AND version = $2;

-- name: ListEventsByAdmin :many
SELECT e.*, v.name as venue_name, v.city
FROM events e
JOIN venues v ON e.venue_id = v.venue_id
WHERE e.created_by = $3
ORDER BY e.created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetEventAnalytics :one
SELECT
    e.event_id,
    e.name,
    e.total_capacity,
    e.available_seats,
    (e.total_capacity - e.available_seats) as tickets_sold,
    ROUND(((e.total_capacity - e.available_seats)::decimal / e.total_capacity::decimal) * 100, 2) as capacity_utilization,
    e.base_price,
    ROUND((e.total_capacity - e.available_seats) * e.base_price::decimal, 2) as estimated_revenue
FROM events e
WHERE e.event_id = $1;

-- name: CheckEventOwnership :one
SELECT event_id, created_by, status, version
FROM events
WHERE event_id = $1 AND created_by = $2;
