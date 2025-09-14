# Event Service Testing Guide

## Service Overview
- **Port**: 8002
- **Database**: `events_db` on PostgreSQL:5433
- **Architecture**: Admin-managed events with venue support
- **Key Features**: Optimistic locking, anti-overselling, internal booking APIs

## Authentication System

### Admin JWT Structure
```go
type AdminClaims struct {
    AdminID     uuid.UUID `json:"admin_id"`
    Role        string    `json:"role"`
    Permissions string    `json:"permissions"`
    jwt.RegisteredClaims
}
```

### Auth Headers
```
Authorization: Bearer <admin_jwt_token>
X-API-Key: <internal_service_key> // For internal endpoints
```

## API Endpoints for Testing

### 1. Admin Authentication
```bash
# Admin Registration
POST /api/v1/auth/admin/register
Content-Type: application/json
{
  "email": "admin@test.com",
  "password": "password123",
  "name": "Test Admin",
  "phone_number": "+1234567890",
  "role": "admin"
}

# Admin Login  
POST /api/v1/auth/admin/login
Content-Type: application/json
{
  "email": "admin@test.com",
  "password": "password123"
}

# Token Refresh
POST /api/v1/auth/admin/refresh
Content-Type: application/json
{
  "refresh_token": "<refresh_token>"
}
```

### 2. Venue Management (Admin Required)
```bash
# Create Venue
POST /api/v1/admin/venues
Authorization: Bearer <admin_token>
Content-Type: application/json
{
  "name": "Test Venue",
  "address": "123 Main St",
  "city": "New York",
  "state": "NY",
  "country": "USA",
  "postal_code": "10001",
  "capacity": 1000,
  "layout_config": {"sections": ["A", "B", "C"]}
}

# List Venues
GET /api/v1/admin/venues?page=1&limit=20&city=New York
Authorization: Bearer <admin_token>

# Update Venue
PUT /api/v1/admin/venues/{venue_id}
Authorization: Bearer <admin_token>
Content-Type: application/json
{
  "name": "Updated Venue Name",
  "capacity": 1200
}

# Delete Venue
DELETE /api/v1/admin/venues/{venue_id}
Authorization: Bearer <admin_token>
```

### 3. Event Management (Admin Required)
```bash
# Create Event
POST /api/v1/admin/events
Authorization: Bearer <admin_token>
Content-Type: application/json
{
  "name": "Concert Event",
  "description": "Amazing concert",
  "venue_id": "<venue_uuid>",
  "event_type": "concert",
  "start_datetime": "2024-12-01T20:00:00Z",
  "end_datetime": "2024-12-01T23:00:00Z",
  "total_capacity": 500,
  "base_price": 99.99,
  "max_tickets_per_booking": 8
}

# List Admin Events
GET /api/v1/admin/events?page=1&limit=20
Authorization: Bearer <admin_token>

# Update Event
PUT /api/v1/admin/events/{event_id}
Authorization: Bearer <admin_token>
Content-Type: application/json
{
  "name": "Updated Event Name",
  "base_price": 109.99,
  "version": 1
}

# Delete Event
DELETE /api/v1/admin/events/{event_id}
Authorization: Bearer <admin_token>
Content-Type: application/json
{
  "version": 1
}

# Get Event Analytics
GET /api/v1/admin/events/{event_id}/analytics
Authorization: Bearer <admin_token>
```

### 4. Public Event Endpoints (No Auth)
```bash
# List Published Events
GET /api/v1/events?page=1&limit=20&event_type=concert&city=New York&date_from=2024-12-01&date_to=2024-12-31

# Get Event Details
GET /api/v1/events/{event_id}

# Get Event Availability
GET /api/v1/events/{event_id}/availability
```

### 5. Internal Service Endpoints (API Key Required)
```bash
# Update Event Availability (Reserve Seats)
POST /internal/events/{event_id}/update-availability
X-API-Key: <internal_service_key>
Content-Type: application/json
{
  "quantity": -5,  // Negative = reserve seats
  "version": 1
}

# Return Seats (Cancellation)
POST /internal/events/{event_id}/return-seats
X-API-Key: <internal_service_key>
Content-Type: application/json
{
  "quantity": 3,  // Positive = return seats
  "version": 2
}

# Get Event for Booking (Row Lock)
GET /internal/events/{event_id}
X-API-Key: <internal_service_key>
```

## Test Data Setup

### Sample Admin User
```sql
-- Will be created via /api/v1/auth/admin/register endpoint
-- Default role: "admin"
-- Default permissions: "{}"
```

### Sample Venue
```json
{
  "name": "Madison Square Garden",
  "address": "4 Pennsylvania Plaza",
  "city": "New York",
  "state": "NY",
  "country": "USA",
  "postal_code": "10001",
  "capacity": 20000,
  "layout_config": {
    "sections": ["Floor", "100s", "200s", "300s"],
    "wheelchair_accessible": true
  }
}
```

### Sample Event
```json
{
  "name": "Rock Concert 2024",
  "description": "Epic rock concert with multiple bands",
  "event_type": "concert",
  "start_datetime": "2024-12-15T20:00:00Z",
  "end_datetime": "2024-12-15T23:30:00Z",
  "total_capacity": 15000,
  "base_price": 125.00,
  "max_tickets_per_booking": 10
}
```

## Critical Testing Scenarios

### 1. Concurrency Control Testing
```bash
# Test optimistic locking
# 1. Get event (note version)
# 2. Try to update with old version (should fail)
# 3. Update with correct version (should succeed)
```

### 2. Anti-Overselling Testing
```bash
# 1. Create event with capacity 100
# 2. Reserve 98 seats via internal API
# 3. Try to reserve 5 more seats (should fail - not enough available)
# 4. Try to reserve 2 seats (should succeed)
# 5. Event status should be 'sold_out'
```

### 3. Seat Return Testing
```bash
# 1. Reserve 50 seats (capacity: 100, available: 50)
# 2. Return 20 seats (available: 70)
# 3. If event was 'sold_out', status should change to 'published'
```

## Database Schema Key Fields

### Events Table
- `available_seats`: Tracks real-time availability
- `version`: For optimistic locking
- `status`: 'draft', 'published', 'sold_out', 'cancelled'
- `base_price`: Stored as DECIMAL, returned as string

### Venues Table
- `capacity`: Maximum venue capacity
- `layout_config`: JSONB for flexible seating layouts

### Admins Table
- `role`: Admin role (admin, super_admin, etc.)
- `permissions`: JSONB for granular permissions
- `is_active`: For soft delete/deactivation

## Response Formats

### Success Response
```json
{
  "event_id": "uuid",
  "name": "Event Name",
  "venue_name": "Venue Name",
  "available_seats": 450,
  "base_price": 99.99,
  "status": "published",
  "version": 3
}
```

### Error Response
```json
{
  "error": {
    "code": "INSUFFICIENT_SEATS",
    "message": "Not enough seats available",
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

## Environment Variables Needed
```bash
EVENT_SERVICE_PORT=8002
EVENT_SERVICE_DB_URL="postgres://user:password@localhost:5433/events_db"
JWT_SECRET="your-jwt-secret"
INTERNAL_API_KEY="your-internal-api-key"
```

## Key Testing Tools
- **Postman/Insomnia**: For API testing
- **curl**: For command-line testing
- **Artillery/k6**: For load testing concurrency
- **PostgreSQL client**: For database state verification