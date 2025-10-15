package inmemory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCacheInstance(t *testing.T) {
	config := &Config{
		DefaultExpiration: 15,
		CleanupInterval:   30,
	}

	mem, err := NewCache(config)
	assert.Nil(t, err)
	assert.NotNil(t, mem)
}

func TestInMemoryCache_Set(t *testing.T) {
	config := &Config{
		DefaultExpiration: 15,
		CleanupInterval:   30,
	}

	mem, err := NewCache(config)
	assert.Nil(t, err)
	assert.NotNil(t, mem)

	err = mem.Set(context.TODO(), "test-key", "test-set-val", time.Minute)
	assert.Nil(t, err)

	val, err := mem.Get(context.TODO(), "test-key")
	assert.Nil(t, err)
	assert.Equal(t, "test-set-val", val)
}

func TestInMemoryGetWithoutSet(t *testing.T) {
	config := &Config{
		DefaultExpiration: 15,
		CleanupInterval:   30,
	}

	mem, err := NewCache(config)

	assert.Nil(t, err)
	assert.NotNil(t, mem)

	val, err := mem.Get(context.Background(), "test-key")
	assert.NotNil(t, err)
	assert.Equal(t, "", val)
}

func TestInMemoryCache_Delete(t *testing.T) {
	config := &Config{
		DefaultExpiration: 15,
		CleanupInterval:   30,
	}

	mem, err := NewCache(config)
	assert.Nil(t, err)
	assert.NotNil(t, mem)

	err = mem.Set(context.TODO(), "test-key", "test-set-val", time.Minute)
	assert.Nil(t, err)

	val, err := mem.Get(context.TODO(), "test-key")
	assert.Nil(t, err)
	assert.Equal(t, "test-set-val", val)

	err = mem.Delete(context.Background(), "test-key")
	assert.Nil(t, err)

	val, err = mem.Get(context.Background(), "test-key")
	assert.NotNil(t, err)
	assert.Equal(t, "", val)
}

func TestInMemoryCacheGetByPattern(t *testing.T) {
	config := &Config{
		DefaultExpiration: 15,
		CleanupInterval:   30,
	}

	mem, err := NewCache(config)
	assert.Nil(t, err)
	assert.NotNil(t, mem)

	// Set multiple keys with a pattern
	err = mem.Set(context.Background(), "user:1", "value1", time.Minute)
	assert.Nil(t, err)
	err = mem.Set(context.Background(), "user:2", "value2", time.Minute)
	assert.Nil(t, err)
	err = mem.Set(context.Background(), "user:3", "value3", time.Minute)
	assert.Nil(t, err)
	err = mem.Set(context.Background(), "other:1", "othervalue", time.Minute)
	assert.Nil(t, err)

	// Test GetLike with pattern
	values, err := mem.GetByPattern(context.Background(), "user:*")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(values))

	// Convert to string slice for easier comparison
	stringValues := make([]string, 0, len(values))
	for _, v := range values {
		stringValues = append(stringValues, v.(string))
	}

	// Check that all user values are present
	assert.Contains(t, stringValues, "value1")
	assert.Contains(t, stringValues, "value2")
	assert.Contains(t, stringValues, "value3")
	assert.NotContains(t, stringValues, "othervalue")

	// Test with non-matching pattern
	values, err = mem.GetByPattern(context.Background(), "nonexistent:*")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(values))
}
