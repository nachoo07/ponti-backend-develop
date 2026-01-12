package pkgrealstack

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/alphacodinggroup/ponti-backend/pkg/cloud/aws/ports"
)

// stack implementa la interfaz Stack para AWS
type stack struct {
	config    ports.Config
	awsConfig aws.Config
	mu        sync.RWMutex
	connected bool
}

// NewStack crea una nueva instancia del stack AWS
func NewStack(cfg ports.Config) (ports.Stack, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if cfg.GetProvider() != ports.ProviderAWS {
		return nil, fmt.Errorf("invalid provider for AWS stack: %s", cfg.GetProvider())
	}

	s := &stack{
		config: cfg,
	}

	if err := s.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect AWS stack: %w", err)
	}

	return s, nil
}

// Connect establece la conexión con AWS
func (s *stack) Connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.connected {
		return nil // Ya conectado
	}

	// Contexto con timeout para la carga de configuración
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// Opciones base de configuración
	var opts []func(*config.LoadOptions) error
	opts = append(opts, []func(*config.LoadOptions) error{
		config.WithRegion(s.config.GetAwsRegion()),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			s.config.GetAwsAccessKeyID(),
			s.config.GetAwsSecretAccessKey(),
			"",
		)),
	}...)

	// Agregar opciones adicionales basadas en la configuración
	if len(s.config.GetServices()) > 0 {
		opts = append(opts, s.getServiceOptions()...)
	}

	// Cargar la configuración de AWS
	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
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

// NewSQSClient crea un nuevo cliente SQS
func (s *stack) NewSQSClient() ports.SQSClient {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.connected {
		if err := s.Connect(); err != nil {
			return nil
		}
	}

	return NewSQSClient(s.awsConfig)
}

// NewLambdaClient crea un nuevo cliente Lambda
func (s *stack) NewLambdaClient() ports.LambdaClient {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.connected {
		if err := s.Connect(); err != nil {
			return nil
		}
	}

	return NewLambdaClient(s.awsConfig)
}

// getServiceOptions retorna las opciones de configuración específicas para los servicios
func (s *stack) getServiceOptions() []func(*config.LoadOptions) error {
	var opts []func(*config.LoadOptions) error

	// Aquí puedes agregar opciones específicas por servicio
	for _, service := range s.config.GetServices() {
		switch service {
		case ports.ServiceSQS:
			opts = append(opts, getSQSOptions()...)
		case ports.ServiceLambda:
			opts = append(opts, getLambdaOptions()...)
			// Agregar más servicios según sea necesario
		}
	}

	return opts
}

func getSQSOptions() []func(*config.LoadOptions) error {
	return []func(*config.LoadOptions) error{
		// Opciones específicas de SQS
	}
}

func getLambdaOptions() []func(*config.LoadOptions) error {
	return []func(*config.LoadOptions) error{
		// Opciones específicas de Lambda
	}
}
