package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache/inmemory"
	"github.com/redhat-data-and-ai/usernaut/pkg/cache/redis"
)

var (
	// ErrInvalidCacheDriver is returned when an invalid cache driver is provided
	ErrInvalidCacheDriver = errors.New("invalid cache driver")
)

const (
	DriverMemory = "memory"
	DriverRedis  = "redis"

	NoExpiration = -1 * time.Second
)

// Cache implements a generic interface for cache clients
type Cache interface {
	// Get returns the value for the given key
	// returns the value if the key was found
	// returns an error if the key was not found
	Get(ctx context.Context, key string) (interface{}, error)

	// GetByPattern returns the value for the given key pattern
	// returns the value if the key matches the pattern
	// returns an error if the key was not found
	GetByPattern(ctx context.Context, keyPattern string) (map[string]interface{}, error)

	// Set sets the value for the given key
	// returns nil if the key was set successfully
	// returns an error if the key was not set successfully
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Delete deletes the value for the given key
	// returns nil if the key was deleted successfully
	// returns an error if the key was not deleted successfully
	Delete(ctx context.Context, key string) error
}

// Config is the configuration for the cache client
type Config struct {
	// Driver is the type of cache client
	Driver string

	// InMemory is the configuration for the inmemory cache client
	InMemory *inmemory.Config

	// Redis is the configuration for the redis client
	Redis *redis.Config
}

// New returns a new cache client
func New(config *Config) (Cache, error) {

	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	switch config.Driver {
	case DriverMemory:
		return inmemory.NewCache(config.InMemory)
	case DriverRedis:
		return redis.NewCache(config.Redis)
	default:
		return nil, ErrInvalidCacheDriver
	}
}
