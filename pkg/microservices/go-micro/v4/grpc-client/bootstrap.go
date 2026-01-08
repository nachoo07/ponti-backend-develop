package pkggomicro

import (
	"github.com/spf13/viper"
)

func Bootstrap() (Client, error) {
	config := newConfig(
		viper.GetString("CONSUL_ADDRESS"),
		viper.GetString("GRPC_SERVER_NAME"),
	)

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return newClient(config)
}
