-- +goose Up
-- Create payments table - transaction records
CREATE TABLE payments (
    payment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL,
    user_id UUID NOT NULL,                   -- Denormalized for query performance
    event_id UUID NOT NULL,                  -- Denormalized for analytics
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    payment_method VARCHAR(50),              -- mock_payment, credit_card, debit_card
    payment_gateway VARCHAR(50),             -- mock_gateway, stripe, etc.
    gateway_transaction_id VARCHAR(255),     -- From payment processor
    status VARCHAR(20) NOT NULL,             -- initiated, processing, completed, failed
    ticket_url TEXT,                         -- Generated QR code/ticket link
    initiated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    failed_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraint
    CONSTRAINT fk_payment_booking
        FOREIGN KEY(booking_id)
        REFERENCES bookings(booking_id)
        ON DELETE CASCADE,
    
    -- Constraints
    CONSTRAINT check_amount CHECK (amount >= 0),
    CONSTRAINT check_payment_status CHECK (status IN ('initiated', 'processing', 'completed', 'failed')),
    CONSTRAINT check_currency CHECK (currency IN ('USD', 'EUR', 'GBP', 'INR'))
);

-- Indexes for performance
CREATE INDEX idx_payments_booking ON payments(booking_id);
CREATE INDEX idx_payments_user ON payments(user_id, status);
CREATE INDEX idx_payments_event ON payments(event_id, status);
CREATE INDEX idx_payments_gateway_transaction ON payments(gateway_transaction_id) WHERE gateway_transaction_id IS NOT NULL;
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_completed_at ON payments(completed_at) WHERE completed_at IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS payments;