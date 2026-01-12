package pkgredis

import (
	"os"
	"strconv"
)

func Bootstrap(address, password string, dbName int) (Cache, error) {
	if address == "" {
		address = os.Getenv("REDIS_ADDRESS")
	}
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
		return nil, err
	}

	return NewCache(config)
}
