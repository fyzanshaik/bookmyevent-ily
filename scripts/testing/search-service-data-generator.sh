#!/bin/bash

set -e

SEARCH_SERVICE_URL="http://localhost:8003"
INTERNAL_API_KEY="your-internal-api-key"
EVENTS_TO_CREATE=10000
BATCH_SIZE=50

echo "🚀 Search Service Data Generator"
echo "================================"
echo "Target URL: $SEARCH_SERVICE_URL"
echo "Events to create: $EVENTS_TO_CREATE"
echo "Batch size: $BATCH_SIZE"
echo ""

if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
    if [ ! -z "$INTERNAL_API_KEY" ]; then
        echo "✅ Using INTERNAL_API_KEY from .env file"
    fi
fi

check_service() {
    echo "🔍 Checking if Search Service is running..."
    if curl -s "$SEARCH_SERVICE_URL/healthz" > /dev/null; then
        echo "✅ Search Service is running"
    else
        echo "❌ Search Service is not running on $SEARCH_SERVICE_URL"
        echo "💡 Start it with: make run SERVICE=search-service"
        exit 1
    fi
}

generate_event_data() {
    local event_id="$1"
    local index="$2"
    
    # Event types
    event_types=("concert" "sports" "theater" "conference" "comedy" "festival" "workshop" "exhibition")
    event_type=${event_types[$((RANDOM % ${#event_types[@]}))]}
    
    # Cities
    cities=("New York" "Los Angeles" "Chicago" "Houston" "Phoenix" "Philadelphia" "San Antonio" "San Diego" "Dallas" "San Jose" "Austin" "Jacksonville" "Fort Worth" "Columbus" "Charlotte" "San Francisco" "Indianapolis" "Seattle" "Denver" "Boston" "El Paso" "Detroit" "Nashville" "Portland" "Memphis" "Oklahoma City" "Las Vegas" "Louisville" "Baltimore" "Milwaukee")
    city=${cities[$((RANDOM % ${#cities[@]}))]}
    
    # States
    states=("NY" "CA" "IL" "TX" "AZ" "PA" "TX" "CA" "TX" "CA" "TX" "FL" "TX" "OH" "NC" "CA" "IN" "WA" "CO" "MA" "TX" "MI" "TN" "OR" "TN" "OK" "NV" "KY" "MD" "WI")
    state=${states[$((RANDOM % ${#states[@]}))]}
    
    # Generate realistic event names based on type
    case $event_type in
        "concert")
            artists=("The Midnight Stars" "Electric Dreams" "Jazz Fusion Collective" "Rock Revolution" "Acoustic Nights" "Symphony of Sound" "Urban Beats" "Classical Crossover" "Indie Underground" "Pop Sensation")
            artist=${artists[$((RANDOM % ${#artists[@]}))]}
            event_name="$artist Live in Concert"
            ;;
        "sports")
            sports=("Basketball Championship" "Soccer Finals" "Tennis Tournament" "Baseball Game" "Football Match" "Hockey Playoffs" "Golf Tournament" "Swimming Competition" "Track & Field" "Boxing Match")
            sport=${sports[$((RANDOM % ${#sports[@]}))]}
            event_name="$city $sport"
            ;;
        "theater")
            plays=("Romeo and Juliet" "The Lion King" "Hamilton" "Phantom of the Opera" "Chicago" "Cats" "Les Miserables" "Wicked" "The Book of Mormon" "Dear Evan Hansen")
            play=${plays[$((RANDOM % ${#plays[@]}))]}
            event_name="$play - Broadway Musical"
            ;;
        "conference")
            conferences=("Tech Innovation Summit" "Digital Marketing Conference" "AI & Machine Learning Expo" "Startup Pitch Day" "Business Leadership Forum" "Data Science Convention" "Cybersecurity Summit" "Cloud Computing Conference" "Mobile App Development" "E-commerce Expo")
            conf=${conferences[$((RANDOM % ${#conferences[@]}))]}
            event_name="$conf 2024"
            ;;
        *)
            generic_names=("Grand Celebration" "Cultural Festival" "Art Exhibition" "Comedy Night" "Food Festival" "Music Festival" "Dance Performance" "Literary Reading" "Film Screening" "Fashion Show")
            generic=${generic_names[$((RANDOM % ${#generic_names[@]}))]}
            event_name="$generic"
            ;;
    esac
    
    # Generate venue names
    venue_types=("Arena" "Theater" "Convention Center" "Stadium" "Hall" "Auditorium" "Center" "Pavilion" "Coliseum" "Forum")
    venue_type=${venue_types[$((RANDOM % ${#venue_types[@]}))]}
    venue_name="$city $venue_type"
    
    # Generate realistic addresses
    street_numbers=$((RANDOM % 9999 + 1))
    street_names=("Main St" "Broadway" "First Ave" "Second St" "Park Ave" "Oak St" "Maple Ave" "Cedar St" "Pine Ave" "Elm St")
    street_name=${street_names[$((RANDOM % ${#street_names[@]}))]}
    venue_address="$street_numbers $street_name"
    
    # Generate random dates (next 6 months)
    days_ahead=$((RANDOM % 180 + 1))
    start_date=$(date -d "+$days_ahead days" --iso-8601)
    start_time=$(printf "%02d:%02d:00" $((RANDOM % 24)) $((RANDOM % 60)))
    start_datetime="${start_date}T${start_time}Z"
    
    # End time (1-4 hours later)
    duration_hours=$((RANDOM % 4 + 1))
    end_datetime=$(date -d "$start_datetime + $duration_hours hours" --iso-8601)
    
    # Generate realistic pricing
    base_prices=(25.00 35.00 45.00 55.00 75.00 95.00 125.00 150.00 200.00 250.00 350.00 500.00)
    base_price=${base_prices[$((RANDOM % ${#base_prices[@]}))]}
    
    # Generate capacity based on venue type
    if [[ $venue_type == "Stadium" || $venue_type == "Arena" ]]; then
        total_capacity=$((RANDOM % 50000 + 20000))
    elif [[ $venue_type == "Theater" || $venue_type == "Auditorium" ]]; then
        total_capacity=$((RANDOM % 2000 + 500))
    else
        total_capacity=$((RANDOM % 10000 + 1000))
    fi
    
    available_seats=$((total_capacity - (RANDOM % (total_capacity / 4))))
    
    # Generate descriptions
    descriptions=(
        "Join us for an unforgettable experience that will leave you wanting more."
        "Don't miss this incredible opportunity to witness greatness in person."
        "Experience the magic and excitement of live entertainment at its finest."
        "A spectacular event featuring world-class performers and production."
        "Come and be part of something truly special and memorable."
        "An evening of entertainment that promises to exceed all expectations."
        "Witness history in the making at this once-in-a-lifetime event."
        "The ultimate experience for fans and newcomers alike."
        "Prepare to be amazed by this extraordinary showcase of talent."
        "An event that brings together the best in entertainment and culture."
    )
    description=${descriptions[$((RANDOM % ${#descriptions[@]}))]}
    
    # Generate UUIDs (simplified for testing)
    event_uuid=$(python3 -c "import uuid; print(str(uuid.uuid4()))")
    venue_uuid=$(python3 -c "import uuid; print(str(uuid.uuid4()))")
    
    # Generate timestamps
    created_at=$(date -d "-$((RANDOM % 30 + 1)) days" --iso-8601)
    updated_at=$(date -d "-$((RANDOM % 7 + 1)) days" --iso-8601)
    
    cat << EOF
{
    "event": {
        "event_id": "$event_uuid",
        "name": "$event_name",
        "description": "$description",
        "venue_id": "$venue_uuid",
        "venue_name": "$venue_name",
        "venue_address": "$venue_address",
        "venue_city": "$city",
        "venue_state": "$state",
        "venue_country": "USA",
        "event_type": "$event_type",
        "start_datetime": "$start_datetime",
        "end_datetime": "$end_datetime",
        "base_price": $base_price,
        "available_seats": $available_seats,
        "total_capacity": $total_capacity,
        "status": "published",
        "version": 1,
        "created_at": "$created_at",
        "updated_at": "$updated_at"
    }
}
EOF
}

create_events_batch() {
    local start_index="$1"
    local batch_size="$2"
    local batch_number="$3"
    
    echo "📦 Creating batch $batch_number (events $start_index-$((start_index + batch_size - 1)))"
    
    local success_count=0
    local error_count=0
    
    for ((i=0; i<batch_size; i++)); do
        local event_index=$((start_index + i))
        
        if [ $event_index -gt $EVENTS_TO_CREATE ]; then
            break
        fi
        
        local event_data=$(generate_event_data "$event_index" "$event_index")
        
        local response=$(curl -s -w "\n%{http_code}" \
            -X POST \
            -H "Content-Type: application/json" \
            -H "X-API-Key: $INTERNAL_API_KEY" \
            -d "$event_data" \
            "$SEARCH_SERVICE_URL/internal/search/events" 2>/dev/null)
        
        local http_code=$(echo "$response" | tail -n1)
        local response_body=$(echo "$response" | head -n -1)
        
        if [ "$http_code" = "200" ]; then
            success_count=$((success_count + 1))
            if [ $((event_index % 100)) -eq 0 ]; then
                echo "  ✅ Event $event_index created successfully"
            fi
        else
            error_count=$((error_count + 1))
            echo "  ❌ Failed to create event $event_index (HTTP $http_code): $response_body"
        fi
        
        # Small delay to avoid overwhelming the service
        sleep 0.01
    done
    
    echo "  📊 Batch $batch_number complete: $success_count success, $error_count errors"
    return $success_count
}

main() {
    echo "🔍 Starting data generation process..."
    echo ""
    
    check_service
    
    echo ""
    echo "🏗️ Generating $EVENTS_TO_CREATE events in batches of $BATCH_SIZE..."
    echo ""
    
    local total_success=0
    local total_batches=$(( (EVENTS_TO_CREATE + BATCH_SIZE - 1) / BATCH_SIZE ))
    local start_time=$(date +%s)
    
    for ((batch=1; batch<=total_batches; batch++)); do
        local start_index=$(( (batch - 1) * BATCH_SIZE + 1 ))
        local current_batch_size=$BATCH_SIZE
        
        if [ $((start_index + BATCH_SIZE - 1)) -gt $EVENTS_TO_CREATE ]; then
            current_batch_size=$((EVENTS_TO_CREATE - start_index + 1))
        fi
        
        create_events_batch "$start_index" "$current_batch_size" "$batch"
        local batch_success=$?
        total_success=$((total_success + batch_success))
        
        # Progress update
        local progress=$((batch * 100 / total_batches))
        echo "  📈 Progress: $progress% ($total_success events created)"
        echo ""
    done
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    echo "🎉 Data generation completed!"
    echo "================================"
    echo "✅ Total events created: $total_success/$EVENTS_TO_CREATE"
    echo "⏱️  Time taken: ${duration}s"
    echo "📊 Average rate: $((total_success / (duration + 1))) events/second"
    echo ""
    
    if [ $total_success -gt 0 ]; then
        echo "🔍 Testing search functionality..."
        echo ""
        
        # Test basic search
        echo "Testing basic search:"
        curl -s "$SEARCH_SERVICE_URL/api/v1/search?limit=5" | python3 -m json.tool | head -20
        
        echo ""
        echo "Testing search with query:"
        curl -s "$SEARCH_SERVICE_URL/api/v1/search?q=concert&limit=3" | python3 -m json.tool | head -15
        
        echo ""
        echo "Testing filters:"
        curl -s "$SEARCH_SERVICE_URL/api/v1/search/filters" | python3 -m json.tool | head -10
        
        echo ""
        echo "🎯 Search service is ready for testing!"
        echo "Try these endpoints:"
        echo "  • GET $SEARCH_SERVICE_URL/api/v1/search"
        echo "  • GET $SEARCH_SERVICE_URL/api/v1/search?q=concert"
        echo "  • GET $SEARCH_SERVICE_URL/api/v1/search?city=New York"
        echo "  • GET $SEARCH_SERVICE_URL/api/v1/search/suggestions?q=con"
        echo "  • GET $SEARCH_SERVICE_URL/api/v1/search/trending"
    else
        echo "❌ No events were created successfully. Check the search service logs."
        exit 1
    fi
}

# Help function
show_help() {
    echo "Search Service Data Generator"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -u, --url URL       Search service URL (default: http://localhost:8003)"
    echo "  -k, --key KEY       Internal API key"
    echo "  -n, --count NUM     Number of events to create (default: 10000)"
    echo "  -b, --batch NUM     Batch size (default: 50)"
    echo "  -h, --help          Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  INTERNAL_API_KEY    Internal API key (can be set in .env file)"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Generate 10,000 events"
    echo "  $0 -n 1000                          # Generate 1,000 events"
    echo "  $0 -u http://localhost:8003 -n 5000 # Custom URL and count"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -u|--url)
            SEARCH_SERVICE_URL="$2"
            shift 2
            ;;
        -k|--key)
            INTERNAL_API_KEY="$2"
            shift 2
            ;;
        -n|--count)
            EVENTS_TO_CREATE="$2"
            shift 2
            ;;
        -b|--batch)
            BATCH_SIZE="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Check if python3 is available (needed for UUID generation)
if ! command -v python3 &> /dev/null; then
    echo "❌ python3 is required but not installed."
    exit 1
fi

# Run main function
main
