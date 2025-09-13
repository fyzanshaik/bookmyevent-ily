-- +goose Up
-- +goose StatementBegin

-- CRITICAL INDEXES FOR HIGH-TRAFFIC QUERIES AND CONCURRENCY
-- Note: Removing CONCURRENTLY for migration compatibility

-- 1. Events table - Core booking performance indexes
-- Composite index for available events (most critical for booking service)
CREATE INDEX idx_events_available_published 
ON events(status, available_seats, start_datetime) 
WHERE status = 'published' AND available_seats > 0;

-- Row-level locking optimization for concurrent updates
CREATE INDEX idx_events_id_version 
ON events(event_id, version);

-- Event browsing and filtering
CREATE INDEX idx_events_type_datetime 
ON events(event_type, start_datetime) 
WHERE status = 'published';

CREATE INDEX idx_events_city_datetime 
ON events(venue_id, start_datetime) 
WHERE status = 'published';

-- Admin queries optimization
CREATE INDEX idx_events_admin_status 
ON events(created_by, status, created_at);

-- 2. Venues table - Location-based searches
CREATE INDEX idx_venues_city_capacity 
ON venues(city, capacity);

CREATE INDEX idx_venues_location 
ON venues(city, state, country);

-- 3. Admins table - Authentication
CREATE INDEX idx_admins_email_active 
ON admins(email) 
WHERE is_active = true;

-- 4. Partial indexes for write-heavy scenarios
-- Events that need capacity tracking (non-draft, non-cancelled)
CREATE INDEX idx_events_capacity_tracking 
ON events(event_id, available_seats, total_capacity, version) 
WHERE status IN ('published', 'sold_out');

-- Events for CDC/sync to Elasticsearch (published only)
CREATE INDEX idx_events_search_sync 
ON events(updated_at, event_id) 
WHERE status = 'published';

-- Future event management (cleanup, notifications)
-- Note: Removing CURRENT_TIMESTAMP predicate due to immutability requirement
CREATE INDEX idx_events_upcoming 
ON events(start_datetime, status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_events_available_published;
DROP INDEX IF EXISTS idx_events_id_version;
DROP INDEX IF EXISTS idx_events_type_datetime;
DROP INDEX IF EXISTS idx_events_city_datetime;
DROP INDEX IF EXISTS idx_events_admin_status;
DROP INDEX IF EXISTS idx_venues_city_capacity;
DROP INDEX IF EXISTS idx_venues_location;
DROP INDEX IF EXISTS idx_admins_email_active;
DROP INDEX IF EXISTS idx_events_capacity_tracking;
DROP INDEX IF EXISTS idx_events_search_sync;
DROP INDEX IF EXISTS idx_events_upcoming;
-- +goose StatementEnd