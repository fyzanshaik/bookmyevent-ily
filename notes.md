# Start containers

docker-compose up -d

# Install tools

go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/pressly/goose/v3/cmd/goose@latest

# Run migrations

goose -dir sql/schema postgres "postgresql://evently:evently123@localhost:5432/evently?sslmode=disable" up

# Generate sqlc code

sqlc generate

Expected structure :
evently/
├── cmd/
│ ├── api/
│ │ ├── main.go
│ │ └── config.go
│ └── worker/
│ ├── main.go
│ └── config.go
├── internal/
│ ├── database/ # sqlc generated + connection
│ │ ├── db.go # Connection logic
│ │ ├── models.go # sqlc generated
│ │ └── \*.sql.go # sqlc generated queries
│ ├── booking/ # We'll build this as needed
│ └── redis/ # Redis client wrapper
├── api/
│ ├── handlers/
│ ├── middleware/
│ └── server.go
├── worker/
│ ├── handlers/
│ ├── jobs/
│ └── server.go
├── pkg/
│ ├── utils/
│ └── constants/
├── sql/
│ ├── schema/ # Goose migrations
│ └── queries/ # SQLC queries
├── docker-compose.yml
├── sqlc.yaml
└── go.mod

User buying process:

- Reserve tickets by saving it inside a redis cluster and updating my read database for events and mark them as "reserved" this should only happen when payment process has been started
- Set a key value pair for the tickets, with userId and eventId
- anytime someone wants to book a ticket we can check this redis cluster and see if its being processed or its due
- user once successfully pays then we send a write request to update that particular row now row level locking can be done for further efficiency
  Maybe optimistic concurrent process could be done? what if both the users are trying to access the available seat at the same time, here optimistically we could say whoever gets first changes the seat status is the one about to proceed with payment gateway.

Functional Requirements

Functional requirements define what the system must do. For Evently, these are the core features and actions users and admins can perform.

User Features
Browse Events: Users can view a list of upcoming events with all relevant details: name, venue, time, and capacity.

Book Tickets: Users can book tickets for an event. The system must update the available ticket count correctly and prevent overselling.

Cancel Tickets: Users can cancel their booked tickets, which should then free up the seats for other users.

View History: Users can access a history of their past and upcoming bookings.

Admin Features
Manage Events: Admins can create, update, and manage the details of events.

View Analytics: Admins can view booking statistics, including total bookings, the most popular events, and how much of an event's capacity has been utilized.

Optional Enhancements (Stretch Goals)
Waitlist System: Users can join a waitlist for an event that is full.

Seat-level Booking: Users can choose and book specific seats within a venue, rather than just a general ticket.

Notifications: The system can inform users (e.g., via email or push notification) if a spot opens up on the waitlist.

Advanced Analytics: Provide more detailed analytics for admins, such as cancellation rates and daily booking statistics.

Creative Features: Any out-of-the-box ideas that improve the user or admin experience.

Non-Functional Requirements
Non-functional requirements describe how the system performs a function. These focus on quality attributes like performance, scalability, and reliability.

Concurrency & Race Conditions
The system must handle simultaneous booking requests without overselling tickets. This is a critical requirement due to the high-traffic nature of ticket sales. The solution should use techniques like optimistic locking, database transactions, or a message queue to ensure data integrity.

Scalability
The system must be designed to handle peak traffic, specifically thousands of concurrent booking requests. Solutions should include caching for read-heavy operations (like browsing events), database indexing to speed up queries, and possibly sharding to distribute the database load.

API Quality
The APIs must be clean, RESTful, and easy to use. They should include proper error handling with clear status codes to communicate success or failure to the client.

Performance
The system needs to respond quickly, especially during peak load times. The design must minimize latency to provide a good user experience.

Main Processes and Expectations
This section outlines the core workflows and what is expected in the final deliverables.

Core Processes
Booking Process: A user requests a ticket. The system first checks for availability, then securely decrements the available ticket count, creates a new booking record, and confirms the booking. All these steps must happen within a single, atomic operation to prevent race conditions.

Cancellation Process: A user requests a cancellation. The system validates the booking, marks it as canceled, and increments the available ticket count for that event. If a waitlist is implemented, a notification may be sent to the first person on the list.

Event Management: An admin uses the system to add new events, modify existing ones (e.g., changing the time or venue), and view event-specific booking data.

### Expected Deliverables

- **Working Backend:** A live, deployed backend application that demonstrates all the core features and any enhancements.
- **Code Repository:** A clean, well-documented GitHub repository with the project code.
- **Design Diagrams:**
  - **High-Level Architecture Diagram:** A visual representation of the main components of the system (APIs, database, cache, etc.).
  - **Entity-Relationship (ER) Diagram:** A diagram showing the relationships between key data entities like users, events, and bookings. This should also represent any new entities for stretch goals (e.g., a `Waitlist` or `Seats` table).
- **Documentation:** A written explanation of the major design decisions, trade-offs, how scalability was addressed, and API documentation. This should also cover any creative features or optimizations.
- **Video Walkthrough:** A short video demonstrating the system in action and explaining the design choices and challenges faced.

---------------------------------------------------------------------------------------------------------->
DOCKER COMPOSE
In your docker-compose.yml file, you define services. A service is a blueprint for a container.

1 services:
2 postgres: # <--- This is the SERVICE name
3 image: postgres:15-alpine
4 container_name: evently_postgres # <--- This is the CONTAINER name
5 # ... other configurations

Think of it this way:

1.  `postgres` (The Service Name): This is the logical name you give to a component of your
    application stack within Docker Compose. When you use a docker compose command (like exec,
    logs, up, down), you are telling Compose which service blueprint you want to interact with.

2.  `evently_postgres` (The Container Name): This is the actual, specific name given to the
    running container instance created from the postgres service blueprint. If you were to use
    the base docker command (not docker compose), you would use this name (e.g., docker stop
    evently_postgres).

In short: `docker compose` commands manage _services_, not individual containers directly.

So, when you run docker compose exec postgres ..., you are telling Docker Compose: "Find the
service named postgres in my configuration file, and run a command inside the container it
manages."

CHECKING USER DB:
docker compose exec postgres psql -U postgres -d users_db
