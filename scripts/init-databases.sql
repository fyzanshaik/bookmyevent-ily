-- Initialize databases for all services

CREATE DATABASE users_db;
CREATE DATABASE events_db;
CREATE DATABASE bookings_db;

GRANT ALL PRIVILEGES ON DATABASE users_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE events_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE bookings_db TO postgres;