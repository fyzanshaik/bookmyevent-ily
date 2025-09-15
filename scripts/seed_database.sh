#!/bin/bash

# Database Seeding Script for BookMyEvent
# This script creates users, admin, venues, and events for testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
USER_SERVICE_URL="http://localhost:8001"
EVENT_SERVICE_URL="http://localhost:8002"
ADMIN_EMAIL="atlanadmin@mail.com"
ADMIN_PASSWORD="11111111"
USER1_EMAIL="atlanuser1@mail.com"
USER2_EMAIL="atlanuser2@mail.com"
USER_PASSWORD="11111111"

# Temporary files
TEMP_DIR="/tmp/bookmyevent_seed"
mkdir -p "$TEMP_DIR"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to make HTTP requests
make_request() {
    local method="$1"
    local url="$2"
    local data="$3"
    local headers="$4"
    local output_file="$5"
    
    
    if [ -n "$data" ]; then
        if [ -n "$headers" ]; then
            curl -s -X "$method" "$url" \
                -H "Content-Type: application/json" \
                $headers \
                -d "$data" \
                -o "$output_file" \
                -w "%{http_code}"
        else
            curl -s -X "$method" "$url" \
                -H "Content-Type: application/json" \
                -d "$data" \
                -o "$output_file" \
                -w "%{http_code}"
        fi
    else
        if [ -n "$headers" ]; then
            curl -s -X "$method" "$url" \
                -H "Content-Type: application/json" \
                $headers \
                -o "$output_file" \
                -w "%{http_code}"
        else
            curl -s -X "$method" "$url" \
                -H "Content-Type: application/json" \
                -o "$output_file" \
                -w "%{http_code}"
        fi
    fi
}

# Function to check if service is running
check_service() {
    local service_name="$1"
    local url="$2"
    
    print_status "Checking if $service_name is running..."
    if curl -s "$url/health" > /dev/null 2>&1; then
        print_success "$service_name is running"
        return 0
    else
        print_error "$service_name is not running at $url"
        return 1
    fi
}

# Function to register a user
register_user() {
    local email="$1"
    local password="$2"
    local name="$3"
    local phone="$4"
    
    print_status "Registering user: $email"
    
    local user_data=$(cat <<EOF
{
    "email": "$email",
    "password": "$password",
    "name": "$name",
    "phone_number": "$phone"
}
EOF
)
    
    local response_file="$TEMP_DIR/user_register_${email//@/_}.json"
    local status_code=$(make_request "POST" "$USER_SERVICE_URL/api/v1/auth/register" "$user_data" "" "$response_file")
    
    if [ "$status_code" = "201" ]; then
        print_success "User $email registered successfully"
        return 0
    elif [ "$status_code" = "409" ]; then
        print_warning "User $email already exists, skipping..."
        return 0
    else
        print_error "Failed to register user $email (Status: $status_code)"
        cat "$response_file"
        return 1
    fi
}

# Function to register admin
register_admin() {
    print_status "Registering admin: $ADMIN_EMAIL"
    
    local admin_data=$(cat <<EOF
{
    "email": "$ADMIN_EMAIL",
    "password": "$ADMIN_PASSWORD",
    "name": "Atlan Admin",
    "phone_number": "987-654-3210",
    "role": "super_admin"
}
EOF
)
    
    local response_file="$TEMP_DIR/admin_register.json"
    local status_code=$(make_request "POST" "$EVENT_SERVICE_URL/api/v1/auth/admin/register" "$admin_data" "" "$response_file")
    
    if [ "$status_code" = "201" ]; then
        print_success "Admin $ADMIN_EMAIL registered successfully"
        return 0
    elif [ "$status_code" = "409" ]; then
        print_warning "Admin $ADMIN_EMAIL already exists, skipping..."
        return 0
    else
        print_error "Failed to register admin $ADMIN_EMAIL (Status: $status_code)"
        cat "$response_file"
        return 1
    fi
}

# Function to login admin and get token
login_admin() {
    print_status "Logging in admin..."
    
    local login_data=$(cat <<EOF
{
    "email": "$ADMIN_EMAIL",
    "password": "$ADMIN_PASSWORD"
}
EOF
)
    
    local response_file="$TEMP_DIR/admin_login.json"
    local status_code=$(make_request "POST" "$EVENT_SERVICE_URL/api/v1/auth/admin/login" "$login_data" "" "$response_file")
    
    if [ "$status_code" = "200" ]; then
        local token=$(jq -r '.access_token // .data.access_token // .token // empty' "$response_file")
        if [ -n "$token" ] && [ "$token" != "null" ]; then
            print_success "Admin login successful"
            echo "$token" > "$TEMP_DIR/admin_token.txt"
            return 0
        else
            print_error "No access token found in login response"
            cat "$response_file"
            return 1
        fi
    else
        print_error "Admin login failed (Status: $status_code)"
        cat "$response_file"
        return 1
    fi
}

