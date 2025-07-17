package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	cachefile string

	cache Cache
)

func TestMain(m *testing.M) {
	tempDir := os.TempDir()
	cachefile = filepath.Join(tempDir, "test_sqlite_cache.db")
	cache, _ = NewSqliteCache(cachefile)

	exitCode := m.Run()

	os.Remove(cachefile)
	os.Exit(exitCode)
}

func TestSqliteCache(t *testing.T) {
	err := cache.Set(context.Background(), "key", "value", time.Second)
	if err != nil {
		t.Fatal(err)
	}

	val, err := cache.Get(context.Background(), "key")
	if err != nil {
		t.Fatal(err)
	}

	if val != "value" {
		t.Fatal("value not equal")
	}
}

func TestKeyExpire(t *testing.T) {
	key := "key_expired"
	err := cache.Set(context.Background(), key, "value", time.Second)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	val, err := cache.Get(context.Background(), key)
	if err != ErrKeyNotFound && val != "" {
		t.Fatal("key should be expired")
	}
}

func TestKeyNotExist(t *testing.T) {
	key := "key_not_exist"
	val, err := cache.Get(context.Background(), key)
	if err != ErrKeyNotFound && val != "" {
		t.Fatal("key should not exist")
	}
}

func TestKeyGetWithFunc(t *testing.T) {
	key := "key_get_with_func"
	expectedVal := "the test value"

	val, err := cache.SetWithFunc(context.Background(), key, func() (string, error) {
		return expectedVal, nil
	}, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if val != expectedVal {
		t.Fatal("value not equal")
	}
}
