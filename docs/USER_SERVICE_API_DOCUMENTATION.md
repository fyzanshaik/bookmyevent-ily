# User Service API Documentation

## Overview

The User Service handles user authentication, profile management, and provides internal authentication services for other microservices. It runs on port **8001** and manages the `users_db` PostgreSQL database.

**Base URL:** `http://localhost:8001`

## Authentication & Authorization

### JWT Tokens
- **Access Token**: Short-lived (15 minutes), used for API authentication
- **Refresh Token**: Long-lived (7 days), used to obtain new access tokens
- **Internal API Key**: `X-API-Key` header required for internal endpoints

### Token Flow
1. Register/Login → Get Access + Refresh tokens
2. Use Access token in `Authorization: Bearer <token>` header
3. When Access token expires → Use Refresh token to get new tokens
4. Logout → Revoke Refresh token



---

## Public API Endpoints

### 🔐 Authentication Endpoints

#### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123", 
  "name": "John Doe",
  "phone_number": "+1234567890"  // Optional
}
```

**Response (200 OK):**
```json
{
  "user_id": "46137163-dc79-48ae-a246-d1ec7ad06af8",
  "email": "user@example.com", 
  "name": "John Doe",
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "a3b8e8240ed28601c0fc..."
}
```

**Errors:**
- `400`: Invalid JSON, missing required fields
- `500`: Email already exists, database error

---

#### Login User
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response (200 OK):**r
```json
{
  "user_id": "46137163-dc79-48ae-a246-d1ec7ad06af8",
  "email": "user@example.com",
  "name": "John Doe", 
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "bf203cfa3266afeb4085..."
}
```

**Errors:**
- `400`: Invalid JSON, missing fields
- `401`: Incorrect email or password

---

#### Refresh Access Token
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "bf203cfa3266afeb4085670cd09d20f28f0d2b04933cce67864a8cdd26d2c2e2"
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "f3cef21577b7771da143..."
}
```

**Errors:**
- `400`: Invalid JSON, missing refresh token
- `401`: Invalid or expired refresh token

**Notes:**
- Old refresh token is revoked and replaced with new one
- Both tokens are rotated for security

---

#### Logout User
```http
POST /api/v1/auth/logout
Content-Type: application/json

{
  "refresh_token": "f3cef21577b7771da143103489a3d7b41da31eb812b4796aa833702eefe6c6df"
}
```

**Response (200 OK):**
```json
{
  "message": "success"
}
```

**Errors:**
- `400`: Invalid JSON, missing refresh token
- `500`: Failed to revoke token

**Notes:**
- Access tokens remain valid until expiry (15 minutes)
- Refresh token is immediately revoked

---

### 👤 Profile Management (Protected)

All profile endpoints require `Authorization: Bearer <access_token>` header.

#### Get User Profile
```http
GET /api/v1/users/profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response (200 OK):**
```json
{
  "user_id": "46137163-dc79-48ae-a246-d1ec7ad06af8",
  "email": "user@example.com",
  "name": "John Doe",
  "phone_number": "+1234567890",
  "created_at": "2025-09-14T14:55:01.506544Z"
}
```

**Errors:**
- `401`: Missing or invalid authorization header, expired token

---

#### Update User Profile
```http
PUT /api/v1/users/profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "name": "Jane Smith",           // Optional
  "phone_number": "+9876543210"  // Optional
}
```

**Response (200 OK):**
```json
{
  "user_id": "46137163-dc79-48ae-a246-d1ec7ad06af8",
  "email": "user@example.com",
  "name": "Jane Smith",
  "phone_number": "+9876543210", 
  "created_at": "2025-09-14T14:55:01.506544Z"
}
```

**Errors:**
- `400`: Invalid JSON
- `401`: Missing or invalid authorization header
- `500`: Database error

**Notes:**
- Email cannot be updated
- Only provided fields are updated
- Phone number can be set to empty string to remove

---

#### Get User Bookings
```http
GET /api/v1/users/bookings
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response (200 OK):**
```json
[
  {
    "booking_id": "uuid",
    "event_name": "Concert Name",
    "date": "2025-12-25T19:00:00Z",
    "quantity": 2,
    "status": "confirmed"
  }
]
```

**Notes:**
- Returns empty array `[]` if no bookings
- Actual implementation fetches from Booking Service

---

## Internal Service Endpoints

These endpoints are used by other microservices and require `X-API-Key` header.

