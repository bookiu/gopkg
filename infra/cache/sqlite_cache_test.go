package cache

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createTempDB(t *testing.T) string {
	tmpDir := t.TempDir()
	return filepath.Join(tmpDir, "test.db")
}

func TestNewSqliteCache(t *testing.T) {
	dbPath := createTempDB(t)

	cache, err := NewSqliteCache(dbPath)
	assert.NoError(t, err)
	assert.NotNil(t, cache)
	assert.NotNil(t, cache.db)

	defer cache.Close()
}

func TestNewSqliteCache_InvalidPath(t *testing.T) {
	// Use an invalid path that will fail
	cache, err := NewSqliteCache("/invalid/path/to/db.db")
	if err == nil {
		defer cache.Close()
	}
	// The behavior depends on sqlite3 driver, it might create directories or fail
	// So we just check that we get a cache or an error, not both nil
	if cache == nil {
		assert.Error(t, err)
	}
}

func TestSqliteCache_Set_Get(t *testing.T) {
	ctx := context.Background()
	dbPath := createTempDB(t)
	cache, err := NewSqliteCache(dbPath)
	assert.NoError(t, err)
	defer cache.Close()

	tests := []struct {
		name    string
		key     string
		value   string
		ttl     time.Duration
		wantErr bool
	}{
		{
			name:    "set and get with ttl",
			key:     "test-key-1",
			value:   "test-value-1",
			ttl:     time.Hour,
			wantErr: false,
		},
		{
			name:    "set and get without ttl",
			key:     "test-key-2",
			value:   "test-value-2",
			ttl:     0,
			wantErr: false,
		},
		{
			name:    "set empty value",
			key:     "test-key-3",
			value:   "",
			ttl:     time.Hour,
			wantErr: false,
		},
		{
			name:    "update existing key",
			key:     "test-key-1",
			value:   "updated-value",
			ttl:     time.Hour,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.Set(ctx, tt.key, tt.value, tt.ttl)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			val, err := cache.Get(ctx, tt.key)
			assert.NoError(t, err)
			assert.Equal(t, tt.value, val)
		})
	}
}

func TestSqliteCache_Get(t *testing.T) {
	ctx := context.Background()
	dbPath := createTempDB(t)
	cache, err := NewSqliteCache(dbPath)
	assert.NoError(t, err)
	defer cache.Close()

	t.Run("get non-existent key", func(t *testing.T) {
		val, err := cache.Get(ctx, "non-existent")
		assert.Error(t, err)
		assert.IsType(t, &KeyNotExistsError{}, err)
		assert.Equal(t, "", val)
	})

	t.Run("get expired key", func(t *testing.T) {
		key := "expired-key"
		value := "expired-value"

		err := cache.Set(ctx, key, value, 1*time.Second)
		assert.NoError(t, err)

		time.Sleep(2 * time.Second)

		val, err := cache.Get(ctx, key)
		assert.Error(t, err)
		assert.IsType(t, &KeyNotExistsError{}, err)
		assert.Equal(t, "", val)
	})

	t.Run("get valid key", func(t *testing.T) {
		key := "valid-key"
		value := "valid-value"

		err := cache.Set(ctx, key, value, time.Hour)
		assert.NoError(t, err)

		val, err := cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, val)
	})

	t.Run("get key with zero ttl", func(t *testing.T) {
		key := "zero-ttl-key"
		value := "zero-ttl-value"

		err := cache.Set(ctx, key, value, 0)
		assert.NoError(t, err)

		val, err := cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, val)
	})
}

func TestSqliteCache_Del(t *testing.T) {
	ctx := context.Background()
	dbPath := createTempDB(t)
	cache, err := NewSqliteCache(dbPath)
	assert.NoError(t, err)
	defer cache.Close()

	t.Run("delete existing key", func(t *testing.T) {
		key := "delete-key"
		value := "delete-value"

		err := cache.Set(ctx, key, value, time.Hour)
		assert.NoError(t, err)

		err = cache.Del(ctx, key)
		assert.NoError(t, err)

		val, err := cache.Get(ctx, key)
		assert.Error(t, err)
		assert.IsType(t, &KeyNotExistsError{}, err)
		assert.Equal(t, "", val)
	})

	t.Run("delete non-existent key", func(t *testing.T) {
		err := cache.Del(ctx, "non-existent-key")
		assert.NoError(t, err)
	})
}

