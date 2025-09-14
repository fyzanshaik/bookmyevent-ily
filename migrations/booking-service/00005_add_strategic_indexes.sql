-- +goose Up
-- Strategic indexes for booking service performance optimization

-- Composite indexes for common query patterns
CREATE INDEX idx_bookings_user_event ON bookings(user_id, event_id);
CREATE INDEX idx_bookings_event_status_date ON bookings(event_id, status, created_at);
CREATE INDEX idx_bookings_status_expires ON bookings(status, expires_at) WHERE status = 'pending';

-- Payment analytics indexes
CREATE INDEX idx_payments_event_completed ON payments(event_id, completed_at) WHERE status = 'completed';
CREATE INDEX idx_payments_user_date ON payments(user_id, completed_at) WHERE status = 'completed';
CREATE INDEX idx_payments_amount_date ON payments(completed_at, amount) WHERE status = 'completed';

-- Waitlist processing indexes
CREATE INDEX idx_waitlist_event_waiting ON waitlist(event_id, position, joined_at) WHERE status = 'waiting';
CREATE INDEX idx_waitlist_offered_processing ON waitlist(offered_at, expires_at) WHERE status = 'offered';

-- Booking reference lookup optimization (case-insensitive)
CREATE INDEX idx_bookings_reference_upper ON bookings(UPPER(booking_reference));

-- Time-based queries for cleanup and analytics
CREATE INDEX idx_bookings_created_at ON bookings(created_at);
CREATE INDEX idx_payments_created_at ON payments(created_at);
CREATE INDEX idx_waitlist_created_at ON waitlist(created_at);

-- Status transition tracking
CREATE INDEX idx_bookings_confirmed_at ON bookings(confirmed_at) WHERE confirmed_at IS NOT NULL;
CREATE INDEX idx_bookings_cancelled_at ON bookings(cancelled_at) WHERE cancelled_at IS NOT NULL;

-- +goose Down
-- Drop strategic indexes
DROP INDEX IF EXISTS idx_bookings_user_event;
DROP INDEX IF EXISTS idx_bookings_event_status_date;
DROP INDEX IF EXISTS idx_bookings_status_expires;
DROP INDEX IF EXISTS idx_payments_event_completed;
DROP INDEX IF EXISTS idx_payments_user_date;
DROP INDEX IF EXISTS idx_payments_amount_date;
DROP INDEX IF EXISTS idx_waitlist_event_waiting;
DROP INDEX IF EXISTS idx_waitlist_offered_processing;
DROP INDEX IF EXISTS idx_bookings_reference_upper;
DROP INDEX IF EXISTS idx_bookings_created_at;
DROP INDEX IF EXISTS idx_payments_created_at;
DROP INDEX IF EXISTS idx_waitlist_created_at;
DROP INDEX IF EXISTS idx_bookings_confirmed_at;
DROP INDEX IF EXISTS idx_bookings_cancelled_at;