package pkgconsul

import (
	"fmt"
	"sync"

	"github.com/hashicorp/consul/api"
)

var (
	instance  Client
	once      sync.Once
	initError error
)

// client es la implementación del cliente de Consul
type client struct {
	client  *api.Client
	address string // Almacenar la dirección aquí
}

// InitializeConsulClient inicializa el cliente de Consul
func newClient(config config) (Client, error) {
	once.Do(func() {
		client := &client{}
		initError = client.connect(config)
		if initError != nil {
			instance = nil
		} else {
			instance = client
		}
	})
	return instance, initError
}

// connect conecta el cliente de Consul
func (c *client) connect(config config) error {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = config.Address

	client, err := api.NewClient(consulConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to Consul: %w", err)
	}

	registration := &api.AgentServiceRegistration{
		ID:      config.ID,
		Name:    config.Name,
		Port:    config.Port,
		Address: config.Service,
		Tags:    config.Tags,
		Check: &api.AgentServiceCheck{
			HTTP:     config.HealthCheck,
			Interval: config.CheckInterval,
			Timeout:  config.CheckTimeout,
		},
	}

	if err := client.Agent().ServiceRegister(registration); err != nil {
		return err
	}

	c.client = client
	c.address = consulConfig.Address // Almacenar la dirección

	return nil
}

// Client devuelve el cliente de Consul
func (c *client) Client() *api.Client {
	return c.client
}

// Address devuelve la dirección del cliente de Consul
func (c *client) Address() string {
	return c.address
}
