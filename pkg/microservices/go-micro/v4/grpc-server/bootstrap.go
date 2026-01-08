package pkggomicro

import (
	"github.com/spf13/viper"
)

func Bootstrap() (Server, error) {
	config := newConfig(
		viper.GetString("GRPC_SERVER_NAME"),
		viper.GetString("GRPC_SERVER_HOST"),
		viper.GetInt("GRPC_SERVER_PORT"),
	)

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return newServer(config)
}
