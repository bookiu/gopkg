package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockSerializer struct {
	serializeFunc   func(key string, data interface{}) ([]byte, error)
	deserializeFunc func(key string, data any) (any, error)
}

func (m *mockSerializer) Serialize(key string, data interface{}) ([]byte, error) {
	if m.serializeFunc != nil {
		return m.serializeFunc(key, data)
	}
	if v, ok := data.([]byte); ok {
		return v, nil
	}
	if v, ok := data.(string); ok {
		return []byte(v), nil
	}
	return nil, fmt.Errorf("unsupported type")
}

func (m *mockSerializer) Deserialize(key string, data any) (any, error) {
	if m.deserializeFunc != nil {
		return m.deserializeFunc(key, data)
	}
	return data, nil
}

func TestNewLocalCache(t *testing.T) {
	cache := NewLocalCache(nil)
	assert.NotNil(t, cache)

	lc, ok := cache.(*localCache)
	assert.True(t, ok)
	assert.NotNil(t, lc.cleanupStop)
}

func TestLocalCache_Set_Get(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		serializer Serializer
		key        string
		value      any
		ttl        time.Duration
		wantErr    bool
	}{
		{
			name:       "set and get byte slice without serializer",
			serializer: nil,
			key:        "test-key",
			value:      []byte("test-value"),
			ttl:        time.Hour,
			wantErr:    false,
		},
		{
			name: "set and get with serializer",
			serializer: &mockSerializer{
				serializeFunc: func(key string, data interface{}) ([]byte, error) {
					return []byte(fmt.Sprintf("%v", data)), nil
				},
				deserializeFunc: func(key string, data any) (any, error) {
					return string(data.([]byte)), nil
				},
			},
			key:     "test-key-2",
			value:   "test-value-2",
			ttl:     time.Hour,
			wantErr: false,
		},
		{
			name:       "set with zero ttl",
			serializer: nil,
			key:        "test-key-3",
			value:      []byte("test-value-3"),
			ttl:        0,
			wantErr:    false,
		},
		{
			name:       "set non-byte value without serializer",
			serializer: nil,
			key:        "test-key-4",
			value:      "string-value",
			ttl:        time.Hour,
			wantErr:    true,
		},
		{
			name: "set with serializer error",
			serializer: &mockSerializer{
				serializeFunc: func(key string, data interface{}) ([]byte, error) {
					return nil, fmt.Errorf("serialization error")
				},
			},
			key:     "test-key-5",
			value:   "test-value-5",
			ttl:     time.Hour,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewLocalCache(tt.serializer)

			err := cache.Set(ctx, tt.key, tt.value, tt.ttl)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Verify Get works
			val, err := cache.Get(ctx, tt.key)
			assert.NoError(t, err)
			assert.NotNil(t, val)
		})
	}
}

func TestLocalCache_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("get non-existent key", func(t *testing.T) {
		cache := NewLocalCache(nil)
		val, err := cache.Get(ctx, "non-existent")
		assert.Error(t, err)
		assert.Nil(t, val)
		assert.IsType(t, &KeyNotExistsError{}, err)
	})

	t.Run("get expired key", func(t *testing.T) {
		cache := NewLocalCache(nil)
		key := "expired-key"
		value := []byte("expired-value")

		err := cache.Set(ctx, key, value, 10*time.Millisecond)
		assert.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		val, err := cache.Get(ctx, key)
		assert.Error(t, err)
		assert.Nil(t, val)
		assert.IsType(t, &KeyNotExistsError{}, err)
	})

	t.Run("get valid key", func(t *testing.T) {
		cache := NewLocalCache(nil)
		key := "valid-key"
		value := []byte("valid-value")

		err := cache.Set(ctx, key, value, time.Hour)
		assert.NoError(t, err)

		val, err := cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, val)
	})
}

func TestLocalCache_Del(t *testing.T) {
	ctx := context.Background()
	cache := NewLocalCache(nil)

	key := "test-key"
	value := []byte("test-value")

	err := cache.Set(ctx, key, value, time.Hour)
	assert.NoError(t, err)

	err = cache.Del(ctx, key)
	assert.NoError(t, err)

	val, err := cache.Get(ctx, key)
	assert.Error(t, err)
	assert.Nil(t, val)
}

func TestLocalCache_Has(t *testing.T) {
	ctx := context.Background()
	cache := NewLocalCache(nil)

	t.Run("has existing key", func(t *testing.T) {
		key := "exists-key"
		value := []byte("exists-value")

		err := cache.Set(ctx, key, value, time.Hour)
		assert.NoError(t, err)

		has, err := cache.Has(ctx, key)
		assert.NoError(t, err)
		assert.True(t, has)
	})

	t.Run("has non-existing key", func(t *testing.T) {
		has, err := cache.Has(ctx, "non-exists-key")
		assert.NoError(t, err)
		assert.False(t, has)
	})
}

