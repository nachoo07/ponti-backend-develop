package pkggorhttp

import (
	"fmt"
	"os"
)

type config struct {
	port       string
	apiVersion string
}

func newConfig(port, apiVersion string) Config {
	if port == "" {
		port = os.Getenv("WS_SERVER_PORT")
	}
	if apiVersion == "" {
		apiVersion = os.Getenv("API_VERSION")
	}
	return &config{
		port:       port,
		apiVersion: apiVersion,
	}
}

func (c *config) GetPort() string {
	return c.port
}

func (c *config) GetAPIVersion() string {
	return c.apiVersion
}

func (c *config) Validate() error {
	if c.port == "" {
		return fmt.Errorf("el puerto del servidor no está configurado")
	}
	return nil
}
