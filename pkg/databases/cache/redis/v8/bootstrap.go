package pkgredis

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// --- MOCK / REDIS FALSO PARA EVITAR CRASH ---
type noOpCache struct{}

func (n *noOpCache) Set(ctx context.Context, key string, value any, expiration ...time.Duration) error {
	return nil // Simulamos que guardamos éxito
}
func (n *noOpCache) Get(ctx context.Context, key string) (string, error) {
	return "", redis.Nil // Simulamos que no existe el dato (Cache Miss)
}
func (n *noOpCache) Delete(ctx context.Context, key string) error { return nil }
func (n *noOpCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, fmt.Errorf("key does not exist")
}
func (n *noOpCache) Exists(ctx context.Context, key string) (bool, error) { return false, nil }
func (n *noOpCache) LPush(ctx context.Context, key string, values ...any) error { return nil }
func (n *noOpCache) LTrim(ctx context.Context, key string, start, stop int64) error { return nil }
func (n *noOpCache) Close() {}
func (n *noOpCache) Client() *redis.Client { return nil }

// ---------------------------------------------

func Bootstrap(address, password string, dbName int) (Cache, error) {
	fmt.Println("🔎 [DEBUG] Iniciando Bootstrap de Redis...")

	// 1. Intentar leer variable de entorno
	if address == "" {
		address = os.Getenv("REDIS_ADDRESS")
	}

	// 🚨 AQUÍ ESTÁ LA MAGIA 🚨
	// Si no hay dirección, en lugar de devolver 'nil' (que causa pánico),
	// devolvemos el 'noOpCache'. La app pensará que Redis funciona.
	if address == "" {
		fmt.Println("⚠️ [DEBUG] No hay REDIS_ADDRESS. Usando 'Redis Mock' (No-Op) para evitar crash.")
		return &noOpCache{}, nil
	}
	// ------------------------

	fmt.Println("🚀 [DEBUG] Configurando Redis REAL en: " + address)

	if password == "" {
		password = os.Getenv("REDIS_PASSWORD")
	}
	if dbName == 0 {
		dbName, _ = strconv.Atoi(os.Getenv("REDIS_DB"))
	}

	config := newConfig(
		address,
		password,
		dbName,
	)

	if err := config.Validate(); err != nil {
		fmt.Printf("❌ [DEBUG] Error validando Redis: %v\n", err)
		return nil, err
	}

	return NewCache(config)
}