package inmemory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCacheInstance(t *testing.T) {
	config := &Config{
		DefaultExpiration: 15 * time.Second,
		CleanupInterval:   30 * time.Minute,
	}

	mem, err := NewCache(config)
	assert.Nil(t, err)
	assert.NotNil(t, mem)
}

func TestInMemoryCache_Set(t *testing.T) {
	config := &Config{
		DefaultExpiration: 15 * time.Second,
		CleanupInterval:   30 * time.Minute,
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
		DefaultExpiration: 15 * time.Second,
		CleanupInterval:   30 * time.Minute,
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
		DefaultExpiration: 15 * time.Second,
		CleanupInterval:   30 * time.Minute,
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
