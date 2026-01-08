package pkggomicro

import (
	"fmt"

	"github.com/google/uuid"
)

type config struct {
	Router        any
	ServerName    string
	consulAddress string
	ServerHost    string
	ServerPort    int
	ServerID      string
}

func newConfig(
	Router any,
	ServerName string,
	consulAddress string,
	ServerHost string,
	ServerPort int,
) Config {
	return &config{
		Router:        Router,
		ServerName:    ServerName,
		consulAddress: consulAddress,
		ServerHost:    ServerHost,
		ServerPort:    ServerPort,
		ServerID:      uuid.New().String(),
	}
}

func (c *config) GetRouter() any {
	return c.Router
}

func (c *config) GetServerName() string {
	return c.ServerName
}

func (c *config) GetServerHost() string {
	return c.ServerHost
}

func (c *config) GetServerPort() int {
	return c.ServerPort
}

func (c *config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.ServerHost, c.ServerPort)
}

func (c *config) GetConsulAddress() string {
	return c.consulAddress
}

func (c *config) GetServerID() string {
	return c.ServerID
}

func (c *config) Validate() error {
	if c.ServerName == "" {
		return fmt.Errorf("missing server name")
	}
	if c.ServerHost == "" {
		return fmt.Errorf("missing server host")
	}
	if c.ServerPort == 0 {
		return fmt.Errorf("missing server port")
	}
	if c.consulAddress == "" {
		return fmt.Errorf("missing consul address")
	}
	return nil
}
