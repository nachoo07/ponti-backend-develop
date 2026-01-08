package pkggomicro

import (
	"fmt"

	"github.com/google/uuid"
)

type config struct {
	serverName string
	serverHost string
	serverPort int
	serverID   string
}

func newConfig(serverName string, serverHost string, serverPort int) Config {
	return &config{
		serverName: serverName,
		serverHost: serverHost,
		serverPort: serverPort,
		serverID:   uuid.New().String(),
	}
}

func (c *config) GetServerName() string {
	return c.serverName
}

func (c *config) GetServerHost() string {
	return c.serverHost
}

func (c *config) GetServerPort() int {
	return c.serverPort
}

func (c *config) GetServerID() string {
	return c.serverID
}

func (c *config) Validate() error {
	if c.serverName == "" {
		return fmt.Errorf("missing service name")
	}
	if c.serverHost == "" {
		return fmt.Errorf("missing server host")
	}
	if c.serverPort == 0 {
		return fmt.Errorf("missing server port")
	}
	return nil
}
