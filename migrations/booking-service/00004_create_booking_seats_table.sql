-- +goose Up
-- Create booking_seats table - for future seat-level booking support
CREATE TABLE booking_seats (
    booking_seat_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL,
    seat_number VARCHAR(10),                 -- e.g., "A1", "B15"
    seat_row VARCHAR(10),                    -- e.g., "A", "B", "MEZZANINE-1"
    seat_section VARCHAR(50),                -- e.g., "ORCHESTRA", "BALCONY", "VIP"
    status VARCHAR(20) DEFAULT 'booked',     -- booked, cancelled, upgraded
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraint
    CONSTRAINT fk_booking_seat_booking
        FOREIGN KEY(booking_id)
        REFERENCES bookings(booking_id)
        ON DELETE CASCADE,
    
    -- Constraints
    CONSTRAINT check_seat_status CHECK (status IN ('booked', 'cancelled', 'upgraded'))
);

-- Indexes for performance
CREATE INDEX idx_booking_seats_booking ON booking_seats(booking_id);
CREATE INDEX idx_booking_seats_status ON booking_seats(status);
CREATE INDEX idx_booking_seats_section_row ON booking_seats(seat_section, seat_row);

-- +goose Down
DROP TABLE IF EXISTS booking_seats;