# Function to create venue
create_venue() {
    local name="$1"
    local address="$2"
    local city="$3"
    local state="$4"
    local country="$5"
    local postal_code="$6"
    local capacity="$7"
    local sections="$8"
    
    local token=$(cat "$TEMP_DIR/admin_token.txt")
    
    local venue_data=$(cat <<EOF
{
    "name": "$name",
    "address": "$address",
    "city": "$city",
    "state": "$state",
    "country": "$country",
    "postal_code": "$postal_code",
    "capacity": $capacity,
    "layout_config": {
        "sections": $sections
    }
}
EOF
)
    
    local response_file="$TEMP_DIR/venue_${name// /_}.json"
    print_status "Creating venue '$name' with token: ${token:0:20}..."
    
    # Make the request directly with proper header handling
    local status_code=$(curl -s -X "POST" "$EVENT_SERVICE_URL/api/v1/admin/venues" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d "$venue_data" \
        -o "$response_file" \
        -w "%{http_code}")
    
    if [ "$status_code" = "201" ]; then
        local venue_id=$(jq -r '.venue_id' "$response_file")
        print_success "Venue '$name' created with ID: $venue_id"
        echo "$venue_id" >> "$TEMP_DIR/venue_ids.txt"
        return 0
    else
        print_error "Failed to create venue '$name' (Status: $status_code)"
        cat "$response_file"
        return 1
    fi
}

# Function to create event
create_event() {
    local title="$1"
    local description="$2"
    local venue_id="$3"
    local start_time="$4"
    local end_time="$5"
    local ticket_price="$6"
    local status="$7"
    local category="$8"
    
    local token=$(cat "$TEMP_DIR/admin_token.txt")
    
    local event_data=$(cat <<EOF
{
    "name": "$title",
    "description": "$description",
    "venue_id": "$venue_id",
    "start_datetime": "$start_time",
    "end_datetime": "$end_time",
    "base_price": $ticket_price,
    "total_capacity": 100,
    "event_type": "$category"
}
EOF
)
    
    local response_file="$TEMP_DIR/event_${title// /_}.json"
    
    # Make the request directly with proper header handling
    local status_code=$(curl -s -X "POST" "$EVENT_SERVICE_URL/api/v1/admin/events" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d "$event_data" \
        -o "$response_file" \
        -w "%{http_code}")
    
    if [ "$status_code" = "201" ]; then
        local event_id=$(jq -r '.event_id' "$response_file")
        print_success "Event '$title' created with ID: $event_id"
        echo "$event_id" >> "$TEMP_DIR/event_ids.txt"
        return 0
    else
        print_error "Failed to create event '$title' (Status: $status_code)"
        cat "$response_file"
        return 1
    fi
}

# Function to get event details
get_event_details() {
    local event_id="$1"
    local token=$(cat "$TEMP_DIR/admin_token.txt")
    
    local response_file="$TEMP_DIR/event_get_${event_id}.json"
    
    local status_code=$(curl -s -X "GET" "$EVENT_SERVICE_URL/api/v1/events/$event_id" \
        -o "$response_file" \
        -w "%{http_code}")
    
    if [ "$status_code" = "200" ]; then
        local version=$(jq -r '.version' "$response_file" 2>/dev/null || echo "1")
        local current_status=$(jq -r '.status' "$response_file" 2>/dev/null || echo "draft")
        echo "$version|$current_status"
        return 0
    else
        print_error "Failed to get event $event_id details (Status: $status_code)"
        cat "$response_file"
        return 1
    fi
}

# Function to update event status
update_event_status() {
    local event_id="$1"
    local status="$2"
    
    local token=$(cat "$TEMP_DIR/admin_token.txt")
    
    # First get the current event details to get the correct version
    local event_details=$(get_event_details "$event_id")
    if [ $? -ne 0 ]; then
        return 1
    fi
    
    IFS='|' read -r version current_status <<< "$event_details"
    
    # Skip if already in the desired status
    if [ "$current_status" = "$status" ]; then
        print_warning "Event $event_id is already in '$status' status, skipping..."
        return 0
    fi
    
    local update_data=$(cat <<EOF
{
    "status": "$status",
    "version": $version
}
EOF
)
    
    
    local response_file="$TEMP_DIR/event_update_${event_id}.json"
    
    # Make the request directly with proper header handling
    local status_code=$(curl -s -X "PUT" "$EVENT_SERVICE_URL/api/v1/admin/events/$event_id" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d "$update_data" \
        -o "$response_file" \
        -w "%{http_code}")
    
    if [ "$status_code" = "200" ]; then
        print_success "Event $event_id status updated to '$status'"
        return 0
    else
        print_error "Failed to update event $event_id status (Status: $status_code)"
        cat "$response_file"
        return 1
    fi
}

