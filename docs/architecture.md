# Evently - Event Booking Platform Technical Specification

## System Overview

Evently is a scalable event booking platform designed to handle high-traffic ticket sales with support for concurrent bookings, waitlist management, and real-time availability tracking. The system can handle 100,000 concurrent requests while maintaining data consistency and preventing overselling.

## Architecture Overview

### High-Level Architecture

```
┌─────────────┐
│   Clients   │
└──────┬──────┘
       │
┌──────▼──────────────────────────────────────┐
│         NGINX Load Balancer                 │
│    (Rate Limiting, SSL Termination)         │
└──────┬──────────────────────────────────────┘
       │
┌──────▼──────────────────────────────────────┐
│            API Gateway Layer                 │
│     (Authentication, Routing, Monitoring)    │
└───┬────────┬────────┬────────┬────────┬────┘
    │        │        │        │        │
┌───▼──┐ ┌──▼──┐ ┌──▼──┐ ┌──▼──┐ ┌──▼──┐
│ User │ │Event│ │Search│ │Book │ │Anal │
│ Svc  │ │ Svc │ │ Svc  │ │ Svc │ │ Svc │
└───┬──┘ └──┬──┘ └──┬──┘ └──┬──┘ └──┬──┘
    │       │       │       │       │
┌───▼───────▼───────▼───────▼───────▼────┐
│         Shared Infrastructure           │
├──────────────────────────────────────────┤
│ PostgreSQL │ Redis │ Elasticsearch │    │
│  (Master)  │Cluster│    Cluster     │    │
│     +      │       │                │    │
│  Replicas  │       │                │    │
└────────────┴───────┴────────────────┴────┘
       │                      ▲
       │     Data ingestion Pipeline │
       └──────────────────────┘

```

## Database Architecture

### Database Distribution

| Service           | Database Access                          | Tables                                      | Purpose                                 |
| ----------------- | ---------------------------------------- | ------------------------------------------- | --------------------------------------- |
| User Service      | PostgreSQL (DB: users_db)                | users, refresh_tokens                       | Authentication & user management        |
| Event Service     | PostgreSQL (DB: events_db)               | events, venues, admins                      | Event management, source of truth       |
| Booking Service   | PostgreSQL (DB: bookings_db)             | bookings, waitlist, booking_seats, payments | Booking transactions & payment tracking |
| Search Service    | Elasticsearch + Redis                    | events index (read-only)                    | Full-text search & caching              |
| Analytics Service | PostgreSQL (read-only access to all DBs) | All tables (read-only)                      | Reporting & analytics                   |

### Master-Replica Configuration

Each PostgreSQL database will have:

- 1 Master (Write operations)
- 2 Read Replicas (Read operations)
- PgBouncer for connection pooling

## Complete Database Schema

### User Service Database (users_db)

```sql
-- Users table
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone_number VARCHAR(20),
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Refresh tokens table
CREATE TABLE refresh_tokens (
    token TEXT PRIMARY KEY,
    user_id UUID NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user
        FOREIGN KEY(user_id)
        REFERENCES users(user_id)
        ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone_number);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires ON refresh_tokens(expires_at) WHERE revoked_at IS NULL;
```

### Event Service Database (events_db)

```sql
-- Venues table
CREATE TABLE venues (
    venue_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    address TEXT NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    country VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20),
    capacity INTEGER NOT NULL,
    layout_config JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Events table
CREATE TABLE events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    venue_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL, -- concert, sports, theater, conference
    start_datetime TIMESTAMP NOT NULL,
    end_datetime TIMESTAMP NOT NULL,
    total_capacity INTEGER NOT NULL,
    available_seats INTEGER NOT NULL,
    base_price DECIMAL(10, 2) NOT NULL,
    max_tickets_per_booking INTEGER DEFAULT 10,
    is_high_traffic BOOLEAN DEFAULT false,
    status VARCHAR(20) DEFAULT 'draft', -- draft, published, sold_out, cancelled
    age_group VARCHAR(50),
    language VARCHAR(50),
    created_by UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_venue
        FOREIGN KEY(venue_id)
        REFERENCES venues(venue_id),
    CONSTRAINT check_seats
        CHECK (available_seats >= 0 AND available_seats <= total_capacity),
    CONSTRAINT check_dates
        CHECK (end_datetime > start_datetime)
);

-- Admins table
CREATE TABLE admins (
    admin_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20),
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'event_manager',
    permissions JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_events_available ON events(available_seats) WHERE available_seats > 0 AND status = 'published';
CREATE INDEX idx_events_datetime ON events(start_datetime, end_datetime);
CREATE INDEX idx_events_venue ON events(venue_id);
CREATE INDEX idx_events_type ON events(event_type);
CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_venues_city ON venues(city);
CREATE INDEX idx_admins_email ON admins(email);
```

