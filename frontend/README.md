# BookMyEvent Frontend

A React-based frontend for the BookMyEvent platform - a comprehensive event booking system.

## Features

### User Features
- **User Authentication**: Register, login, profile management
- **Event Discovery**: Search and filter events with real-time results
- **Booking System**: Two-phase booking (reserve → payment → confirm)
- **Dashboard**: View booking history and manage profile
- **Waitlist**: Join waitlists for sold-out events

### Admin Features
- **Admin Authentication**: Separate admin login system
- **Venue Management**: Create, edit, and manage venues
- **Event Management**: Create, publish, and manage events
- **Dashboard**: Overview of events and venues

## Tech Stack

- **React 19** - Frontend framework
- **React Router** - Client-side routing
- **Tailwind CSS** - Styling framework
- **Axios** - HTTP client
- **Lucide React** - Icon library
- **Vite** - Build tool

## Getting Started

### Prerequisites

- Node.js 18+ or Bun
- Backend services running (User, Event, Search, Booking services)

### Installation

```bash
# Install dependencies
bun install

# Start development server
bun run dev
```

The application will be available at `http://localhost:3000`

### Backend Services

Make sure the following backend services are running:

- **User Service**: `http://localhost:8001`
- **Event Service**: `http://localhost:8002`
- **Search Service**: `http://localhost:8003`
- **Booking Service**: `http://localhost:8004`

## Usage

### User Flow

1. **Registration/Login**: Create account or login at `/register` or `/login`
2. **Browse Events**: Explore events at `/events` with search and filters
3. **Event Details**: View event information at `/events/:eventId`
4. **Booking**: Book tickets through the booking flow at `/book/:eventId`
5. **Dashboard**: Manage bookings and profile at `/dashboard`

### Admin Flow

1. **Admin Login**: Access admin panel at `/admin/login`
2. **Create Venues**: Manage venues at `/admin/venues`
3. **Create Events**: Create and publish events at `/admin/events`
4. **Dashboard**: Monitor events at `/admin/dashboard`

## API Integration

The frontend integrates with the following backend services:

- **Authentication**: JWT-based auth with token refresh
- **Event Search**: Real-time search with Elasticsearch
- **Booking**: Two-phase booking with concurrency control
- **Admin Operations**: Venue and event management

## Demo Credentials

### Admin Demo
```
Email: admin@bookmyevent.com
Password: admin123
```

### User Demo
Create your own user account through the registration flow.

## Development

### Project Structure

```
src/
├── components/          # Reusable components
│   ├── auth/           # Authentication components
│   └── layout/         # Layout components
├── contexts/           # React contexts
├── pages/              # Page components
│   ├── admin/          # Admin pages
│   └── auth/           # Auth pages
├── services/           # API services
└── App.jsx             # Main app component
```

### Key Components

- **AuthContext**: Manages user and admin authentication state
- **API Service**: Centralized API communication with interceptors
- **Protected Routes**: Route guards for authenticated areas
- **Responsive Design**: Mobile-first design with Tailwind CSS

## Features Demo

### Booking Flow
1. Select event and quantity
2. Reserve tickets (5-minute hold)
3. Complete payment (demo mode)
4. Receive confirmation with ticket URL

### Search & Filtering
- Text search across events
- Filter by city, type, price range
- Real-time results with facets

### Admin Management
- Create and manage venues
- Create events with full details
- Publish/unpublish events
- View event statistics

## Build

```bash
# Build for production
bun run build

# Preview production build
bun run preview
```

## Contributing

1. Follow the existing code structure
2. Use TypeScript for new components (optional)
3. Ensure responsive design
4. Test with both user and admin flows

## License

MIT License - see LICENSE file for details