package search

import (
	"time"

	"github.com/fyzanshaik/bookmyevent-ily/internal/config"
	"github.com/fyzanshaik/bookmyevent-ily/internal/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type APIConfig struct {
	Config        *config.SearchServiceConfig
	Logger        *logger.Logger
	ESClient      *ElasticsearchClient
	RedisClient   *redis.Client
	EventServiceClient *EventServiceClient
}

type SearchRequest struct {
	Query     string  `json:"q"`
	City      string  `json:"city"`
	EventType string  `json:"type"`
	DateFrom  string  `json:"date_from"`
	DateTo    string  `json:"date_to"`
	MinPrice  float64 `json:"min_price"`
	MaxPrice  float64 `json:"max_price"`
	Page      int     `json:"page"`
	Limit     int     `json:"limit"`
}

type SearchResponse struct {
	Results   []EventSearchResult `json:"results"`
	Total     int64               `json:"total"`
	Page      int                 `json:"page"`
	Limit     int                 `json:"limit"`
	QueryTime string              `json:"query_time"`
	Facets    SearchFacets        `json:"facets"`
}

type EventSearchResult struct {
	EventID       uuid.UUID `json:"event_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	VenueName     string    `json:"venue_name"`
	VenueCity     string    `json:"venue_city"`
	VenueAddress  string    `json:"venue_address,omitempty"`
	EventType     string    `json:"event_type"`
	StartDateTime time.Time `json:"start_datetime"`
	EndDateTime   time.Time `json:"end_datetime"`
	BasePrice     float64   `json:"base_price"`
	AvailableSeats int32    `json:"available_seats"`
	Status        string    `json:"status"`
	Score         float64   `json:"score"`
}

type SearchFacets struct {
	Cities     []FacetItem `json:"cities"`
	EventTypes []FacetItem `json:"event_types"`
	PriceRange PriceRange  `json:"price_range"`
}

type FacetItem struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

type PriceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type SuggestionsRequest struct {
	Query string `json:"q"`
	Limit int    `json:"limit"`
}

type SuggestionsResponse struct {
	Suggestions []string `json:"suggestions"`
}

type FiltersResponse struct {
	Cities     []string   `json:"cities"`
	EventTypes []string   `json:"event_types"`
	PriceRange PriceRange `json:"price_range"`
}

type TrendingEventsResponse struct {
	Events []EventSearchResult `json:"events"`
}

type EventDocument struct {
	EventID       uuid.UUID `json:"event_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	VenueID       uuid.UUID `json:"venue_id"`
	VenueName     string    `json:"venue_name"`
	VenueAddress  string    `json:"venue_address,omitempty"`
	VenueCity     string    `json:"venue_city"`
	VenueState    string    `json:"venue_state,omitempty"`
	VenueCountry  string    `json:"venue_country"`
	EventType     string    `json:"event_type"`
	StartDateTime time.Time `json:"start_datetime"`
	EndDateTime   time.Time `json:"end_datetime"`
	BasePrice     float64   `json:"base_price"`
	AvailableSeats int32    `json:"available_seats"`
	TotalCapacity int32     `json:"total_capacity"`
	Status        string    `json:"status"`
	Version       int32     `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type FullResyncRequest struct {
	ForceReindex bool `json:"force_reindex"`
}

type FullResyncResponse struct {
	Message       string `json:"message"`
	EventsIndexed int    `json:"events_indexed"`
	TimeTaken     string `json:"time_taken"`
}

type IndexEventRequest struct {
	Event EventDocument `json:"event"`
}

type IndexEventResponse struct {
	Status  string `json:"status"`
	EventID string `json:"event_id"`
}

type DeleteEventRequest struct {
	EventID uuid.UUID `json:"event_id"`
}

type DeleteEventResponse struct {
	Status  string `json:"status"`
	EventID string `json:"event_id"`
}

type HealthResponse struct {
	Status        string `json:"status"`
	Elasticsearch string `json:"elasticsearch"`
	Redis         string `json:"redis"`
	Service       string `json:"service"`
}
