package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

// Config holds all required info for initializing redis driver
type Config struct {
	Host     string
	Port     string
	Database int32
	Username string
	Password string
}

// RedisCache holds the handler for the redisclient and auxiliary info
type RedisCache struct {
	client redis.UniversalClient
}

// NewRedisClient inits a RedisCache instance
func NewCache(config *Config) (*RedisCache, error) {
	if config == nil {
		config = getDefaultConfig()
	}

	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	options := &redis.UniversalOptions{
		Addrs:    []string{addr},
		Username: config.Username,
		Password: config.Password,
		DB:       int(config.Database),
	}

	redisClient := redis.NewUniversalClient(options)

	// Enable OpenTelemetry instrumentation
	if err := redisotel.InstrumentTracing(redisClient); err != nil {
		return nil, fmt.Errorf("failed to instrument redis: %w", err)
	}
	if err := redisotel.InstrumentMetrics(redisClient); err != nil {
		return nil, fmt.Errorf("failed to instrument redis metrics: %w", err)
	}

	rc := RedisCache{
		client: redisClient,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := rc.client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return &rc, nil
}

func getDefaultConfig() *Config {
	return &Config{
		Username: "",
		Host:     "localhost",
		Port:     "6379",
		Database: 0,
		Password: "",
	}
}

// Set - sets a key value pair in redis
func (rc *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return rc.client.Set(ctx, key, value, ttl).Err()
}

// Get - gets a value from redis
func (rc *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	val, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (rc *RedisCache) GetByPattern(ctx context.Context, keyPattern string) (map[string]interface{}, error) {
	// First, collect all keys matching the pattern
	var keys []string
	iter := rc.client.Scan(ctx, 0, keyPattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	// If no keys found, return empty map
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	// Use MGET to retrieve all values in a single round trip
	vals, err := rc.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	// Build the result map, handling nil values (expired keys)
	values := make(map[string]interface{}, len(keys))
	for i, key := range keys {
		if vals[i] != nil {
			values[key] = vals[i]
		}
		// Skip nil values (keys that expired between SCAN and MGET)
	}

	return values, nil
}

// Delete - deletes a key from redis
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	return rc.client.Del(ctx, key).Err()
}

// Disconnect ... disconnects from the redis server
func (rc *RedisCache) Disconnect() error {
	err := rc.client.Close()
	if err != nil {
		return err
	}
	return nil
}
