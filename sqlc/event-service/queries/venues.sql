
-- name: CreateVenue :one
INSERT INTO venues (
    name, address, city, state, country, postal_code, capacity, layout_config
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetVenueByID :one
SELECT * FROM venues WHERE venue_id = $1;

-- name: ListVenues :many
SELECT * FROM venues
WHERE ($3::text IS NULL OR city ILIKE '%' || $3 || '%')
  AND ($4::text IS NULL OR state ILIKE '%' || $4 || '%')
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: CountVenues :one
SELECT COUNT(*) FROM venues
WHERE ($1::text IS NULL OR city ILIKE '%' || $1 || '%')
  AND ($2::text IS NULL OR state ILIKE '%' || $2 || '%');

-- name: UpdateVenue :one
UPDATE venues
SET name = COALESCE($2, name),
    address = COALESCE($3, address),
    city = COALESCE($4, city),
    state = COALESCE($5, state),
    country = COALESCE($6, country),
    postal_code = COALESCE($7, postal_code),
    capacity = COALESCE($8, capacity),
    layout_config = COALESCE($9, layout_config),
    updated_at = CURRENT_TIMESTAMP
WHERE venue_id = $1
RETURNING *;

-- name: DeleteVenue :exec
DELETE FROM venues WHERE venue_id = $1;

-- name: GetVenuesByCity :many
SELECT venue_id, name, capacity, address FROM venues
WHERE city = $1
ORDER BY name;

-- name: SearchVenues :many
SELECT * FROM venues
WHERE name ILIKE '%' || $1 || '%'
   OR city ILIKE '%' || $1 || '%'
   OR address ILIKE '%' || $1 || '%'
ORDER BY
    CASE WHEN name ILIKE $1 || '%' THEN 1 ELSE 2 END,
    name
LIMIT 10;
