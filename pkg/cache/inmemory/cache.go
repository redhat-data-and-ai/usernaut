package inmemory

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

// InMemoryCache holds the handler for the in-memory cache using go-cache
type InMemoryCache struct {
	client *gocache.Cache
}

// Config is the configuration for the in-memory cache
type Config struct {
	DefaultExpiration int32
	CleanupInterval   int32
}

// InMemoryCacheConfig is the configuration for the in-memory cache
func NewCache(config *Config) (*InMemoryCache, error) {
	if config == nil {
		config = getDefaultConfig()
	}

	defaultExpiration := time.Duration(config.DefaultExpiration) * time.Second
	cleanupExpiration := time.Duration(config.CleanupInterval) * time.Second

	client := gocache.New(defaultExpiration, cleanupExpiration)

	inMem := &InMemoryCache{
		client: client,
	}

	return inMem, nil
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

// ScanKeys returns all keys matching the given pattern from in-memory cache
func (imc *InMemoryCache) ScanKeys(ctx context.Context, pattern string) ([]string, error) {
	items := imc.client.Items()
	var keys []string

	for key := range items {
		matched, err := filepath.Match(pattern, key)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %s: %w", pattern, err)
		}
		if matched {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// Flushes out all the keys from Cache.
func (imc *InMemoryCache) Flush(ctx context.Context) {
	imc.client.Flush()
}

// getDefaultConfig returns the default configuration for the in-memory cache
func getDefaultConfig() *Config {
	return &Config{
		DefaultExpiration: -1,
		CleanupInterval:   -1,
	}
}
