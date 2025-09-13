-- Initialize databases for all services

-- Create User Service database
CREATE DATABASE users_db;

-- Create Event Service database (for future use)
CREATE DATABASE events_db;

-- Create Booking Service database (for future use)
CREATE DATABASE bookings_db;

-- Grant permissions (optional, since we're using the postgres user)
GRANT ALL PRIVILEGES ON DATABASE users_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE events_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE bookings_db TO postgres;