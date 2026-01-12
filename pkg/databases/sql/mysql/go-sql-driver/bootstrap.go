package pkgmysql

import (
	"github.com/spf13/viper"
)

// Bootstrap inicializa la configuración y crea una instancia de repositorio MySQL.
func Bootstrap() (Repository, error) {
	config := config{
		User:     viper.GetString("MYSQL_USER"),
		Password: viper.GetString("MYSQL_PASSWORD"),
		Host:     viper.GetString("MYSQL_HOST"),
		Port:     viper.GetString("MYSQL_PORT"),
		Database: viper.GetString("MYSQL_DATABASE"),
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return newRepository(config)
}
