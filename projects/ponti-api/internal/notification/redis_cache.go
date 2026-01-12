package notification

import (
	"context"
	"time"

	redis "github.com/alphacodinggroup/ponti-backend/pkg/databases/cache/redis/v8"
)

type cache struct {
	cache redis.Cache
}

func NewRedisCache(c redis.Cache) Cache {
	return &cache{
		cache: c,
	}
}

func (c *cache) StoreRefreshToken(ctx context.Context, userID, refreshToken string, expiration time.Time) error {
	return c.cache.Set(ctx, userID, refreshToken, time.Until(expiration))
}

func (c *cache) RetrieveRefreshToken(ctx context.Context, userID string) (string, error) {
	return c.cache.Get(ctx, userID)
}

func (c *cache) Close() {
	c.cache.Close()
}