### Booking Service Database (bookings_db)

```sql
-- Bookings table
CREATE TABLE bookings (
    booking_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    event_id UUID NOT NULL,
    booking_reference VARCHAR(20) UNIQUE NOT NULL, -- Human-readable reference
    quantity INTEGER NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    status VARCHAR(20) NOT NULL, -- pending, confirmed, cancelled, expired
    payment_status VARCHAR(20) NOT NULL, -- pending, completed, failed, refunded
    idempotency_key VARCHAR(255) UNIQUE,
    booked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    confirmed_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_quantity CHECK (quantity > 0)
);

-- Payments table
CREATE TABLE payments (
    payment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL,
    user_id UUID NOT NULL,
    event_id UUID NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    payment_method VARCHAR(50), -- mock_payment, credit_card, debit_card
    payment_gateway VARCHAR(50), -- mock_gateway
    gateway_transaction_id VARCHAR(255),
    status VARCHAR(20) NOT NULL, -- initiated, processing, completed, failed
    ticket_url TEXT, -- Generated ticket URL
    initiated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    failed_at TIMESTAMP,
    error_message TEXT,
    CONSTRAINT fk_booking
        FOREIGN KEY(booking_id)
        REFERENCES bookings(booking_id)
);

-- Waitlist table
CREATE TABLE waitlist (
    waitlist_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL,
    user_id UUID NOT NULL,
    position INTEGER NOT NULL,
    quantity_requested INTEGER NOT NULL,
    status VARCHAR(20) DEFAULT 'waiting', -- waiting, offered, expired, converted
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    offered_at TIMESTAMP,
    expires_at TIMESTAMP,
    converted_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_user_event
        UNIQUE(user_id, event_id),
    CONSTRAINT check_quantity
        CHECK (quantity_requested > 0)
);

-- Booking seats (for future seat-level booking)
CREATE TABLE booking_seats (
    booking_seat_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL,
    seat_number VARCHAR(10),
    seat_row VARCHAR(10),
    seat_section VARCHAR(50),
    status VARCHAR(20) DEFAULT 'booked',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_booking
        FOREIGN KEY(booking_id)
        REFERENCES bookings(booking_id)
        ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_bookings_user ON bookings(user_id, status);
CREATE INDEX idx_bookings_event ON bookings(event_id, status);
CREATE INDEX idx_bookings_reference ON bookings(booking_reference);
CREATE INDEX idx_bookings_idempotency ON bookings(idempotency_key);
CREATE INDEX idx_payments_booking ON payments(booking_id);
CREATE INDEX idx_payments_user ON payments(user_id);
CREATE INDEX idx_waitlist_event_position ON waitlist(event_id, position) WHERE status = 'waiting';
CREATE INDEX idx_waitlist_user ON waitlist(user_id);
```

## Service Specifications

### 1. User Service

**Responsibilities:**

- User authentication and authorization
- JWT token management
- User profile management
- Internal authentication verification for other services

**API Endpoints:**

| Method | Endpoint                 | Description          | Request Body                            | Response                                              |
| ------ | ------------------------ | -------------------- | --------------------------------------- | ----------------------------------------------------- |
| POST   | `/api/v1/auth/register`  | Register new user    | `{email, password, name, phone_number}` | `{user_id, email, name, access_token, refresh_token}` |
| POST   | `/api/v1/auth/login`     | User login           | `{email, password}`                     | `{user_id, access_token, refresh_token}`              |
| POST   | `/api/v1/auth/refresh`   | Refresh access token | `{refresh_token}`                       | `{access_token, refresh_token}`                       |
| POST   | `/api/v1/auth/logout`    | Logout user          | `{refresh_token}`                       | `{message: "success"}`                                |
| GET    | `/api/v1/users/profile`  | Get user profile     | -                                       | `{user_id, email, name, phone_number, created_at}`    |
| PUT    | `/api/v1/users/profile`  | Update user profile  | `{name?, phone_number?}`                | `{user_id, email, name, phone_number}`                |
| GET    | `/api/v1/users/bookings` | Get user bookings    | -                                       | `[{booking_id, event_name, date, quantity, status}]`  |

