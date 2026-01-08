package pkgsqlite

import (
	"github.com/spf13/viper"
)

func Bootstrap() (Repository, error) {
	config := newConfig(
		viper.GetString("SQLITE_DB_PATH"),
		viper.GetBool("SQLITE_IN_MEMORY"),
	)

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return newRepository(config)
}
