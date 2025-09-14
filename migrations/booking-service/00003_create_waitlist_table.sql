-- +goose Up
-- Create waitlist table - queue management
CREATE TABLE waitlist (
    waitlist_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL,                  -- Links to Event Service
    user_id UUID NOT NULL,                   -- Links to User Service
    position INTEGER NOT NULL,               -- Queue position (1, 2, 3...)
    quantity_requested INTEGER NOT NULL,
    status VARCHAR(20) DEFAULT 'waiting',    -- waiting, offered, expired, converted
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    offered_at TIMESTAMP,                    -- When offer was made
    expires_at TIMESTAMP,                    -- 2-minute exclusive booking window
    converted_at TIMESTAMP,                  -- When they successfully booked
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Unique constraint: one waitlist entry per user per event
    CONSTRAINT unique_user_event
        UNIQUE(user_id, event_id),
    
    -- Constraints
    CONSTRAINT check_quantity_requested CHECK (quantity_requested > 0),
    CONSTRAINT check_position CHECK (position > 0),
    CONSTRAINT check_waitlist_status CHECK (status IN ('waiting', 'offered', 'expired', 'converted'))
);

-- Indexes for performance
CREATE INDEX idx_waitlist_event_position ON waitlist(event_id, position) WHERE status = 'waiting';
CREATE INDEX idx_waitlist_user ON waitlist(user_id);
CREATE INDEX idx_waitlist_status ON waitlist(status);
CREATE INDEX idx_waitlist_offered_expires ON waitlist(expires_at) WHERE status = 'offered' AND expires_at IS NOT NULL;
CREATE INDEX idx_waitlist_event_status ON waitlist(event_id, status);

-- +goose Down
DROP TABLE IF EXISTS waitlist;