**Internal Endpoints (Service-to-Service):**
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/internal/auth/verify` | Verify JWT token validity |
| GET | `/internal/users/:userId` | Get user details for other services |

### 2. Event Service

**Responsibilities:**

- Event CRUD operations
- Venue management
- Real-time availability tracking
- Admin operations

**API Endpoints:**

| Method | Endpoint                          | Description            | Request Body                                | Response                                                    |
| ------ | --------------------------------- | ---------------------- | ------------------------------------------- | ----------------------------------------------------------- |
| GET    | `/api/v1/events`                  | List all events        | Query: `?page=1&limit=20&city=&type=&date=` | `{events: [...], total, page, limit}`                       |
| GET    | `/api/v1/events/:id`              | Get event details      | -                                           | `{event_id, name, venue, datetime, available_seats, price}` |
| GET    | `/api/v1/events/:id/availability` | Real-time availability | -                                           | `{available_seats, status}`                                 |

**Admin Endpoints:**
| Method | Endpoint | Description | Request Body | Response |
|--------|----------|-------------|--------------|----------|
| POST | `/api/v1/admin/events` | Create event | `{name, description, venue_id, event_type, start_datetime, end_datetime, total_capacity, base_price}` | `{event_id, ...eventData}` |
| PUT | `/api/v1/admin/events/:id` | Update event | `{partial event data}` | `{event_id, ...updatedData}` |
| DELETE | `/api/v1/admin/events/:id` | Delete event | - | `{message: "success"}` |
| GET | `/api/v1/admin/events/:id/analytics` | Event analytics | - | `{total_bookings, revenue, capacity_utilization}` |

**Internal Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/internal/events/:id/update-availability` | Update available seats (called by Booking Service) |
| GET | `/internal/events/:id` | Get event details for booking validation |

### 3. Search Service

**Responsibilities:**

- Full-text search on events
- Filtering and sorting
- Search suggestions
- Caching search results

**API Endpoints:**

| Method | Endpoint                     | Description       | Request Body                                                                      | Response                                   |
| ------ | ---------------------------- | ----------------- | --------------------------------------------------------------------------------- | ------------------------------------------ |
| GET    | `/api/v1/search`             | Search events     | Query: `?q=concert&city=NYC&type=music&date_from=&date_to=&min_price=&max_price=` | `{results: [...], total, facets: {}}`      |
| GET    | `/api/v1/search/suggestions` | Autocomplete      | Query: `?q=con`                                                                   | `{suggestions: ["concert", "conference"]}` |
| GET    | `/api/v1/search/filters`     | Available filters | -                                                                                 | `{cities: [], types: [], price_range: {}}` |
| GET    | `/api/v1/search/trending`    | Trending events   | -                                                                                 | `{events: [...]}`                          |

### 4. Booking Service

**Responsibilities:**

- Ticket reservation and booking
- Payment processing (mock)
- Booking cancellation
- Waitlist management
- Maintaining booking consistency

**API Endpoints:**

| Method | Endpoint                              | Description              | Request Body                            | Response                                      |
| ------ | ------------------------------------- | ------------------------ | --------------------------------------- | --------------------------------------------- |
| POST   | `/api/v1/bookings/check-availability` | Check if seats available | `{event_id, quantity}`                  | `{available: true/false, available_seats}`    |
| POST   | `/api/v1/bookings/reserve`            | Reserve seats (Step 1)   | `{event_id, quantity, idempotency_key}` | `{reservation_id, expires_at, amount}`        |
| POST   | `/api/v1/bookings/confirm`            | Confirm booking (Step 2) | `{reservation_id, payment_token}`       | `{booking_id, booking_reference, ticket_url}` |
| GET    | `/api/v1/bookings/:id`                | Get booking details      | -                                       | `{booking details}`                           |
| DELETE | `/api/v1/bookings/:id`                | Cancel booking           | -                                       | `{message: "success", refund_status}`         |
| GET    | `/api/v1/bookings/user/:userId`       | User booking history     | -                                       | `{bookings: [...]}`                           |

**Waitlist Endpoints:**
| Method | Endpoint | Description | Request Body | Response |
|--------|----------|-------------|--------------|----------|
| POST | `/api/v1/waitlist/join` | Join waitlist | `{event_id, quantity}` | `{waitlist_id, position, estimated_wait}` |
| GET | `/api/v1/waitlist/position` | Check position | Query: `?event_id=` | `{position, status}` |
| DELETE | `/api/v1/waitlist/leave` | Leave waitlist | `{event_id}` | `{message: "success"}` |

