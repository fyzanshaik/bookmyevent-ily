package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/fyzanshaik/bookmyevent-ily/internal/logger"
	"github.com/fyzanshaik/bookmyevent-ily/internal/utils"
	"github.com/google/uuid"
)

type ElasticsearchClient struct {
	client    *elasticsearch.Client
	indexName string
	logger    *logger.Logger
}

func NewElasticsearchClient(url, indexName string, logger *logger.Logger) (*ElasticsearchClient, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{url},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	return &ElasticsearchClient{
		client:    client,
		indexName: indexName,
		logger:    logger,
	}, nil
}

func (e *ElasticsearchClient) CreateIndex(ctx context.Context) error {
	mapping := map[string]any{
		"mappings": map[string]any{
			"properties": map[string]any{
				"event_id":    map[string]any{"type": "keyword"},
				"name": map[string]any{
					"type":     "text",
					"analyzer": "standard",
					"fields": map[string]any{
						"keyword": map[string]any{"type": "keyword"},
						"suggest": map[string]any{
							"type":                    "completion",
							"analyzer":                "simple",
							"preserve_separators":     true,
							"preserve_position_increments": true,
							"max_input_length":        50,
						},
					},
				},
				"description": map[string]any{
					"type":     "text",
					"analyzer": "standard",
				},
				"venue_id":      map[string]any{"type": "keyword"},
				"venue_name":    map[string]any{"type": "text", "fields": map[string]any{"keyword": map[string]any{"type": "keyword"}}},
				"venue_address": map[string]any{"type": "text"},
				"venue_city":    map[string]any{"type": "keyword"},
				"venue_state":   map[string]any{"type": "keyword"},
				"venue_country": map[string]any{"type": "keyword"},
				"event_type":    map[string]any{"type": "keyword"},
				"start_datetime": map[string]any{"type": "date"},
				"end_datetime":   map[string]any{"type": "date"},
				"base_price":     map[string]any{"type": "float"},
				"available_seats": map[string]any{"type": "integer"},
				"total_capacity":  map[string]any{"type": "integer"},
				"status":         map[string]any{"type": "keyword"},
				"version":        map[string]any{"type": "long"},
				"created_at":     map[string]any{"type": "date"},
				"updated_at":     map[string]any{"type": "date"},
			},
		},
		"settings": map[string]any{
			"number_of_shards":   1,
			"number_of_replicas": 0,
			"analysis": map[string]any{
				"analyzer": map[string]any{
					"event_search": map[string]any{
						"type":      "custom",
						"tokenizer": "standard",
						"filter":    []string{"lowercase", "stop"},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(mapping); err != nil {
		return fmt.Errorf("failed to encode mapping: %w", err)
	}

	req := esapi.IndicesCreateRequest{
		Index: e.indexName,
		Body:  &buf,
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if strings.Contains(res.String(), "resource_already_exists_exception") {
			e.logger.Info("Index already exists", "index", e.indexName)
			return nil
		}
		return fmt.Errorf("failed to create index: %s", res.String())
	}

	e.logger.Info("Created Elasticsearch index", "index", e.indexName)
	return nil
}

func (e *ElasticsearchClient) IndexEvent(ctx context.Context, eventID uuid.UUID, doc EventDocument) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(doc); err != nil {
		return fmt.Errorf("failed to encode document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      e.indexName,
		DocumentID: eventID.String(),
		Body:       &buf,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index document: %s", res.String())
	}

	return nil
}

func (e *ElasticsearchClient) DeleteEvent(ctx context.Context, eventID uuid.UUID) error {
	req := esapi.DeleteRequest{
		Index:      e.indexName,
		DocumentID: eventID.String(),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("failed to delete document: %s", res.String())
	}

	return nil
}

func (e *ElasticsearchClient) Search(ctx context.Context, searchReq SearchRequest) (*SearchResponse, error) {
	query := e.buildSearchQuery(searchReq)
	
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode search query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{e.indexName},
		Body:  &buf,
	}

	start := time.Now()
	res, err := req.Do(ctx, e.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search failed: %s", res.String())
	}

	var searchResult map[string]any
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search result: %w", err)
	}

	return e.parseSearchResponse(searchResult, searchReq, time.Since(start))
}

func (e *ElasticsearchClient) buildSearchQuery(req SearchRequest) map[string]any {
	from := 0
	if req.Page > 1 {
		from = (req.Page - 1) * req.Limit
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}

	query := map[string]any{
		"from": from,
		"size": req.Limit,
		"query": map[string]any{
			"bool": map[string]any{
				"must": []any{
					map[string]any{
						"term": map[string]any{
							"status": "published",
						},
					},
				},
				"filter": []any{},
			},
		},
		"sort": []any{
			map[string]any{
				"start_datetime": map[string]any{
					"order": "asc",
				},
			},
		},
		"aggs": map[string]any{
			"cities": map[string]any{
				"terms": map[string]any{
					"field": "venue_city",
					"size":  20,
				},
			},
			"event_types": map[string]any{
				"terms": map[string]any{
					"field": "event_type",
					"size":  20,
				},
			},
			"price_stats": map[string]any{
				"stats": map[string]any{
					"field": "base_price",
				},
			},
		},
	}

	boolQuery := query["query"].(map[string]any)["bool"].(map[string]any)

	if req.Query != "" {
		boolQuery["must"] = append(boolQuery["must"].([]any), map[string]any{
			"multi_match": map[string]any{
				"query": req.Query,
				"fields": []string{
					"name^3",
					"description^2",
					"venue_name^2",
					"venue_city",
					"event_type",
				},
				"type":                "best_fields",
				"minimum_should_match": "75%",
			},
		})
	}

	filters := boolQuery["filter"].([]any)

	if req.City != "" {
		filters = append(filters, map[string]any{
			"term": map[string]any{
				"venue_city": req.City,
			},
		})
	}

	if req.EventType != "" {
		filters = append(filters, map[string]any{
			"term": map[string]any{
				"event_type": req.EventType,
			},
		})
	}

	if req.DateFrom != "" || req.DateTo != "" {
		dateRange := map[string]any{}
		if req.DateFrom != "" {
			dateRange["gte"] = req.DateFrom
		}
		if req.DateTo != "" {
			dateRange["lte"] = req.DateTo
		}
		filters = append(filters, map[string]any{
			"range": map[string]any{
				"start_datetime": dateRange,
			},
		})
	}

	if req.MinPrice > 0 || req.MaxPrice > 0 {
		priceRange := map[string]any{}
		if req.MinPrice > 0 {
			priceRange["gte"] = req.MinPrice
		}
		if req.MaxPrice > 0 {
			priceRange["lte"] = req.MaxPrice
		}
		filters = append(filters, map[string]any{
			"range": map[string]any{
				"base_price": priceRange,
			},
		})
	}

	boolQuery["filter"] = filters

	return query
}

func (e *ElasticsearchClient) parseSearchResponse(result map[string]any, req SearchRequest, queryTime time.Duration) (*SearchResponse, error) {
	hits, ok := result["hits"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid search response: missing hits")
	}
	
	var total int64
	if totalInfo, ok := hits["total"].(map[string]any); ok {
		if value, ok := totalInfo["value"].(float64); ok {
			total = int64(value)
		}
	}
	
	var events []EventSearchResult
	if hitsArray, ok := hits["hits"].([]any); ok {
		for _, hit := range hitsArray {
			hitMap, ok := hit.(map[string]any)
			if !ok {
				continue
			}
			
			source, ok := hitMap["_source"].(map[string]any)
			if !ok {
				continue
			}
			
			var score float64
			if scoreVal, ok := hitMap["_score"]; ok && scoreVal != nil {
				if s, ok := scoreVal.(float64); ok {
					score = s
				}
			}

			eventIDStr := utils.GetStringFromInterface(source["event_id"])
			eventID, _ := uuid.Parse(eventIDStr)
			
			startDateStr := utils.GetStringFromInterface(source["start_datetime"])
			startDateTime, _ := time.Parse(time.RFC3339, startDateStr)
			
			endDateStr := utils.GetStringFromInterface(source["end_datetime"])
			endDateTime, _ := time.Parse(time.RFC3339, endDateStr)

			event := EventSearchResult{
				EventID:       eventID,
				Name:          utils.GetStringFromInterface(source["name"]),
				VenueName:     utils.GetStringFromInterface(source["venue_name"]),
				VenueCity:     utils.GetStringFromInterface(source["venue_city"]),
				VenueAddress:  utils.GetStringFromInterface(source["venue_address"]),
				EventType:     utils.GetStringFromInterface(source["event_type"]),
				StartDateTime: startDateTime,
				EndDateTime:   endDateTime,
				BasePrice:     utils.GetFloatFromInterface(source["base_price"]),
				AvailableSeats: int32(utils.GetFloatFromInterface(source["available_seats"])),
				Status:        utils.GetStringFromInterface(source["status"]),
				Score:         score,
			}

			if desc, ok := source["description"]; ok && desc != nil {
				event.Description = utils.GetStringFromInterface(desc)
			}

			events = append(events, event)
		}
	}

	facets := SearchFacets{}
	if aggs, ok := result["aggregations"].(map[string]any); ok {
		if cities, ok := aggs["cities"].(map[string]any); ok {
			if buckets, ok := cities["buckets"].([]any); ok {
				for _, bucket := range buckets {
					if b, ok := bucket.(map[string]any); ok {
						facets.Cities = append(facets.Cities, FacetItem{
							Value: utils.GetStringFromInterface(b["key"]),
							Count: int64(utils.GetFloatFromInterface(b["doc_count"])),
						})
					}
				}
			}
		}

		if eventTypes, ok := aggs["event_types"].(map[string]any); ok {
			if buckets, ok := eventTypes["buckets"].([]any); ok {
				for _, bucket := range buckets {
					if b, ok := bucket.(map[string]any); ok {
						facets.EventTypes = append(facets.EventTypes, FacetItem{
							Value: utils.GetStringFromInterface(b["key"]),
							Count: int64(utils.GetFloatFromInterface(b["doc_count"])),
						})
					}
				}
			}
		}

		if priceStats, ok := aggs["price_stats"].(map[string]any); ok {
			if min, ok := priceStats["min"]; ok && min != nil {
				facets.PriceRange.Min = utils.GetFloatFromInterface(min)
			}
			if max, ok := priceStats["max"]; ok && max != nil {
				facets.PriceRange.Max = utils.GetFloatFromInterface(max)
			}
		}
	}

	return &SearchResponse{
		Results:   events,
		Total:     total,
		Page:      req.Page,
		Limit:     req.Limit,
		QueryTime: queryTime.String(),
		Facets:    facets,
	}, nil
}

func (e *ElasticsearchClient) GetSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	searchQuery := map[string]any{
		"suggest": map[string]any{
			"event_suggestions": map[string]any{
				"prefix": query,
				"completion": map[string]any{
					"field": "name.suggest",
					"size":  limit,
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, fmt.Errorf("failed to encode suggest query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{e.indexName},
		Body:  &buf,
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute suggest: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("suggest failed: %s", res.String())
	}

	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode suggest result: %w", err)
	}

	var suggestions []string
	if suggest, ok := result["suggest"].(map[string]any); ok {
		if eventSuggestions, ok := suggest["event_suggestions"].([]any); ok {
			for _, suggestion := range eventSuggestions {
				s := suggestion.(map[string]any)
				if options, ok := s["options"].([]any); ok {
					for _, option := range options {
						o := option.(map[string]any)
						suggestions = append(suggestions, o["text"].(string))
					}
				}
			}
		}
	}

	return suggestions, nil
}

func (e *ElasticsearchClient) BulkIndex(ctx context.Context, events []EventDocument) error {
	if len(events) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, event := range events {
		meta := map[string]any{
			"index": map[string]any{
				"_index": e.indexName,
				"_id":    event.EventID.String(),
			},
		}

		metaBytes, _ := json.Marshal(meta)
		buf.Write(metaBytes)
		buf.WriteByte('\n')

		eventBytes, _ := json.Marshal(event)
		buf.Write(eventBytes)
		buf.WriteByte('\n')
	}

	req := esapi.BulkRequest{
		Body:    &buf,
		Refresh: "true",
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return fmt.Errorf("failed to bulk index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk index failed: %s", res.String())
	}

	e.logger.Info("Bulk indexed events", "count", len(events))
	return nil
}

func (e *ElasticsearchClient) DeleteIndex(ctx context.Context) error {
	req := esapi.IndicesDeleteRequest{
		Index: []string{e.indexName},
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("failed to delete index: %s", res.String())
	}

	e.logger.Info("Deleted Elasticsearch index", "index", e.indexName)
	return nil
}



