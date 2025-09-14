-- +goose Up
-- Create bookings table - core booking entity
CREATE TABLE bookings (
    booking_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,                    -- Links to User Service
    event_id UUID NOT NULL,                   -- Links to Event Service
    booking_reference VARCHAR(20) UNIQUE NOT NULL, -- Human-readable: EVT-ABC123
    quantity INTEGER NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    status VARCHAR(20) NOT NULL,              -- pending, confirmed, cancelled, expired
    payment_status VARCHAR(20) NOT NULL,      -- pending, completed, failed, refunded
    idempotency_key VARCHAR(255) UNIQUE,      -- Prevent duplicate bookings
    booked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,                     -- 5-minute expiry for pending bookings
    confirmed_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT check_quantity CHECK (quantity > 0),
    CONSTRAINT check_total_amount CHECK (total_amount >= 0),
    CONSTRAINT check_status CHECK (status IN ('pending', 'confirmed', 'cancelled', 'expired')),
    CONSTRAINT check_payment_status CHECK (payment_status IN ('pending', 'completed', 'failed', 'refunded'))
);

-- Indexes for performance
CREATE INDEX idx_bookings_user ON bookings(user_id, status);
CREATE INDEX idx_bookings_event ON bookings(event_id, status);
CREATE INDEX idx_bookings_reference ON bookings(booking_reference);
CREATE INDEX idx_bookings_idempotency ON bookings(idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_expires_at ON bookings(expires_at) WHERE expires_at IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS bookings;