package inmemory

import (
	"context"
	"fmt"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

// InMemoryCache holds the handler for the in-memory cache using go-cache
type InMemoryCache struct {
	client *gocache.Cache
}

// Config is the configuration for the in-memory cache
type Config struct {
	DefaultExpiration time.Duration
	CleanupInterval   time.Duration
}

// InMemoryCacheConfig is the configuration for the in-memory cache
func NewCache(config *Config) (*InMemoryCache, error) {
	if config == nil {
		config = getDefaultConfig()
	}

	client := gocache.New(config.DefaultExpiration, config.CleanupInterval)

	imc := InMemoryCache{
		client: client,
	}

	return &imc, nil
}

// Set implements inmemory
func (imc *InMemoryCache) Set(
	ctx context.Context,
	key string,
	value string,
	ttl time.Duration,
) error {
	imc.client.Set(key, value, ttl)
	return nil
}

// Get implements Cache.
func (imc *InMemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	val, found := imc.client.Get(key)
	if !found {
		return "", fmt.Errorf("key not found")
	}
	return val, nil
}

// Delete implements Cache.
func (imc *InMemoryCache) Delete(ctx context.Context, key string) error {
	_, found := imc.client.Get(key)
	if found {
		imc.client.Delete(key)
	}
	return nil
}

// getDefaultConfig returns the default configuration for the in-memory cache
func getDefaultConfig() *Config {
	return &Config{
		DefaultExpiration: 5 * time.Minute,
		CleanupInterval:   10 * time.Minute,
	}
}
