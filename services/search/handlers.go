package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/fyzanshaik/bookmyevent-ily/internal/utils"
	"github.com/google/uuid"
)

func (e *ElasticsearchClient) HealthCheck(ctx context.Context) error {
	res, err := e.client.Ping()
	if err != nil {
		return fmt.Errorf("elasticsearch ping failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch health check failed: %s", res.String())
	}

	return nil
}


func (cfg *APIConfig) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	esStatus := "connected"
	if err := cfg.ESClient.HealthCheck(r.Context()); err != nil {
		cfg.Logger.Error("Elasticsearch health check failed", "error", err)
		esStatus = "disconnected"
	}

	redisStatus := "connected"
	if err := cfg.RedisClient.Ping(r.Context()).Err(); err != nil {
		cfg.Logger.Error("Redis health check failed", "error", err)
		redisStatus = "disconnected"
	}

	status := "ready"
	if esStatus == "disconnected" || redisStatus == "disconnected" {
		status = "not ready"
		utils.RespondWithJSON(w, http.StatusServiceUnavailable, HealthResponse{
			Status:        status,
			Elasticsearch: esStatus,
			Redis:         redisStatus,
			Service:       "search-service",
		})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, HealthResponse{
		Status:        status,
		Elasticsearch: esStatus,
		Redis:         redisStatus,
		Service:       "search-service",
	})
}

func (cfg *APIConfig) SearchEvents(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	city := r.URL.Query().Get("city")
	eventType := r.URL.Query().Get("type")
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	minPrice, _ := strconv.ParseFloat(r.URL.Query().Get("min_price"), 64)
	maxPrice, _ := strconv.ParseFloat(r.URL.Query().Get("max_price"), 64)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	searchReq := SearchRequest{
		Query:     query,
		City:      city,
		EventType: eventType,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		MinPrice:  minPrice,
		MaxPrice:  maxPrice,
		Page:      page,
		Limit:     limit,
	}

	cacheKey := fmt.Sprintf("search:%s:%s:%s:%s:%s:%.2f:%.2f:%d:%d", 
		query, city, eventType, dateFrom, dateTo, minPrice, maxPrice, page, limit)

	if cached := cfg.getCachedSearchResult(r.Context(), cacheKey); cached != nil {
		cfg.Logger.Info("Returning cached search result", "cache_key", cacheKey)
		utils.RespondWithJSON(w, http.StatusOK, cached)
		return
	}

	result, err := cfg.ESClient.Search(r.Context(), searchReq)
	if err != nil {
		cfg.Logger.Error("Search failed", "error", err, "query", query)
		utils.RespondWithError(w, http.StatusInternalServerError, "Search failed")
		return
	}

	cfg.cacheSearchResult(r.Context(), cacheKey, result)

	cfg.Logger.Info("Search completed", 
		"query", query,
		"results", len(result.Results),
		"total", result.Total,
		"query_time", result.QueryTime)

	utils.RespondWithJSON(w, http.StatusOK, result)
}

