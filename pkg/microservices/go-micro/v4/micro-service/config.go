package pkggomicro

import (
	"fmt"

	"go-micro.dev/v4/broker"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/server"
)

type config struct {
	server        server.Server
	client        client.Client
	broker        broker.Broker
	consulAddress string
}

func newConfig(
	server server.Server,
	client client.Client,
	broker broker.Broker,
	consulAddress string,
) Config {
	return &config{
		server:        server,
		client:        client,
		broker:        broker,
		consulAddress: consulAddress,
	}
}

func (c *config) GetServer() server.Server {
	return c.server
}

func (c *config) GetClient() client.Client {
	return c.client
}

func (c *config) GetBroker() client.Client {
	return c.client
}

func (c *config) GetConsulAddress() string {
	return c.consulAddress
}

func (c *config) Validate() error {
	if c.server == nil {
		return fmt.Errorf("missing server")
	}
	if c.client == nil {
		return fmt.Errorf("missing client")
	}
	return nil
}
