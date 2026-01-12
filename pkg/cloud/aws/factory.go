package pkgaws

import (
	"fmt"

	localstack "github.com/alphacodinggroup/ponti-backend/pkg/cloud/aws/localstack"
	ports "github.com/alphacodinggroup/ponti-backend/pkg/cloud/aws/ports"
	realstack "github.com/alphacodinggroup/ponti-backend/pkg/cloud/aws/realstack"
)

// awsProvider implementa StackFactory para AWS real
type awsProvider struct{}

// localstackProvider implementa StackFactory para Localstack
type localstackProvider struct{}

// StackFactory define la interfaz para la creación de stacks
type StackFactory interface {
	// CreateStack crea un nuevo stack basado en la configuración proporcionada
	CreateStack(config ports.Config) (ports.Stack, error)
}

// NewStackFactory crea un nuevo factory basado en el provider especificado
func NewStackFactory(provider string) (StackFactory, error) {
	// Validar que el provider sea válido
	if provider == "" {
		return nil, &ConfigError{
			Field:   "provider",
			Message: "provider cannot be empty",
		}
	}

	switch provider {
	case ports.ProviderAWS:
		return &awsProvider{}, nil
	case ports.ProviderLocalstack:
		return &localstackProvider{}, nil
	default:
		return nil, &ConfigError{
			Field:   "provider",
			Message: fmt.Sprintf("unsupported provider: %s", provider),
		}
	}
}

// CreateStack implementación para AWS real
func (p *awsProvider) CreateStack(config ports.Config) (ports.Stack, error) {
	// Validar que la configuración coincida con el provider
	if config.GetProvider() != ports.ProviderAWS {
		return nil, &ConfigError{
			Field:   "provider",
			Message: "config provider does not match AWS provider",
		}
	}

	stack, err := realstack.NewStack(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS stack: %w", err)
	}

	return stack, nil
}

// CreateStack implementación para Localstack
func (p *localstackProvider) CreateStack(config ports.Config) (ports.Stack, error) {
	// Validar que la configuración coincida con el provider
	if config.GetProvider() != ports.ProviderLocalstack {
		return nil, &ConfigError{
			Field:   "provider",
			Message: "config provider does not match Localstack provider",
		}
	}

	// Validar endpoint para Localstack
	if config.GetEndpoint() == "" {
		return nil, &ConfigError{
			Field:   "endpoint",
			Message: "endpoint is required for Localstack",
		}
	}

	stack, err := localstack.NewStack(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Localstack stack: %w", err)
	}

	return stack, nil
}