### 5. Analytics Service

**Responsibilities:**

- Generate reports and analytics
- Revenue tracking
- Booking statistics
- Performance metrics

**API Endpoints:**

| Method | Endpoint                                       | Description          | Request Body                    | Response                                           |
| ------ | ---------------------------------------------- | -------------------- | ------------------------------- | -------------------------------------------------- |
| GET    | `/api/v1/admin/analytics/revenue`              | Revenue report       | Query: `?start_date=&end_date=` | `{total_revenue, daily_breakdown, top_events}`     |
| GET    | `/api/v1/admin/analytics/events/popular`       | Popular events       | Query: `?limit=10`              | `{events: [...]}`                                  |
| GET    | `/api/v1/admin/analytics/bookings/stats`       | Booking statistics   | -                               | `{total_bookings, success_rate, avg_booking_size}` |
| GET    | `/api/v1/admin/analytics/cancellation-rate`    | Cancellation metrics | -                               | `{rate, reasons, trends}`                          |
| GET    | `/api/v1/admin/analytics/capacity-utilization` | Venue utilization    | -                               | `{venues: [{venue_id, utilization_rate}]}`         |

## Core Business Flows

### 1. User Registration Flow

```
1. Client → POST /api/v1/auth/register
2. User Service:
   - Validate email uniqueness
   - Hash password (bcrypt)
   - Create user record
   - Generate JWT + Refresh Token
   - Store refresh token in database
3. Return tokens to client
```

### 2. Event Browsing Flow

```
1. Client → GET /api/v1/search?q=concert
2. Search Service:
   - Check Redis cache
   - If miss: Query Elasticsearch
   - Apply filters and sorting
   - Cache results (TTL: 60s)
3. Return paginated results
```

### 3. Booking Flow (Critical Path)

```
Phase 1: Reservation
1. Client → POST /api/v1/bookings/reserve
2. Booking Service:
   - Verify user authentication (internal call to User Service)
   - Check event details (internal call to Event Service)
   - BEGIN TRANSACTION
   - Atomic update: UPDATE events SET available_seats = available_seats - :quantity
   - Create booking record (status: 'pending')
   - Store reservation in Redis (TTL: 300s)
   - COMMIT TRANSACTION
3. Return reservation_id with 5-minute expiry

Phase 2: Payment (Mock)
4. Client → POST /api/v1/bookings/confirm
5. Booking Service:
   - Validate reservation exists in Redis
   - Mock payment processing
   - Update booking status to 'confirmed'
   - Create payment record
   - Generate ticket URL
   - Remove reservation from Redis
6. Return booking confirmation with ticket

Expiry Handling:
- Background job runs every 60 seconds
- Find expired bookings (status='pending' AND expires_at < NOW)
- Return seats to pool
- Update booking status to 'expired'
```

### 4. Cancellation Flow

```
1. Client → DELETE /api/v1/bookings/:id
2. Booking Service:
   - Verify booking ownership
   - Check if event hasn't ended
   - BEGIN TRANSACTION
   - Update booking status to 'cancelled'
   - Return seats to event pool
   - Create refund record (mock)
   - COMMIT TRANSACTION
3. If waitlist exists: Trigger waitlist processing
4. Return cancellation confirmation
```

### 5. Waitlist Flow

```
Join Waitlist:
1. Client → POST /api/v1/waitlist/join
2. Booking Service:
   - Check if event is sold out
   - Add user to waitlist with position
   - Store in database
3. Return position and estimated wait

Process Waitlist (Background Job):
1. Triggered when seats become available
2. Get next user in waitlist
3. Send notification (expires in 2 minutes)
4. Update waitlist entry status to 'offered'
5. User has exclusive booking window
```

## Data Synchronization

### PostgreSQL → Elasticsearch CDC Pipeline

```
Architecture:
PostgreSQL (events_db) → Debezium → Kafka → Kafka Connect → Elasticsearch

Configuration:
1. Debezium monitors PostgreSQL WAL
2. Captures changes to 'events' and 'venues' tables
3. Publishes to Kafka topics: 'events.public.events', 'events.public.venues'
4. Kafka Connect Elasticsearch Sink consumes and indexes

Kafka Topics:
- events.public.events (partitions: 3, replication: 2)
- events.public.venues (partitions: 3, replication: 2)
```

