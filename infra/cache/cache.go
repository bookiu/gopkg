package cache

import (
	"context"
	"errors"
	"time"
)

var ErrKeyNotFound = errors.New("key not found")

type Cache interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	Has(ctx context.Context, key string) (bool, error)
	SetWithFunc(ctx context.Context, key string, fn func() (any, error), ttl time.Duration) (any, error)
}

type Serializer interface {
	Serialize(key string, data interface{}) ([]byte, error)
	Deserialize(key string, data any) (any, error)
}
