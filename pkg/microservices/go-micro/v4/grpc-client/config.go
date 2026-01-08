package pkggomicro

import (
	"fmt"
)

type config struct {
	consulAddress string
	serverName    string
}

func newConfig(ca, sn string) Config {
	return &config{
		consulAddress: ca,
		serverName:    sn,
	}
}

func (c *config) GetConsulAddress() string {
	return c.consulAddress
}

func (c *config) GetServerName() string {
	return c.serverName
}

func (c *config) Validate() error {
	if c.consulAddress == "" {
		return fmt.Errorf("missing consul address")
	}
	if c.serverName == "" {
		return fmt.Errorf("missing service name")
	}
	return nil
}