## Infrastructure Components

### Container Services

```yaml
Services:
  # Databases
  - postgres-master (users_db, events_db, bookings_db)
  - postgres-replica-1
  - postgres-replica-2
  - pgbouncer (connection pooling)

  # Caching & Search
  - redis-master
  - redis-replica
  - elasticsearch-node-1
  - elasticsearch-node-2
  - elasticsearch-node-3

  # Message Queue
  - kafka-broker-1
  - kafka-broker-2
  - zookeeper
  - kafka-connect
  - debezium

  # Load Balancing
  - nginx

  # Application Services
  - user-service (3 instances)
  - event-service (3 instances)
  - search-service (3 instances)
  - booking-service (5 instances)
  - analytics-service (2 instances)

  # Monitoring
  - prometheus
  - grafana
  - elasticsearch-apm
```

### Redis Configuration

```
Use Cases:
1. Session Storage (User Service)
   - Key: session:{session_id}
   - TTL: 24 hours

2. Booking Reservations (Booking Service)
   - Key: reservation:{reservation_id}
   - Value: {user_id, event_id, quantity, expires_at}
   - TTL: 300 seconds

3. Search Cache (Search Service)
   - Key: search:{query_hash}
   - TTL: 60 seconds

4. Rate Limiting (API Gateway)
   - Key: rate_limit:{user_id}:{endpoint}
   - TTL: 60 seconds

5. Distributed Locks (Booking Service)
   - Key: lock:event:{event_id}
   - TTL: 10 seconds
```

## Security & Performance

### Authentication Flow

```
1. JWT Access Token (15 minutes expiry)
2. Refresh Token (7 days expiry)
3. Service-to-Service: Internal API keys or mTLS
```

### Rate Limiting

```
Per User:
- Search/Browse: 100 requests/minute
- Booking: 10 requests/minute
- Authentication: 5 requests/minute

Global:
- Per endpoint: 10,000 requests/second
```

### Database Optimization

```
Connection Pools:
- Master: 100 connections
- Each Replica: 200 connections
- PgBouncer: 1000 client connections

Query Optimization:
- Use prepared statements
- Batch operations where possible
- Implement query result caching
```

### Monitoring Metrics

```
Key Metrics:
- Booking success rate (target: >95%)
- API latency p99 (target: <500ms)
- Database replication lag (target: <1s)
- Search response time (target: <200ms)
- Concurrent users supported: 100,000
```

## Error Handling

### Standard Error Response Format

```json
{
  "error": {
    "code": "INSUFFICIENT_SEATS",
    "message": "Not enough seats available",
    "details": {
      "requested": 5,
      "available": 3
    },
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "req_123456"
  }
}
```

### HTTP Status Codes

- 200: Success
- 201: Created
- 400: Bad Request
- 401: Unauthorized
- 403: Forbidden
- 404: Not Found
- 409: Conflict (e.g., seats no longer available)
- 422: Unprocessable Entity
- 429: Too Many Requests
- 500: Internal Server Error
- 503: Service Unavailable

## Deployment Strategy

### Environment Variables

```
Each service requires:
- DATABASE_URL (master)
- DATABASE_REPLICA_URL
- REDIS_URL
- ELASTICSEARCH_URL
- KAFKA_BROKERS
- JWT_SECRET
- SERVICE_API_KEY
- LOG_LEVEL
- PORT
```

### Health Checks

```
Each service exposes:
- /health - Basic health check
- /health/live - Liveness probe
- /health/ready - Readiness probe (includes dependency checks)
```

## Implementation Priority

1. **Phase 1: Core Foundation**
   - User Service (authentication)
   - Event Service (CRUD)
   - Basic database setup

2. **Phase 2: Booking System**
   - Booking Service (reservation + confirmation)
   - Payment mock
   - Redis integration

3. **Phase 3: Search & Discovery**
   - Search Service
   - Elasticsearch integration
   - CDC pipeline setup

4. **Phase 4: Advanced Features**
   - Waitlist functionality
   - Analytics Service
   - Admin dashboard

5. **Phase 5: Production Readiness**
   - Monitoring setup
   - Load testing
   - Performance optimization
   - Security hardening

This architecture ensures:

- **No overselling** through atomic operations
- **High availability** through service redundancy
- **Scalability** to 100k concurrent users
- **Data consistency** across services
- **Fault tolerance** through circuit breakers and retries
