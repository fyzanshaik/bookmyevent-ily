# Event Service - Complete API Specification

## Service Overview
The Event Service is the **single source of truth** for all event data in the Evently platform. It handles event management, venue operations, admin authentication, and provides critical internal APIs for the booking system with concurrency control.

## API Endpoints

### 🔐 Admin Authentication Endpoints
| Method | Endpoint | Description | Request Body | Response | Middleware |
|--------|----------|-------------|--------------|----------|------------|
| POST | `/api/v1/auth/admin/register` | Admin registration | `{email, password, name, phone_number?, role?}` | `{admin_id, email, name, access_token, refresh_token}` | None |
| POST | `/api/v1/auth/admin/login` | Admin login | `{email, password}` | `{admin_id, access_token, refresh_token, role, permissions}` | None |
| POST | `/api/v1/auth/admin/refresh` | Refresh admin token | `{refresh_token}` | `{access_token, refresh_token}` | None |

### 📅 Public Event Endpoints (High Traffic)
| Method | Endpoint | Description | Query Params | Response | Middleware |
|--------|----------|-------------|--------------|----------|------------|
| GET | `/api/v1/events` | List published events | `?page=1&limit=20&type=concert&city=NYC&date_from=2024-01-01&date_to=2024-12-31` | `{events: [...], total, page, limit, has_more}` | None |
| GET | `/api/v1/events/:id` | Get event details + venue | None | `{event_id, name, description, venue: {...}, datetime, available_seats, base_price, status}` | None |
| GET | `/api/v1/events/:id/availability` | Real-time availability | None | `{available_seats: 150, status: "published", last_updated: "..."}` | None |

### 🔧 Admin Event Management (Protected)
| Method | Endpoint | Description | Request Body | Response | Middleware |
|--------|----------|-------------|--------------|----------|------------|
| POST | `/api/v1/admin/events` | Create new event | `{name, description, venue_id, event_type, start_datetime, end_datetime, total_capacity, base_price, max_tickets_per_booking?}` | `{event_id, ...event_data, version: 1}` | RequireAdminAuth |
| PUT | `/api/v1/admin/events/:id` | Update event | `{name?, description?, ...partial_updates, version}` | `{event_id, ...updated_data, version: n+1}` | RequireAdminAuth, CheckOwnership |
| DELETE | `/api/v1/admin/events/:id` | Cancel event | `{version}` | `{message: "Event cancelled", event_id}` | RequireAdminAuth, CheckOwnership |
| GET | `/api/v1/admin/events` | List admin's events | `?page=1&limit=20&status=published` | `{events: [...], total, page, limit}` | RequireAdminAuth |
| GET | `/api/v1/admin/events/:id/analytics` | Event analytics | None | `{tickets_sold, revenue, capacity_utilization, booking_trends}` | RequireAdminAuth, CheckOwnership |

### 🏢 Admin Venue Management (Protected)
| Method | Endpoint | Description | Request Body | Response | Middleware |
|--------|----------|-------------|--------------|----------|------------|
| POST | `/api/v1/admin/venues` | Create venue | `{name, address, city, state?, country, postal_code?, capacity, layout_config?}` | `{venue_id, ...venue_data}` | RequireAdminAuth |
| GET | `/api/v1/admin/venues` | List venues | `?page=1&limit=20&city=NYC&search=madison` | `{venues: [...], total, page, limit}` | RequireAdminAuth |
| PUT | `/api/v1/admin/venues/:id` | Update venue | `{name?, address?, ...partial_updates}` | `{venue_id, ...updated_data}` | RequireAdminAuth |
| DELETE | `/api/v1/admin/venues/:id` | Delete venue | None | `{message: "Venue deleted"}` | RequireAdminAuth |

### ⚙️ Internal Service Endpoints (Service-to-Service)
| Method | Endpoint | Description | Request Body | Response | Middleware |
|--------|----------|-------------|--------------|----------|------------|
| POST | `/internal/events/:id/update-availability` | **CRITICAL**: Update seats atomically | `{quantity: -2, version: 42}` | `{event_id, available_seats, status, version}` | RequireInternalAuth |
| GET | `/internal/events/:id` | Get event for booking validation | None | `{event_id, available_seats, max_tickets_per_booking, base_price, version, status}` | RequireInternalAuth |
| POST | `/internal/events/:id/return-seats` | Return seats (cancellations) | `{quantity: 2, version: 43}` | `{event_id, available_seats, status, version}` | RequireInternalAuth |

### 💊 Health & Monitoring
| Method | Endpoint | Description | Response | Middleware |
|--------|----------|-------------|----------|------------|
| GET | `/healthz` | Basic health check | `{status: "healthy"}` | None |
| GET | `/health/ready` | Readiness probe | `{status: "ready", database: "connected"}` | None |

## Service Responsibilities

### Core Responsibilities
1. **Event Lifecycle Management** - CRUD operations with status transitions
2. **Venue Management** - Venue CRUD and event-venue associations  
3. **Admin Authentication** - JWT-based admin access control
4. **Concurrency Control** - Thread-safe seat updates with optimistic locking
5. **Anti-Overselling** - Atomic operations preventing negative seat counts
6. **Search Data Preparation** - Efficient queries for CDC to Elasticsearch

### Service Integration Points

**→ Outbound Calls:**
- **User Service**: `/internal/auth/verify` (verify admin tokens)

**← Inbound Calls:**
- **Booking Service**: `/internal/events/:id/update-availability` (CRITICAL PATH)
- **Booking Service**: `/internal/events/:id` (event validation)
- **Booking Service**: `/internal/events/:id/return-seats` (cancellations)
- **Search Service**: CDC pipeline sync (database-level)
- **Public Users**: Event browsing and details

## Error Handling Patterns

### Standard Error Response Format
```json
{
  "error": {
    "code": "INSUFFICIENT_SEATS",
    "message": "Not enough seats available", 
    "details": {
      "requested": 5,
      "available": 3,
      "event_id": "event-123"
    },
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### Critical Error Codes
- `CONCURRENT_UPDATE` - Version mismatch during optimistic locking
- `INSUFFICIENT_SEATS` - Not enough seats for booking request
- `EVENT_NOT_FOUND` - Event doesn't exist or not published
- `UNAUTHORIZED` - Invalid admin token or insufficient permissions
- `EVENT_NOT_MODIFIABLE` - Event cannot be modified (already started/cancelled)
- `VENUE_IN_USE` - Cannot delete venue with associated events

## Performance Requirements

### High-Traffic Endpoints
- `GET /api/v1/events` - **Target: <200ms p99, 1000+ RPS**
- `GET /api/v1/events/:id/availability` - **Target: <100ms p99, 2000+ RPS**
- `POST /internal/events/:id/update-availability` - **Target: <50ms p99, 500+ RPS**

### Concurrency Targets
- **1000+ concurrent booking requests** without overselling
- **Row-level lock duration: <10ms**
- **Version conflict retry success rate: >95%**

## Data Consistency

### Strong Consistency (Critical)
- Seat availability updates
- Event status transitions
- Admin authentication

### Eventual Consistency (Acceptable)
- Search index synchronization
- Analytics data aggregation
- Venue metadata updates

This specification ensures the Event Service can handle high-traffic ticket booking scenarios while maintaining data integrity and preventing overselling.