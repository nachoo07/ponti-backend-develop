package pkglocalstack

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/alphacodinggroup/ponti-backend/pkg/cloud/aws/ports"
)

// stack implementa la interfaz Stack para Localstack
type stack struct {
	config      ports.Config
	awsConfig   aws.Config
	mu          sync.RWMutex
	connected   bool
	initialized time.Time
}

// NewStack crea una nueva instancia del stack de Localstack
func NewStack(cfg ports.Config) (ports.Stack, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if cfg.GetProvider() != ports.ProviderLocalstack {
		return nil, fmt.Errorf("invalid provider for Localstack: %s", cfg.GetProvider())
	}

	if err := validateLocalstackEndpoint(cfg.GetEndpoint()); err != nil {
		return nil, fmt.Errorf("invalid Localstack endpoint: %w", err)
	}

	s := &stack{
		config:      cfg,
		initialized: time.Now(),
	}

	if err := s.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to Localstack: %w", err)
	}

	return s, nil
}

// Connect establece la conexión con Localstack
func (s *stack) Connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.connected {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// Configurar las opciones básicas de AWS
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(s.config.GetAwsRegion()),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			s.config.GetAwsAccessKeyID(),
			s.config.GetAwsSecretAccessKey(),
			"",
		)),
		config.WithRetryMode(aws.RetryModeStandard),
		config.WithClientLogMode(aws.LogRetries | aws.LogRequest),
	}

	// Cargar la configuración
	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to load Localstack config: %w", err)
	}

	s.awsConfig = awsCfg
	s.connected = true
	return nil
}

// GetConfig retorna la configuración de AWS actual
func (s *stack) GetConfig() aws.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.awsConfig
}

// NewSQSClient crea un nuevo cliente SQS para Localstack
func (s *stack) NewSQSClient() ports.SQSClient {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.connected {
		if err := s.Connect(); err != nil {
			return nil
		}
	}

	return NewSQSClient(s.awsConfig, s.config.GetEndpoint())
}

// NewLambdaClient crea un nuevo cliente Lambda para Localstack
func (s *stack) NewLambdaClient() ports.LambdaClient {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.connected {
		if err := s.Connect(); err != nil {
			return nil
		}
	}

	return NewLambdaClient(s.awsConfig, s.config.GetEndpoint())
}

// validateLocalstackEndpoint valida el endpoint de Localstack
func validateLocalstackEndpoint(endpoint string) error {
	if endpoint == "" {
		return fmt.Errorf("empty endpoint")
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint URL: %w", err)
	}

	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("invalid endpoint format: scheme and host are required")
	}

	return nil
}