func TestSqliteCache_Has(t *testing.T) {
	ctx := context.Background()
	dbPath := createTempDB(t)
	cache, err := NewSqliteCache(dbPath)
	assert.NoError(t, err)
	defer cache.Close()

	t.Run("has existing key", func(t *testing.T) {
		key := "exists-key"
		value := "exists-value"

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

	t.Run("has expired key", func(t *testing.T) {
		key := "expired-key-has"
		value := "expired-value"

		err := cache.Set(ctx, key, value, 1*time.Second)
		assert.NoError(t, err)

		time.Sleep(2 * time.Second)

		has, err := cache.Has(ctx, key)
		assert.NoError(t, err)
		assert.False(t, has)
	})

	t.Run("has key with zero ttl", func(t *testing.T) {
		key := "zero-ttl-has-key"
		value := "zero-ttl-value"

		err := cache.Set(ctx, key, value, 0)
		assert.NoError(t, err)

		has, err := cache.Has(ctx, key)
		assert.NoError(t, err)
		assert.True(t, has)
	})
}

func TestSqliteCache_SetWithFunc(t *testing.T) {
	ctx := context.Background()
	dbPath := createTempDB(t)
	cache, err := NewSqliteCache(dbPath)
	assert.NoError(t, err)
	defer cache.Close()

	t.Run("successful set with func", func(t *testing.T) {
		key := "func-key"
		expectedValue := "func-value"

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
		key := "error-func-key"
		expectedErr := fmt.Errorf("function error")

		fn := func() (any, error) {
			return nil, expectedErr
		}

		val, err := cache.SetWithFunc(ctx, key, fn, time.Hour)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, "", val)
	})

	t.Run("set with func and zero ttl", func(t *testing.T) {
		key := "zero-ttl-func-key"
		expectedValue := "zero-ttl-func-value"

		fn := func() (any, error) {
			return expectedValue, nil
		}

		val, err := cache.SetWithFunc(ctx, key, fn, 0)
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, val)
	})
}

func TestSqliteCache_CleanupExpiredKeys(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running test in short mode")
	}

	ctx := context.Background()
	dbPath := createTempDB(t)
	cache, err := NewSqliteCache(dbPath)
	assert.NoError(t, err)
	defer cache.Close()

	// Set some keys with short TTL
	err = cache.Set(ctx, "short-ttl-1", "value1", 1*time.Second)
	assert.NoError(t, err)

	err = cache.Set(ctx, "short-ttl-2", "value2", 1*time.Second)
	assert.NoError(t, err)

	err = cache.Set(ctx, "long-ttl", "value3", time.Hour)
	assert.NoError(t, err)

	// Wait for keys to expire and cleanup to run (cleanup runs every minute)
	time.Sleep(70 * time.Second)

	// Short TTL keys should be removed by cleanup
	has1, _ := cache.Has(ctx, "short-ttl-1")
	has2, _ := cache.Has(ctx, "short-ttl-2")
	hasLong, _ := cache.Has(ctx, "long-ttl")

	assert.False(t, has1)
	assert.False(t, has2)
	assert.True(t, hasLong)
}

func TestSqliteCache_Close(t *testing.T) {
	dbPath := createTempDB(t)
	cache, err := NewSqliteCache(dbPath)
	assert.NoError(t, err)

	err = cache.Close()
	assert.NoError(t, err)

	// Trying to use cache after close should fail
	ctx := context.Background()
	err = cache.Set(ctx, "key", "value", time.Hour)
	assert.Error(t, err)
}

func TestSqliteCache_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	dbPath := createTempDB(t)
	cache, err := NewSqliteCache(dbPath)
	assert.NoError(t, err)
	defer cache.Close()

	// Test concurrent writes and reads
	done := make(chan bool)
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			key := fmt.Sprintf("concurrent-key-%d", id)
			value := fmt.Sprintf("value-%d", id)

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

func TestSqliteCache_PersistenceAcrossInstances(t *testing.T) {
	ctx := context.Background()
	dbPath := createTempDB(t)

	// Create first cache instance and set a value
	cache1, err := NewSqliteCache(dbPath)
	assert.NoError(t, err)

	key := "persistent-key"
	value := "persistent-value"
	err = cache1.Set(ctx, key, value, time.Hour)
	assert.NoError(t, err)
	cache1.Close()

	// Create second cache instance and verify value persists
	cache2, err := NewSqliteCache(dbPath)
	assert.NoError(t, err)
	defer cache2.Close()

	val, err := cache2.Get(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, value, val)
}

func TestSqliteCache_LargeValue(t *testing.T) {
	ctx := context.Background()
	dbPath := createTempDB(t)
	cache, err := NewSqliteCache(dbPath)
	assert.NoError(t, err)
	defer cache.Close()

	// Create a large value
	largeValue := string(make([]byte, 1024*1024)) // 1MB
	key := "large-key"

	err = cache.Set(ctx, key, largeValue, time.Hour)
	assert.NoError(t, err)

	val, err := cache.Get(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, largeValue, val)
}

func TestSqliteCache_SpecialCharactersInKey(t *testing.T) {
	ctx := context.Background()
	dbPath := createTempDB(t)
	cache, err := NewSqliteCache(dbPath)
	assert.NoError(t, err)
	defer cache.Close()

	tests := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "key with spaces",
			key:   "key with spaces",
			value: "value1",
		},
		{
			name:  "key with special chars",
			key:   "key:with:colons",
			value: "value2",
		},
		{
			name:  "key with unicode",
			key:   "键值",
			value: "value3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.Set(ctx, tt.key, tt.value, time.Hour)
			assert.NoError(t, err)

			val, err := cache.Get(ctx, tt.key)
			assert.NoError(t, err)
			assert.Equal(t, tt.value, val)
		})
	}
}
