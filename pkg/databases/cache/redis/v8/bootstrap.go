package pkgredis

import (
	"fmt" // <--- No olvides importar esto
	"os"
	"strconv"
)

func Bootstrap(address, password string, dbName int) (Cache, error) {
	fmt.Println("🔎 [DEBUG] Iniciando Bootstrap de Redis...") 

	if address == "" {
		address = os.Getenv("REDIS_ADDRESS")
	}

	// Tu parche actual
	if address == "" {
		fmt.Println("⚠️ [DEBUG] Redis Address vacía. Retornando NIL (Modo Sin Caché).") 
		return nil, nil
	}

	fmt.Println("🚀 [DEBUG] Configurando Redis con dirección: " + address)

	if password == "" {
		password = os.Getenv("REDIS_PASSWORD")
	}
	if dbName == 0 {
		dbName, _ = strconv.Atoi(os.Getenv("REDIS_DB"))
	}

	config := newConfig(address, password, dbName)

	if err := config.Validate(); err != nil {
		fmt.Printf("❌ [DEBUG] Error validando Redis: %v\n", err)
		return nil, err
	}

	return NewCache(config)
}