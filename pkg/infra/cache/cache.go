package cache

import (
	"context"
	"errors"
	"time"
)

var ErrKeyNotFound = errors.New("key not found")

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	Has(ctx context.Context, key string) (bool, error)
	SetWithFunc(ctx context.Context, key string, fn func() (string, error), ttl time.Duration) (string, error)
}
