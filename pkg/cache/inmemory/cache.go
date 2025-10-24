package inmemory

import (
	"context"
	"fmt"
	"regexp"
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

// GetByPattern like implements Cache.
func (imc *InMemoryCache) GetByPattern(ctx context.Context, keyPattern string) (map[string]interface{}, error) {
	keys, err := imc.ScanKeys(ctx, keyPattern)
	if err != nil {
		return nil, fmt.Errorf("error scanning keys: %w", err)
	}

	values := make(map[string]interface{})
	for _, key := range keys {
		val, found := imc.client.Get(key)
		if found {
			values[key] = val
		}
	}
	return values, nil
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
// Pattern is a glob pattern (like Redis SCAN), where * matches any sequence of characters
// and ? matches any single character
func (imc *InMemoryCache) ScanKeys(ctx context.Context, pattern string) ([]string, error) {
	items := imc.client.Items()
	var keys []string

	// Convert glob pattern to regex pattern
	regexPattern := globToRegex(pattern)
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern %s: %w", pattern, err)
	}

	for key := range items {
		if regex.MatchString(key) {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// globToRegex converts a glob pattern to a regex pattern
// * matches any sequence of characters
// ? matches any single character
// All other regex special characters are escaped
func globToRegex(glob string) string {
	var result string
	for i := 0; i < len(glob); i++ {
		c := glob[i]
		switch c {
		case '*':
			result += ".*"
		case '?':
			result += "."
		case '.', '+', '(', ')', '|', '[', ']', '{', '}', '^', '$', '\\':
			result += "\\" + string(c)
		default:
			result += string(c)
		}
	}
	return "^" + result + "$"
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
