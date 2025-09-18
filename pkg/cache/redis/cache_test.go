package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisInstanceWithInvalidConfig(t *testing.T) {
	redis, err := NewCache(&Config{
		Host: "fakelocalhost",
		Port: "6379",
	})

	// since redis server is not running it will return error
	assert.NotNil(t, err)
	assert.Nil(t, redis)
}

func TestNewRedisInstance_SetGet(t *testing.T) {
	// Create a miniredis server
	srv, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Error starting miniredis server: %v", err)
	}
	defer srv.Close()

	config := &Config{
		Host: srv.Host(),
		Port: srv.Port(),
	}

	cache, err := NewCache(config)
	assert.Nil(t, err)
	assert.NotNil(t, cache)

	err = cache.Set(context.Background(), "test-key", "test-set-val", time.Minute)
	assert.Nil(t, err)

	val, err := cache.Get(context.Background(), "test-key")
	assert.Nil(t, err)
	assert.Equal(t, "test-set-val", val)

	err = cache.Delete(context.Background(), "test-key")
	assert.Nil(t, err)

	val, err = cache.Get(context.Background(), "test-key")
	assert.NotNil(t, err)
	assert.Equal(t, "", val)
}

func TestRedisCache_GetLike(t *testing.T) {
	// Create a miniredis server
	srv, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Error starting miniredis server: %v", err)
	}
	defer srv.Close()

	config := &Config{
		Host: srv.Host(),
		Port: srv.Port(),
	}

	cache, err := NewCache(config)
	assert.Nil(t, err)
	assert.NotNil(t, cache)

	// Set multiple keys with a pattern
	err = cache.Set(context.Background(), "user:1", "value1", time.Minute)
	assert.Nil(t, err)
	err = cache.Set(context.Background(), "user:2", "value2", time.Minute)
	assert.Nil(t, err)
	err = cache.Set(context.Background(), "user:3", "value3", time.Minute)
	assert.Nil(t, err)
	err = cache.Set(context.Background(), "other:1", "othervalue", time.Minute)
	assert.Nil(t, err)

	// Test GetLike with pattern
	values, err := cache.GetByPattern(context.Background(), "user:*")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(values))

	// Convert to string slice for easier comparison
	var stringValues []string
	for _, v := range values {
		stringValues = append(stringValues, v.(string))
	}

	// Check that all user values are present
	assert.Contains(t, stringValues, "value1")
	assert.Contains(t, stringValues, "value2")
	assert.Contains(t, stringValues, "value3")
	assert.NotContains(t, stringValues, "othervalue")

	// Test with non-matching pattern
	values, err = cache.GetByPattern(context.Background(), "nonexistent:*")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(values))
}