func TestLocalCache_SetWithFunc(t *testing.T) {
	ctx := context.Background()

	t.Run("successful set with func", func(t *testing.T) {
		cache := NewLocalCache(nil)
		key := "func-key"
		expectedValue := []byte("func-value")

		fn := func() (any, error) {
			return expectedValue, nil
		}

		val, err := cache.SetWithFunc(ctx, key, fn, time.Hour)
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, val)

		// Verify it was cached
		cachedVal, err := cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, cachedVal)
	})

	t.Run("func returns error", func(t *testing.T) {
		cache := NewLocalCache(nil)
		key := "error-func-key"
		expectedErr := fmt.Errorf("function error")

		fn := func() (any, error) {
			return nil, expectedErr
		}

		val, err := cache.SetWithFunc(ctx, key, fn, time.Hour)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, val)
	})

	t.Run("set fails after func succeeds", func(t *testing.T) {
		serializer := &mockSerializer{
			serializeFunc: func(key string, data interface{}) ([]byte, error) {
				return nil, fmt.Errorf("serialization error")
			},
		}
		cache := NewLocalCache(serializer)
		key := "set-error-key"

		fn := func() (any, error) {
			return "value", nil
		}

		val, err := cache.SetWithFunc(ctx, key, fn, time.Hour)
		assert.Error(t, err)
		assert.Nil(t, val)
	})
}

func TestLocalCache_Serialize(t *testing.T) {
	t.Run("serialize with nil serializer and byte slice", func(t *testing.T) {
		cache := &localCache{serializer: nil}
		data := []byte("test")
		result, err := cache.serialize("key", data)
		assert.NoError(t, err)
		assert.Equal(t, data, result)
	})

	t.Run("serialize with nil serializer and non-byte slice", func(t *testing.T) {
		cache := &localCache{serializer: nil}
		data := "test"
		result, err := cache.serialize("key", data)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("serialize with serializer", func(t *testing.T) {
		serializer := &mockSerializer{
			serializeFunc: func(key string, data interface{}) ([]byte, error) {
				return []byte("serialized"), nil
			},
		}
		cache := &localCache{serializer: serializer}
		result, err := cache.serialize("key", "data")
		assert.NoError(t, err)
		assert.Equal(t, []byte("serialized"), result)
	})
}

func TestLocalCache_Deserialize(t *testing.T) {
	t.Run("deserialize with nil serializer", func(t *testing.T) {
		cache := &localCache{serializer: nil}
		data := []byte("test")
		result, err := cache.deserialize("key", data)
		assert.NoError(t, err)
		assert.Equal(t, data, result)
	})

	t.Run("deserialize with serializer", func(t *testing.T) {
		serializer := &mockSerializer{
			deserializeFunc: func(key string, data any) (any, error) {
				return "deserialized", nil
			},
		}
		cache := &localCache{serializer: serializer}
		result, err := cache.deserialize("key", []byte("data"))
		assert.NoError(t, err)
		assert.Equal(t, "deserialized", result)
	})
}

func TestLocalCache_CleanupExpired(t *testing.T) {
	ctx := context.Background()
	cache := NewLocalCache(nil)

	// Set some keys with short TTL
	err := cache.Set(ctx, "short-ttl-1", []byte("value1"), 100*time.Millisecond)
	assert.NoError(t, err)

	err = cache.Set(ctx, "short-ttl-2", []byte("value2"), 100*time.Millisecond)
	assert.NoError(t, err)

	err = cache.Set(ctx, "long-ttl", []byte("value3"), time.Hour)
	assert.NoError(t, err)

	// Wait for keys to expire
	time.Sleep(150 * time.Millisecond)

	// Try to get expired keys - they should be detected as expired
	_, err1 := cache.Get(ctx, "short-ttl-1")
	_, err2 := cache.Get(ctx, "short-ttl-2")
	_, err3 := cache.Get(ctx, "long-ttl")

	// Short TTL keys should return KeyNotExistsError
	assert.Error(t, err1)
	assert.IsType(t, &KeyNotExistsError{}, err1)
	assert.Error(t, err2)
	assert.IsType(t, &KeyNotExistsError{}, err2)

	// Long TTL key should still be valid
	assert.NoError(t, err3)
}

func TestLocalCache_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	cache := NewLocalCache(nil)

	// Test concurrent writes and reads
	done := make(chan bool)
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			key := fmt.Sprintf("concurrent-key-%d", id)
			value := []byte(fmt.Sprintf("value-%d", id))

			err := cache.Set(ctx, key, value, time.Hour)
			assert.NoError(t, err)

			val, err := cache.Get(ctx, key)
			assert.NoError(t, err)
			assert.Equal(t, value, val)

			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
