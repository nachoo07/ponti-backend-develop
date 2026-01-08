package pkggomicro

import (
	"github.com/spf13/viper"

	pkgclient "github.com/alphacodinggroup/ponti-backend/pkg/microservices/go-micro/v4/grpc-client"
	pkgserver "github.com/alphacodinggroup/ponti-backend/pkg/microservices/go-micro/v4/grpc-server"
	pkgbroker "github.com/alphacodinggroup/ponti-backend/pkg/microservices/go-micro/v4/rabbitmq-broker"
)

func Bootstrap(server pkgserver.Server, client pkgclient.Client, broker pkgbroker.Broker) (Service, error) {
	config := newConfig(
		server.GetServer(),
		client.GetClient(),
		broker.GetBroker(),
		viper.GetString("CONSUL_ADDRESS"),
	)

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return newService(config)
}
