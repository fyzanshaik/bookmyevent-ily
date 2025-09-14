# Event Service API Documentation

**Port:** 8002 | **Base URL:** `http://localhost:8002` | **Database:** `events_db`

The Event Service manages events, venues, and provides booking integration APIs with atomic seat management.

---

## 🔐 Admin Authentication

### Register Admin
```http
POST /api/v1/auth/admin/register
Content-Type: application/json

{
  "email": "admin@bookmyevent.com",
  "password": "admin123",
  "name": "Event Admin",
  "phone_number": "+1800555000",
  "role": "event_manager"
}
```
**Response:** Admin credentials + JWT tokens

### Login Admin
```http
POST /api/v1/auth/admin/login
Content-Type: application/json

{
  "email": "admin@bookmyevent.com", 
  "password": "admin123"
}
```
**Response:** JWT access + refresh tokens

### Refresh Token
```http
POST /api/v1/auth/admin/refresh
Content-Type: application/json

{
  "refresh_token": "5c114af28d0f..."
}
```

---

## 🏢 Venue Management (Admin Only)

**All venue endpoints require:** `Authorization: Bearer <admin_token>`

### Create Venue
```http
POST /api/v1/admin/venues
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "Madison Square Garden",
  "address": "4 Pennsylvania Plaza", 
  "city": "New York",
  "state": "NY",
  "country": "USA",
  "postal_code": "10001",
  "capacity": 20000
}
```
**Response:** Venue details with `venue_id`

### List Venues
```http
GET /api/v1/admin/venues?page=1&limit=10&city=NYC
Authorization: Bearer <admin_token>
```

### Update/Delete Venue
```http
PUT /api/v1/admin/venues/{venue_id}
DELETE /api/v1/admin/venues/{venue_id}
```

---

## 🎪 Event Management

### Create Event (Admin)
```http
POST /api/v1/admin/events
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "The Rolling Stones World Tour",
  "description": "Legendary rock band live",
  "venue_id": "2b583951-1171-4ee6-94f4-5ef0839a1014",
  "event_type": "concert",
  "start_datetime": "2025-12-15T20:00:00Z",
  "end_datetime": "2025-12-15T23:00:00Z",
  "total_capacity": 18000,
  "base_price": 150.00,
  "max_tickets_per_booking": 8
}
```
**Response:** Event created in `draft` status with `version: 1`

### Publish Event (Admin) 
```http
PUT /api/v1/admin/events/{event_id}
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "status": "published",
  "version": 1
}
```
**Purpose:** Make event visible to public APIs

### List Admin Events
```http
GET /api/v1/admin/events?status=published&page=1&limit=20
Authorization: Bearer <admin_token>
```

---

## 🌐 Public Event APIs

### List Published Events
```http
GET /api/v1/events?event_type=concert&city=New York&page=1&limit=10
```
**Response:**
```json
{
  "events": [
    {
      "event_id": "7204c97d-...",
      "name": "The Rolling Stones World Tour", 
      "venue_name": "Madison Square Garden",
      "available_seats": 17997,
      "base_price": 150,
      "status": "published"
    }
  ],
  "total": 3,
  "page": 1, 
  "limit": 10,
  "has_more": false
}
```

### Get Event Details
```http
GET /api/v1/events/{event_id}
```
**Response:** Complete event + venue information

### Check Real-time Availability
```http
GET /api/v1/events/{event_id}/availability
```
**Response:**
```json
{
  "available_seats": 17997,
  "status": "published", 
  "last_updated": "2025-09-14T15:33:58Z"
}
```
**Purpose:** Real-time seat availability for booking UI

---

## ⚙️ Internal Service APIs

**Required:** `X-API-Key: internal-service-communication-key-change-in-production`

### Get Event for Booking Validation
```http
GET /internal/events/{event_id}
X-API-Key: internal-service-communication-key-change-in-production
```
**Response:**
```json
{
  "event_id": "7204c97d-...",
  "available_seats": 17997,
  "max_tickets_per_booking": 8,
  "base_price": 150,
  "version": 4,
  "status": "published",
  "name": "The Rolling Stones World Tour"
}
```
**Purpose:** Booking Service validates event details before reserving seats

### Update Seat Availability (Critical)
```http
POST /internal/events/{event_id}/update-availability
X-API-Key: internal-service-communication-key-change-in-production
Content-Type: application/json

{
  "quantity": -5,  // Reduce by 5 seats (negative = reduce)
  "version": 4     // Current version for optimistic locking
}
```
**Response:**
```json
{
  "event_id": "7204c97d-...",
  "available_seats": 17992, 
  "status": "published",
  "version": 5
}
```
**Purpose:** Atomically reserve seats during booking process
**Note:** Uses optimistic locking to prevent overselling

### Return Seats (Cancellations)
```http
POST /internal/events/{event_id}/return-seats
X-API-Key: internal-service-communication-key-change-in-production
Content-Type: application/json

{
  "quantity": 2,   // Return 2 seats (positive = add back)
  "version": 5     // Current version
}
```
**Response:**
```json
{
  "event_id": "7204c97d-...",
  "available_seats": 17994,
  "status": "published", 
  "version": 6
}
```
**Purpose:** Return seats to pool after booking cancellation

---

## 🔄 Event Lifecycle

1. **Admin creates event** → `status: "draft"`, `version: 1`
2. **Admin publishes event** → `status: "published"` 
3. **Event appears in public APIs** → Available for discovery
4. **Booking Service reserves seats** → Reduces `available_seats`
5. **Booking cancellation** → Returns seats via return-seats API

## 🛡️ Concurrency Control

- **Optimistic Locking:** All seat updates require current `version` 
- **Atomic Operations:** Seat availability updates are database transactions
- **Race Condition Prevention:** Version mismatches return error, client must retry

## 🚨 Error Responses

```json
{
  "error": "Event was updated by another process. Please retry."
}
```
**Cause:** Version conflict in seat update
**Action:** Get latest event version and retry

## 📊 Event Status Types

- `draft` - Created but not public
- `published` - Live and bookable  
- `sold_out` - No available seats
- `cancelled` - Event cancelled

## 🧪 Quick Test Commands

```bash
# Create admin
curl -X POST http://localhost:8002/api/v1/auth/admin/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"test123","name":"Admin"}'

# List public events  
curl http://localhost:8002/api/v1/events

# Check availability
curl http://localhost:8002/api/v1/events/{event_id}/availability

# Internal: Reserve 2 seats
curl -X POST http://localhost:8002/internal/events/{event_id}/update-availability \
  -H "X-API-Key: internal-service-communication-key-change-in-production" \
  -H "Content-Type: application/json" \
  -d '{"quantity":-2,"version":1}'
```

## 🔗 Integration Points

- **Search Service:** Events auto-sync to Elasticsearch when created/updated
- **Booking Service:** Uses internal APIs for seat management
- **User Service:** Admin authentication uses same JWT infrastructure