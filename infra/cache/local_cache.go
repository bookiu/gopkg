package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// entry represents a cached query result
type entry struct {
	Value      interface{}
	ExpireTime time.Time
}

// localCache manages the query result cache
type localCache struct {
	cleanupTicker *time.Ticker
	cleanupStop   chan struct{}

	store      sync.Map
	serializer Serializer
}

// NewLocalCache creates a new query cache with specified default TTL
func NewLocalCache(serializer Serializer) Cache {
	cache := &localCache{
		cleanupStop: make(chan struct{}),
		serializer:  serializer,
	}

	// Start cleanup goroutine if cache is enabled
	go cache.cleanupExpired()

	return cache
}

// Del delete a value from cache
func (c *localCache) Del(ctx context.Context, key string) error {
	c.store.Delete(key)
	return nil
}

// Get retrieves a value from cache
func (c *localCache) Get(ctx context.Context, key string) (any, error) {
	val, exists := c.store.Load(key)
	if !exists {
		return nil, newKeyNotExistsError(key)
	}
	e := val.(*entry)

	// Check if entry has expired
	if time.Now().After(e.ExpireTime) {
		_ = c.Del(ctx, key)
		return nil, newKeyNotExistsError(key)
	}

	deserialized, err := c.deserialize(key, e.Value)
	if err != nil {
		return nil, err
	}
	return deserialized, nil
}

// Set stores a value in cache with custom TTL
func (c *localCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	v, err := c.serialize(key, value)
	if err != nil {
		return err
	}
	if ttl == 0 {
		ttl = time.Hour * 24 * 365 * 99
	}

	c.store.Store(key, &entry{
		Value:      v,
		ExpireTime: time.Now().Add(ttl),
	})
	return nil
}

// Has checks if a key exists in cache
func (c *localCache) Has(ctx context.Context, key string) (bool, error) {
	_, ok := c.store.Load(key)
	if ok {
		return true, nil
	}
	return false, nil
}

// SetWithFunc is to set cache key which value from function return
func (c *localCache) SetWithFunc(ctx context.Context, key string, fn func() (any, error), ttl time.Duration) (any, error) {
	value, err := fn()
	if err != nil {
		return nil, err
	}
	err = c.Set(ctx, key, value, ttl)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (c *localCache) serialize(key string, value any) ([]byte, error) {
	if c.serializer == nil {
		if v, ok := value.([]byte); ok {
			return v, nil
		}
		return nil, fmt.Errorf("cache serializer is nil and value is not []byte type")
	}

	return c.serializer.Serialize(key, value)
}

func (c *localCache) deserialize(key string, data any) (any, error) {
	if c.serializer == nil {
		return data, nil
	}
	return c.serializer.Deserialize(key, data)
}

// cleanupExpired periodically removes expired entries
func (c *localCache) cleanupExpired() {
	// Cleanup interval: use defaultTTL or minimum 1 minute
	interval := time.Minute

	c.cleanupTicker = time.NewTicker(interval)
	defer c.cleanupTicker.Stop()

	for {
		select {
		case <-c.cleanupTicker.C:
			now := time.Now()
			c.store.Range(func(k, v interface{}) bool {
				if now.After(v.(*entry).ExpireTime) {
					_ = c.Del(context.Background(), k.(string))
				}
				return true
			})
		case <-c.cleanupStop:
			return
		}
	}
}
