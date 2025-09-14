package booking

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(redisURL string) (*RedisClient, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	rdb := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{
		client: rdb,
	}, nil
}

func (r *RedisClient) SetReservation(ctx context.Context, reservationID uuid.UUID, data *ReservationData, ttl time.Duration) error {
	key := fmt.Sprintf("booking:reservation:%s", reservationID)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal reservation data: %w", err)
	}

	return r.client.Set(ctx, key, jsonData, ttl).Err()
}

func (r *RedisClient) GetReservation(ctx context.Context, reservationID uuid.UUID) (*ReservationData, error) {
	key := fmt.Sprintf("booking:reservation:%s", reservationID)

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("reservation not found or expired")
		}
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	var data ReservationData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal reservation data: %w", err)
	}

	return &data, nil
}

func (r *RedisClient) DeleteReservation(ctx context.Context, reservationID uuid.UUID) error {
	key := fmt.Sprintf("booking:reservation:%s", reservationID)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisClient) IncrementRateLimit(ctx context.Context, userID uuid.UUID, ttl time.Duration) (int64, error) {
	key := fmt.Sprintf("booking:rate_limit:%s", userID)

	pipe := r.client.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment rate limit: %w", err)
	}

	return incrCmd.Val(), nil
}

func (r *RedisClient) CheckRateLimit(ctx context.Context, userID uuid.UUID, limit int64) (bool, error) {
	key := fmt.Sprintf("booking:rate_limit:%s", userID)

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return true, nil
		}
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}

	count, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return false, fmt.Errorf("failed to parse rate limit count: %w", err)
	}

	return count < limit, nil
}

func (r *RedisClient) SetDistributedLock(ctx context.Context, lockKey string, value string, ttl time.Duration) (bool, error) {
	key := fmt.Sprintf("booking:lock:%s", lockKey)

	result := r.client.SetNX(ctx, key, value, ttl)
	return result.Val(), result.Err()
}

func (r *RedisClient) ReleaseDistributedLock(ctx context.Context, lockKey string, value string) error {
	key := fmt.Sprintf("booking:lock:%s", lockKey)

	script := redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`)

	return script.Run(ctx, r.client, []string{key}, value).Err()
}

func (r *RedisClient) CacheEventAvailability(ctx context.Context, eventID uuid.UUID, availableSeats int32, ttl time.Duration) error {
	key := fmt.Sprintf("booking:event_availability:%s", eventID)
	return r.client.Set(ctx, key, availableSeats, ttl).Err()
}

func (r *RedisClient) GetCachedEventAvailability(ctx context.Context, eventID uuid.UUID) (int32, error) {
	key := fmt.Sprintf("booking:event_availability:%s", eventID)

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, fmt.Errorf("availability not cached")
		}
		return 0, fmt.Errorf("failed to get cached availability: %w", err)
	}

	seats, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse cached availability: %w", err)
	}

	return int32(seats), nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
