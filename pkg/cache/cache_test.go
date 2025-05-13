package cache

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redhat-data-and-ai/usernaut/pkg/cache/inmemory"
	"github.com/redhat-data-and-ai/usernaut/pkg/cache/redis"
	"github.com/stretchr/testify/assert"
)

func TestNewInMemoryCacheInstance(t *testing.T) {
	config := Config{
		Driver: "memory",
		InMemory: &inmemory.Config{
			DefaultExpiration: 15 * time.Second,
			CleanupInterval:   30 * time.Second,
		},
	}

	mem, err := New(&config)
	assert.Nil(t, err)
	assert.NotNil(t, mem)
}

func TestNewRedisInstance(t *testing.T) {
	srv, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Error starting miniredis server: %v", err)
	}
	defer srv.Close()

	config := &Config{
		Driver: "redis",
		Redis: &redis.Config{
			Host: srv.Host(),
			Port: srv.Port(),
		},
	}

	client, err := New(config)
	assert.Nil(t, err)
	assert.NotNil(t, client)
}
