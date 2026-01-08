package pkgredis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// Cache define los métodos para interactuar con Redis
type Cache interface {
	Set(ctx context.Context, key string, value any, expiration ...time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Exists(ctx context.Context, key string) (bool, error)
	LPush(ctx context.Context, key string, values ...any) error
	LTrim(ctx context.Context, key string, start, stop int64) error
	Close()
	Client() *redis.Client
}

// Config define los métodos que la configuración de Redis debe implementar
type Config interface {
	GetAddress() string
	GetPassword() string
	GetDB() int
	GetDefaultExpiration() time.Duration
	Validate() error
}
