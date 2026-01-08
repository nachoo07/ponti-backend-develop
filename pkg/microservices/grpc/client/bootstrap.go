package pkgcgrpcclient

import (
	"github.com/spf13/viper"
)

func Bootstrap(grpcServerHostKey, grpcServerPortKey string) (Client, error) {
	config := newConfig(
		viper.GetString(grpcServerHostKey),
		viper.GetInt(grpcServerPortKey),
		nil, // Configuración TLS, si es necesario
	)

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return newClient(config)
}
