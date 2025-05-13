package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisInstanceWithNilConfig(t *testing.T) {
	redis, err := NewCache(nil)

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