### 🔍 Token Verification
```http
POST /internal/auth/verify
X-API-Key: internal-service-communication-key-change-in-production
Content-Type: application/json

{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response (200 OK):**
```json
{
  "user_id": "46137163-dc79-48ae-a246-d1ec7ad06af8",
  "email": "user@example.com",
  "valid": true
}
```

**Errors:**
- `400`: Invalid JSON, missing token
- `401`: Missing or invalid API key, invalid token

---

### 👤 Get User Details
```http
GET /internal/users/{userId}
X-API-Key: internal-service-communication-key-change-in-production
```

**Response (200 OK):**
```json
{
  "user_id": "46137163-dc79-48ae-a246-d1ec7ad06af8",
  "email": "user@example.com",
  "name": "John Doe",
  "phone_number": "+1234567890",
  "created_at": "2025-09-14T14:55:01.506544Z"
}
```

**Errors:**
- `400`: Invalid user ID format
- `401`: Missing or invalid API key  
- `404`: User not found

---

## Health Check Endpoints

### Basic Health Check
```http
GET /healthz
```

**Response (200 OK):**
```json
{
  "status": "healthy"
}
```

### Readiness Check
```http
GET /health/ready
```

**Response (200 OK):**
```json
{
  "status": "ready",
  "database": "connected"
}
```

---

## Error Response Format

All errors follow this standard format:

```json
{
  "error": "Human readable error message"
}
```

### HTTP Status Codes
- **200**: Success
- **201**: Created (not used in User Service)
- **400**: Bad Request - Invalid input, validation errors
- **401**: Unauthorized - Authentication required, invalid token
- **403**: Forbidden (not used in User Service)
- **404**: Not Found - User not found
- **500**: Internal Server Error - Database errors, system failures

---

## Data Models

### User
```typescript
interface User {
  user_id: string;      // UUID
  email: string;        // Unique, requiredr
  name: string;         // Required
  phone_number?: string; // Optional
  created_at: string;   // ISO timestamp
}
```

### Request Models
```typescript
interface CreateUserRequest {
  email: string;
  password: string;
  name: string;
  phone_number?: string;
}r

interface UserLoginRequest {
  email: string;
  password: string;
}

interface RefreshTokenRequest {
  refresh_token: string;
}

interface LogoutRequest {
  refresh_token: string;
}

interface UpdateUserRequest {r
  name?: string;
  phone_number?: string;
}

interface VerifyTokenRequest {
  token: string;
}
```

### Response Models  
```typescript
interface AuthResponse {
  user_id: string;
  email: string;
  name: string;
  access_token: string;
  refresh_token: string;
}

interface RefreshTokenResponse {
  access_token: string;
  refresh_token: string;
}

interface LogoutResponse {
  message: "success";
}

interface TokenVerificationResponse {
  user_id: string;
  email: string;
  valid: true;
}
```

---

## Testing Examples

### Complete Authentication Flow
```bash
# 1. Register
curl -X POST http://localhost:8001/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123","name":"Test User"}'

# 2. Login  
curl -X POST http://localhost:8001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123"}'

# 3. Use token for profile
curl -H "Authorization: Bearer <access_token>" \
  http://localhost:8001/api/v1/users/profile

# 4. Refresh token
curl -X POST http://localhost:8001/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'

# 5. Logout
curl -X POST http://localhost:8001/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

### Internal Service Usage
```bash
# Verify token from another service
curl -X POST http://localhost:8001/internal/auth/verify \
  -H "X-API-Key: internal-service-communication-key-change-in-production" \
  -H "Content-Type: application/json" \
  -d '{"token":"<jwt_token>"}'

# Get user details
curl -H "X-API-Key: internal-service-communication-key-change-in-production" \
  http://localhost:8001/internal/users/46137163-dc79-48ae-a246-d1ec7ad06af8
```

---

## Recent Fixes Applied ✅

1. **Token Refresh Fixed**: Now accepts refresh token in request body instead of Authorization header
2. **Logout Fixed**: Now accepts refresh token in request body instead of Authorization header  
3. **Internal Token Verification**: Working correctly - validates JWT and returns user info
4. **All Endpoints Tested**: Comprehensive testing completed with various scenarios

## Database

- **Database**: `users_db` on PostgreSQL port 5434
- **Tables**: `users`, `refresh_tokens`, `goose_db_version`
- **Connection**: Managed via SQLC generated code
- **Migrations**: Located in `migrations/user-service/`

## Configuration

Key environment variables:
- `USER_SERVICE_PORT=8001`
- `USER_SERVICE_DB_URL` - PostgreSQL connection
- `JWT_SECRET` - Token signing key
- `INTERNAL_API_KEY` - Service-to-service auth key