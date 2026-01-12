package pkgaws

import (
	"fmt"

	"github.com/alphacodinggroup/ponti-backend/pkg/cloud/aws/ports"
)

// config estructura principal de configuración
type config struct {
	provider        string
	awsAccessKeyID  string
	awsSecretAccess string
	awsRegion       string
	endpoint        string
	edgePort        int
	webUIPort       int
	services        []string
	dataDir         string
}

// ConfigOption define un modificador de configuración
type ConfigOption func(*config)

// newConfig crea una nueva configuración con opciones
func newConfig(provider string, opts ...ConfigOption) ports.Config {
	cfg := &config{
		provider:  provider,
		services:  make([]string, 0),
		edgePort:  4566, // Puerto por defecto de Localstack
		webUIPort: 4571, // Puerto por defecto del UI de Localstack
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// Opciones de configuración
func WithCredentials(accessKey, secretKey string) ConfigOption {
	return func(c *config) {
		c.awsAccessKeyID = accessKey
		c.awsSecretAccess = secretKey
	}
}

func WithRegion(region string) ConfigOption {
	return func(c *config) {
		c.awsRegion = region
	}
}

func WithLocalstackConfig(endpoint string, edgePort, webUIPort int) ConfigOption {
	return func(c *config) {
		c.endpoint = endpoint
		if edgePort > 0 {
			c.edgePort = edgePort
		}
		if webUIPort > 0 {
			c.webUIPort = webUIPort
		}
	}
}

func WithServices(services []string) ConfigOption {
	return func(c *config) {
		c.services = services
	}
}

func WithDataDir(dataDir string) ConfigOption {
	return func(c *config) {
		c.dataDir = dataDir
	}
}

// Implementación de la interfaz config
func (c *config) GetProvider() string {
	return c.provider
}

func (c *config) GetAwsAccessKeyID() string {
	return c.awsAccessKeyID
}

func (c *config) GetAwsSecretAccessKey() string {
	return c.awsSecretAccess
}

func (c *config) GetAwsRegion() string {
	return c.awsRegion
}

func (c *config) GetEndpoint() string {
	return c.endpoint
}

func (c *config) SetEndpoint(endpoint string) {
	c.endpoint = endpoint
}

func (c *config) GetServices() []string {
	return c.services
}

func (c *config) SetServices(services []string) {
	c.services = services
}

func (c *config) GetEdgePort() int {
	return c.edgePort
}

func (c *config) GetWebUIPort() int {
	return c.webUIPort
}

func (c *config) GetDataDir() string {
	return c.dataDir
}

// Validate verifica que la configuración sea válida
func (c *config) Validate() error {
	// Validaciones básicas
	if c.awsAccessKeyID == "" {
		return fmt.Errorf("AWS_ACCESS_KEY_ID is required")
	}
	if c.awsSecretAccess == "" {
		return fmt.Errorf("AWS_SECRET_ACCESS_KEY is required")
	}
	if c.awsRegion == "" {
		return fmt.Errorf("AWS_REGION is required")
	}

	// Validaciones específicas de Localstack
	if c.provider == ports.ProviderLocalstack {
		if c.endpoint == "" {
			return fmt.Errorf("endpoint is required for localstack")
		}
		if c.edgePort <= 0 {
			return fmt.Errorf("invalid edge port for localstack")
		}
		if c.webUIPort <= 0 {
			return fmt.Errorf("invalid web UI port for localstack")
		}
	}

	// Validación de servicios si están especificados
	if len(c.services) > 0 {
		for _, service := range c.services {
			if !ports.ValidServices[service] {
				return fmt.Errorf("invalid service: %s", service)
			}
		}
	}

	return nil
}