# Main execution
main() {
    print_status "Starting database seeding process..."
    
    # Check if services are running
    if ! check_service "User Service" "$USER_SERVICE_URL"; then
        exit 1
    fi
    
    if ! check_service "Event Service" "$EVENT_SERVICE_URL"; then
        exit 1
    fi
    
    # Register users
    print_status "=== Registering Users ==="
    register_user "$USER1_EMAIL" "$USER_PASSWORD" "Atlan User 1" "9876543210"
    register_user "$USER2_EMAIL" "$USER_PASSWORD" "Atlan User 2" "9876543211"
    
    # Register admin
    print_status "=== Registering Admin ==="
    register_admin
    
    # Login admin
    print_status "=== Logging in Admin ==="
    login_admin
    
    # Create venues
    print_status "=== Creating Venues ==="
    
    # Indian venues data
    declare -a venues=(
        "Akshardham Temple|Akshardham Temple Road|Delhi|Delhi|India|110092|25|[{\"name\":\"Main Hall\",\"rows\":5,\"seats_per_row\":5}]"
        "Red Fort|Netaji Subhash Marg|Delhi|Delhi|India|110006|30|[{\"name\":\"Diwan-i-Aam\",\"rows\":6,\"seats_per_row\":5}]"
        "Gateway of India|Apollo Bandar|Mumbai|Maharashtra|India|400001|40|[{\"name\":\"Main Area\",\"rows\":8,\"seats_per_row\":5}]"
        "Taj Mahal|Dharmapuri|Agra|Uttar Pradesh|India|282001|35|[{\"name\":\"Garden View\",\"rows\":7,\"seats_per_row\":5}]"
        "Hawa Mahal|Hawa Mahal Rd|Jaipur|Rajasthan|India|302002|20|[{\"name\":\"Palace View\",\"rows\":4,\"seats_per_row\":5}]"
        "Golden Temple|Golden Temple Road|Amritsar|Punjab|India|143006|45|[{\"name\":\"Main Hall\",\"rows\":9,\"seats_per_row\":5}]"
        "Meenakshi Temple|Temple Street|Madurai|Tamil Nadu|India|625001|15|[{\"name\":\"Temple Hall\",\"rows\":3,\"seats_per_row\":5}]"
        "Victoria Memorial|Queen's Way|Kolkata|West Bengal|India|700071|50|[{\"name\":\"Main Gallery\",\"rows\":10,\"seats_per_row\":5}]"
        "Charminar|Charminar Road|Hyderabad|Telangana|India|500002|28|[{\"name\":\"Monument View\",\"rows\":6,\"seats_per_row\":5}]"
        "Lotus Temple|Bahapur|Delhi|Delhi|India|110019|22|[{\"name\":\"Prayer Hall\",\"rows\":5,\"seats_per_row\":5}]"
    )
    
    for venue_info in "${venues[@]}"; do
        IFS='|' read -r name address city state country postal_code capacity sections <<< "$venue_info"
        create_venue "$name" "$address" "$city" "$state" "$country" "$postal_code" "$capacity" "$sections"
    done
    
    # Read venue IDs
    if [ -f "$TEMP_DIR/venue_ids.txt" ]; then
        mapfile -t venue_ids < "$TEMP_DIR/venue_ids.txt"
    else
        print_error "No venue IDs found"
        exit 1
    fi
    
    # Create events
    print_status "=== Creating Events ==="
    
    # Published events (20)
    declare -a published_events=(
        "Classical Music Concert|An evening of classical Indian music|0|2025-10-15T18:00:00Z|2025-10-15T21:00:00Z|1500|published|music"
        "Bollywood Dance Show|High-energy Bollywood dance performance|1|2025-10-20T19:00:00Z|2025-10-20T22:00:00Z|2000|published|dance"
        "Sitar Recital|Traditional sitar performance|2|2025-10-25T17:00:00Z|2025-10-25T19:00:00Z|1200|published|music"
        "Folk Dance Festival|Traditional folk dances from across India|3|2025-11-01T18:30:00Z|2025-11-01T21:30:00Z|1800|published|dance"
        "Tabla Workshop|Learn the art of tabla playing|4|2025-11-05T16:00:00Z|2025-11-05T18:00:00Z|800|published|workshop"
        "Carnatic Music Concert|South Indian classical music|5|2025-11-10T19:00:00Z|2025-11-10T22:00:00Z|1600|published|music"
        "Kathak Performance|Traditional Kathak dance|6|2025-11-15T18:00:00Z|2025-11-15T20:00:00Z|1400|published|dance"
        "Hindustani Vocal|Classical Hindustani singing|7|2025-11-20T17:30:00Z|2025-11-20T20:30:00Z|1300|published|music"
        "Bharatanatyam Show|Classical Bharatanatyam dance|8|2025-11-25T19:00:00Z|2025-11-25T21:00:00Z|1700|published|dance"
        "Fusion Music Night|Modern fusion of Indian and Western music|9|2025-12-01T20:00:00Z|2025-12-01T23:00:00Z|2200|published|music"
        "Qawwali Evening|Sufi qawwali performance|0|2025-12-05T18:00:00Z|2025-12-05T21:00:00Z|1900|published|music"
        "Odissi Dance|Classical Odissi dance performance|1|2025-12-10T17:00:00Z|2025-12-10T19:00:00Z|1500|published|dance"
        "Ghazal Night|Romantic ghazal singing|2|2025-12-15T19:30:00Z|2025-12-15T22:30:00Z|1800|published|music"
        "Kuchipudi Performance|Traditional Kuchipudi dance|3|2025-12-20T18:00:00Z|2025-12-20T20:00:00Z|1600|published|dance"
        "Raga Concert|Morning raga performance|4|2025-12-25T10:00:00Z|2025-12-25T12:00:00Z|1000|published|music"
        "Contemporary Dance|Modern Indian contemporary dance|5|2025-12-30T19:00:00Z|2025-12-30T21:30:00Z|2000|published|dance"
        "Folk Music Festival|Traditional folk music from different states|6|2026-01-05T18:00:00Z|2026-01-05T21:00:00Z|1700|published|music"
        "Classical Instrumental|Various classical instruments|7|2026-01-10T17:30:00Z|2026-01-10T20:30:00Z|1400|published|music"
        "Bhangra Performance|Energetic Punjabi bhangra|8|2026-01-15T19:00:00Z|2026-01-15T21:30:00Z|2100|published|dance"
        "Sufi Music Night|Mystical Sufi music performance|9|2026-01-20T18:30:00Z|2026-01-20T21:30:00Z|1900|published|music"
    )
    
    for event_info in "${published_events[@]}"; do
        IFS='|' read -r title description venue_idx start_time end_time ticket_price status category <<< "$event_info"
        create_event "$title" "$description" "${venue_ids[$venue_idx]}" "$start_time" "$end_time" "$ticket_price" "$status" "$category"
    done
    
    # Draft events (5)
    declare -a draft_events=(
        "Jazz Fusion|Indian jazz fusion experiment|0|2026-02-01T20:00:00Z|2026-02-01T23:00:00Z|2500|draft|music"
        "Modern Dance|Contemporary dance exploration|1|2026-02-05T19:00:00Z|2026-02-05T21:00:00Z|1800|draft|dance"
        "Electronic Raga|Electronic music with raga elements|2|2026-02-10T21:00:00Z|2026-02-10T23:30:00Z|3000|draft|music"
        "Experimental Theater|Fusion of dance and theater|3|2026-02-15T18:00:00Z|2026-02-15T20:30:00Z|2200|draft|theater"
        "World Music|Global music with Indian influences|4|2026-02-20T19:30:00Z|2026-02-20T22:00:00Z|2800|draft|music"
    )
    
    for event_info in "${draft_events[@]}"; do
        IFS='|' read -r title description venue_idx start_time end_time ticket_price status category <<< "$event_info"
        create_event "$title" "$description" "${venue_ids[$venue_idx]}" "$start_time" "$end_time" "$ticket_price" "$status" "$category"
    done
    
    # Update first 20 events to published status
    print_status "=== Updating Event Status to Published ==="
    if [ -f "$TEMP_DIR/event_ids.txt" ]; then
        mapfile -t event_ids < "$TEMP_DIR/event_ids.txt"
        local published_count=0
        # Temporarily disable set -e for this section
        set +e
        for event_id in "${event_ids[@]}"; do
            if [ $published_count -lt 20 ]; then
                if update_event_status "$event_id" "published"; then
                    ((published_count++))
                else
                    print_error "Failed to update event $event_id, continuing with next..."
                    # Don't break, continue with next event
                fi
            else
                break
            fi
        done
        set -e
        print_success "Updated $published_count events to published status"
    else
        print_warning "No event IDs found for status update"
    fi
    
    # Cleanup
    rm -rf "$TEMP_DIR"
    
    print_success "Database seeding completed successfully!"
    print_status "Created:"
    print_status "  - 2 users"
    print_status "  - 1 admin"
    print_status "  - 10 venues"
    print_status "  - 20 published events"
    print_status "  - 5 draft events"
}

# Run main function
main "$@"
