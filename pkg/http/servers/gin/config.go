package pkggin

import (
	"fmt"
)

type Config struct {
	routerPort string
	apiVersion string
}

func newConfig(routerPort, ApiVersion string) *Config {
	return &Config{
		routerPort: routerPort,
		apiVersion: ApiVersion,
	}
}

func (c *Config) GetRouterPort() string {
	return c.routerPort
}

func (c *Config) GetApiVersion() string {
	return c.routerPort
}

func (c *Config) Validate() error {
	if c.routerPort == "" {
		return fmt.Errorf("router port is not Configured")
	}
	return nil
}