func (cfg *APIConfig) GetSuggestions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 20 {
		limit = 10
	}

	suggestions, err := cfg.ESClient.GetSuggestions(r.Context(), query, limit)
	if err != nil {
		cfg.Logger.Error("Failed to get suggestions", "error", err, "query", query)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get suggestions")
		return
	}

	response := SuggestionsResponse{
		Suggestions: suggestions,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) GetFilters(w http.ResponseWriter, r *http.Request) {
	emptySearch := SearchRequest{
		Page:  1,
		Limit: 1,
	}

	result, err := cfg.ESClient.Search(r.Context(), emptySearch)
	if err != nil {
		cfg.Logger.Error("Failed to get filters", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get filters")
		return
	}

	var cities []string
	for _, city := range result.Facets.Cities {
		cities = append(cities, city.Value)
	}

	var eventTypes []string
	for _, eventType := range result.Facets.EventTypes {
		eventTypes = append(eventTypes, eventType.Value)
	}

	response := FiltersResponse{
		Cities:     cities,
		EventTypes: eventTypes,
		PriceRange: result.Facets.PriceRange,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) GetTrendingEvents(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	searchReq := SearchRequest{
		Page:  1,
		Limit: limit,
	}

	result, err := cfg.ESClient.Search(r.Context(), searchReq)
	if err != nil {
		cfg.Logger.Error("Failed to get trending events", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get trending events")
		return
	}

	response := TrendingEventsResponse{
		Events: result.Results,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) IndexEvent(w http.ResponseWriter, r *http.Request) {
	var req IndexEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cfg.Logger.WithFields(map[string]any{"error": err.Error()}).Warn("Invalid JSON in event indexing")
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Event.EventID == uuid.Nil {
		cfg.Logger.Warn("Event indexing without event ID")
		utils.RespondWithError(w, http.StatusBadRequest, "Event ID is required")
		return
	}

	if err := cfg.ESClient.IndexEvent(r.Context(), req.Event.EventID, req.Event); err != nil {
		cfg.Logger.Error("Failed to index event", "error", err, "event_id", req.Event.EventID)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to index event")
		return
	}

	cfg.invalidateSearchCache(r.Context())

	cfg.Logger.WithFields(map[string]any{"event_id": req.Event.EventID, "event_name": req.Event.Name}).Info("Event indexed successfully")

	response := IndexEventResponse{
		Status:  "indexed",
		EventID: req.Event.EventID.String(),
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.PathValue("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		cfg.Logger.WithFields(map[string]any{"event_id_str": eventIDStr, "error": err.Error()}).Warn("Invalid event ID format in delete request")
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid event ID format")
		return
	}

	if err := cfg.ESClient.DeleteEvent(r.Context(), eventID); err != nil {
		cfg.Logger.Error("Failed to delete event", "error", err, "event_id", eventID)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete event")
		return
	}

	cfg.invalidateSearchCache(r.Context())

	cfg.Logger.WithFields(map[string]any{"event_id": eventID}).Info("Event deleted from search index")

	response := DeleteEventResponse{
		Status:  "deleted",
		EventID: eventID.String(),
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) FullResync(w http.ResponseWriter, r *http.Request) {
	var req FullResyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.ForceReindex = false
	}

	start := time.Now()

	cfg.Logger.Info("Starting full resync", "force_reindex", req.ForceReindex)

	if req.ForceReindex {
		if err := cfg.ESClient.DeleteIndex(r.Context()); err != nil {
			cfg.Logger.Error("Failed to delete index during resync", "error", err)
		}
	}

	if err := cfg.ESClient.CreateIndex(r.Context()); err != nil {
		cfg.Logger.Error("Failed to create index during resync", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create index")
		return
	}

	events, err := cfg.EventServiceClient.GetAllPublishedEvents(r.Context())
	if err != nil {
		cfg.Logger.Error("Failed to fetch events from Event Service", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch events")
		return
	}

	if len(events) == 0 {
		cfg.Logger.Info("No events to index")
		response := FullResyncResponse{
			Message:       "No events to index",
			EventsIndexed: 0,
			TimeTaken:     time.Since(start).String(),
		}
		utils.RespondWithJSON(w, http.StatusOK, response)
		return
	}

	documents := make([]EventDocument, len(events))
	for i, event := range events {
		documents[i] = cfg.convertEventToDocument(event)
	}

	batchSize := 100
	totalIndexed := 0

	for i := 0; i < len(documents); i += batchSize {
		end := i + batchSize
		if end > len(documents) {
			end = len(documents)
		}

		batch := documents[i:end]
		if err := cfg.ESClient.BulkIndex(r.Context(), batch); err != nil {
			cfg.Logger.Error("Failed to bulk index batch", "error", err, "batch_start", i, "batch_size", len(batch))
			continue
		}

		totalIndexed += len(batch)
		cfg.Logger.Info("Indexed batch", "batch", i/batchSize+1, "events_in_batch", len(batch), "total_indexed", totalIndexed)
	}

	timeTaken := time.Since(start)

	cfg.invalidateSearchCache(r.Context())

	cfg.Logger.Info("Full resync completed", 
		"events_indexed", totalIndexed,
		"time_taken", timeTaken,
		"events_per_second", float64(totalIndexed)/timeTaken.Seconds())

	response := FullResyncResponse{
		Message:       "Full resync completed successfully",
		EventsIndexed: totalIndexed,
		TimeTaken:     timeTaken.String(),
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) convertEventToDocument(event EventServiceEvent) EventDocument {
	doc := EventDocument{
		EventID:       event.EventID,
		Name:          event.Name,
		VenueID:       event.VenueID,
		EventType:     event.EventType,
		StartDateTime: event.StartDatetime,
		EndDateTime:   event.EndDatetime,
		BasePrice:     event.BasePrice,
		AvailableSeats: event.AvailableSeats,
		TotalCapacity: event.TotalCapacity,
		Status:        event.Status,
		Version:       event.Version,
		CreatedAt:     event.CreatedAt,
		UpdatedAt:     event.UpdatedAt,
	}

	if event.Description != nil {
		doc.Description = *event.Description
	}

	if event.VenueName != nil {
		doc.VenueName = *event.VenueName
	}

	if event.VenueAddress != nil {
		doc.VenueAddress = *event.VenueAddress
	}

	if event.VenueCity != nil {
		doc.VenueCity = *event.VenueCity
	}

	if event.VenueState != nil {
		doc.VenueState = *event.VenueState
	}

	if event.VenueCountry != nil {
		doc.VenueCountry = *event.VenueCountry
	}

	return doc
}

func (cfg *APIConfig) getCachedSearchResult(ctx context.Context, key string) *SearchResponse {
	val, err := cfg.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return nil
	}

	var result SearchResponse
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		cfg.Logger.Error("Failed to unmarshal cached search result", "error", err)
		return nil
	}

	return &result
}

func (cfg *APIConfig) cacheSearchResult(ctx context.Context, key string, result *SearchResponse) {
	data, err := json.Marshal(result)
	if err != nil {
		cfg.Logger.Error("Failed to marshal search result for caching", "error", err)
		return
	}

	if err := cfg.RedisClient.Set(ctx, key, data, cfg.Config.CacheExpiry).Err(); err != nil {
		cfg.Logger.Error("Failed to cache search result", "error", err)
	}
}

func (cfg *APIConfig) invalidateSearchCache(ctx context.Context) {
	pattern := "search:*"
	
	keys, err := cfg.RedisClient.Keys(ctx, pattern).Result()
	if err != nil {
		cfg.Logger.Error("Failed to get cache keys for invalidation", "error", err, "pattern", pattern)
		return
	}

	if len(keys) == 0 {
		cfg.Logger.Debug("No cache keys found to invalidate", "pattern", pattern)
		return
	}

	deletedCount, err := cfg.RedisClient.Del(ctx, keys...).Result()
	if err != nil {
		cfg.Logger.Error("Failed to delete cache keys", "error", err, "keys_count", len(keys))
		return
	}

	cfg.Logger.Info("Search cache invalidated", "keys_deleted", deletedCount, "pattern", pattern)
}
