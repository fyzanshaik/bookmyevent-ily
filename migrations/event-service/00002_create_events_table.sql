-- +goose Up
-- +goose StatementBegin
CREATE TABLE events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    venue_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL, -- concert, sports, theater, conference
    start_datetime TIMESTAMP NOT NULL,
    end_datetime TIMESTAMP NOT NULL,
    total_capacity INTEGER NOT NULL CHECK (total_capacity > 0),
    available_seats INTEGER NOT NULL CHECK (available_seats >= 0),
    base_price DECIMAL(10, 2) NOT NULL CHECK (base_price >= 0),
    max_tickets_per_booking INTEGER DEFAULT 10 CHECK (max_tickets_per_booking > 0),
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'sold_out', 'cancelled')),
    
    -- Optimistic locking and concurrency control
    version INTEGER DEFAULT 1 NOT NULL,
    
    -- Admin tracking
    created_by UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints for data integrity and anti-overselling
    CONSTRAINT fk_venue
        FOREIGN KEY(venue_id)
        REFERENCES venues(venue_id),
    CONSTRAINT check_seats_capacity
        CHECK (available_seats <= total_capacity),
    CONSTRAINT check_dates
        CHECK (end_datetime > start_datetime),
    CONSTRAINT check_future_event
        CHECK (start_datetime > CURRENT_TIMESTAMP)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS events;
-- +goose StatementEnd