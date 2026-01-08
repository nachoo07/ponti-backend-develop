package pkgredis

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	instance *cache
	once     sync.Once
)

type cache struct {
	client            *redis.Client
	defaultExpiration time.Duration
}

func NewCache(c Config) (Cache, error) {
	if c.GetAddress() == "" {
		return nil, errors.New("redis address is required")
	}

	var err error
	once.Do(func() {
		instance = &cache{
			defaultExpiration: c.GetDefaultExpiration(),
		}
		err = instance.connect(c)
	})

	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (ch *cache) connect(c Config) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.GetAddress(),
		Password: c.GetPassword(),
		DB:       c.GetDB(),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	ch.client = rdb
	return nil
}

// Set almacena un valor en Redis con una clave y un tiempo de expiración opcional
func (ch *cache) Set(ctx context.Context, key string, value any, expiration ...time.Duration) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}

	exp := ch.defaultExpiration
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	return ch.client.Set(ctx, key, value, exp).Err()
}

// Get recupera un valor de Redis usando una clave
func (ch *cache) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", errors.New("key cannot be empty")
	}

	result, err := ch.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", redis.Nil
	} else if err != nil {
		return "", fmt.Errorf("failed to get key: %w", err)
	}
	return result, nil
}

// Delete elimina una clave de Redis
func (ch *cache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}

	_, err := ch.client.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}
	return nil
}

// TTL obtiene el tiempo de expiración restante de una clave
func (ch *cache) TTL(ctx context.Context, key string) (time.Duration, error) {
	if key == "" {
		return 0, errors.New("key cannot be empty")
	}

	ttl, err := ch.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}
	if ttl == -1 {
		return 0, fmt.Errorf("key exists but has no expiration")
	}
	if ttl == -2 {
		return 0, fmt.Errorf("key does not exist")
	}
	return ttl, nil
}

// Exists verifica si una clave existe en Redis
func (ch *cache) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, errors.New("key cannot be empty")
	}

	count, err := ch.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return count > 0, nil
}

// Close cierra la conexión con el servidor Redis
func (ch *cache) Close() {
	if ch.client != nil {
		_ = ch.client.Close()
	}
}

// Client devuelve el cliente Redis para operaciones avanzadas
func (ch *cache) Client() *redis.Client {
	return ch.client
}

// LPush inserta uno o más valores al inicio de una lista en Redis.
func (ch *cache) LPush(ctx context.Context, key string, values ...any) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}
	return ch.client.LPush(ctx, key, values...).Err()
}

// LTrim recorta una lista en Redis para conservar únicamente los elementos en el rango dado.
func (ch *cache) LTrim(ctx context.Context, key string, start, stop int64) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}
	return ch.client.LTrim(ctx, key, start, stop).Err()
}
