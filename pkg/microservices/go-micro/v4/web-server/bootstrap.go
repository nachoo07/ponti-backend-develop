package pkggomicro

import (
	"github.com/spf13/viper"
)

func Bootstrap(webRouter any) (Server, error) {
	config := newConfig(
		webRouter,
		viper.GetString("HTTP_SERVER_NAME"),
		viper.GetString("CONSUL_ADDRESS"),
		viper.GetString("HTTP_SERVER_HOST"),
		viper.GetInt("HTTP_SERVER_PORT"),
	)

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return newServer(config)
}
