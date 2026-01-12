package pkgaws

import (
	"fmt"
	"strings"
)

// ErrorCode define los códigos de error específicos
type ErrorCode string

const (
	// Códigos de error para Provider
	ErrProviderNotFound   ErrorCode = "PROVIDER_NOT_FOUND"
	ErrProviderInvalid    ErrorCode = "PROVIDER_INVALID"
	ErrProviderConnection ErrorCode = "PROVIDER_CONNECTION"
	ErrProviderTimeout    ErrorCode = "PROVIDER_TIMEOUT"

	// Códigos de error para Config
	ErrConfigInvalid    ErrorCode = "CONFIG_INVALID"
	ErrConfigMissing    ErrorCode = "CONFIG_MISSING"
	ErrConfigValidation ErrorCode = "CONFIG_VALIDATION"

	// Códigos de error para Servicios
	ErrServiceNotFound    ErrorCode = "SERVICE_NOT_FOUND"
	ErrServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrServiceTimeout     ErrorCode = "SERVICE_TIMEOUT"

	// Códigos de error para operaciones
	ErrOperationFailed  ErrorCode = "OPERATION_FAILED"
	ErrOperationTimeout ErrorCode = "OPERATION_TIMEOUT"
	ErrOperationInvalid ErrorCode = "OPERATION_INVALID"
)

// ProviderError representa errores relacionados con los providers AWS
type ProviderError struct {
	Provider string
	Code     ErrorCode
	Message  string
	Cause    error
	Details  map[string]any
}

func (e *ProviderError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] Provider %s error: %s", e.Code, e.Provider, e.Message))

	if e.Cause != nil {
		sb.WriteString(fmt.Sprintf(" - caused by: %v", e.Cause))
	}

	if len(e.Details) > 0 {
		sb.WriteString(" - details:")
		for k, v := range e.Details {
			sb.WriteString(fmt.Sprintf(" %s=%v", k, v))
		}
	}

	return sb.String()
}

func (e *ProviderError) Unwrap() error {
	return e.Cause
}

// NewProviderError crea un nuevo error de provider
func NewProviderError(provider string, code ErrorCode, message string, cause error) *ProviderError {
	return &ProviderError{
		Provider: provider,
		Code:     code,
		Message:  message,
		Cause:    cause,
		Details:  make(map[string]any),
	}
}

// WithDetail agrega un detalle al error
func (e *ProviderError) WithDetail(key string, value any) *ProviderError {
	e.Details[key] = value
	return e
}

// ConfigError representa errores de configuración
type ConfigError struct {
	Field   string
	Code    ErrorCode
	Message string
	Cause   error
	Details map[string]any
}

func (e *ConfigError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] Config error in field '%s': %s", e.Code, e.Field, e.Message))

	if e.Cause != nil {
		sb.WriteString(fmt.Sprintf(" - caused by: %v", e.Cause))
	}

	if len(e.Details) > 0 {
		sb.WriteString(" - details:")
		for k, v := range e.Details {
			sb.WriteString(fmt.Sprintf(" %s=%v", k, v))
		}
	}

	return sb.String()
}

func (e *ConfigError) Unwrap() error {
	return e.Cause
}

// NewConfigError crea un nuevo error de configuración
func NewConfigError(field string, code ErrorCode, message string, cause error) *ConfigError {
	return &ConfigError{
		Field:   field,
		Code:    code,
		Message: message,
		Cause:   cause,
		Details: make(map[string]any),
	}
}

// WithDetail agrega un detalle al error
func (e *ConfigError) WithDetail(key string, value any) *ConfigError {
	e.Details[key] = value
	return e
}

// ServiceError representa errores de servicios específicos (SQS, Lambda, etc.)
type ServiceError struct {
	Service string
	Code    ErrorCode
	Message string
	Cause   error
	Details map[string]any
}

func (e *ServiceError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] Service %s error: %s", e.Code, e.Service, e.Message))

	if e.Cause != nil {
		sb.WriteString(fmt.Sprintf(" - caused by: %v", e.Cause))
	}

	if len(e.Details) > 0 {
		sb.WriteString(" - details:")
		for k, v := range e.Details {
			sb.WriteString(fmt.Sprintf(" %s=%v", k, v))
		}
	}

	return sb.String()
}

func (e *ServiceError) Unwrap() error {
	return e.Cause
}

// NewServiceError crea un nuevo error de servicio
func NewServiceError(service string, code ErrorCode, message string, cause error) *ServiceError {
	return &ServiceError{
		Service: service,
		Code:    code,
		Message: message,
		Cause:   cause,
		Details: make(map[string]any),
	}
}

// WithDetail agrega un detalle al error
func (e *ServiceError) WithDetail(key string, value any) *ServiceError {
	e.Details[key] = value
	return e
}

// IsProviderError verifica si un error es de tipo ProviderError
func IsProviderError(err error) bool {
	_, ok := err.(*ProviderError)
	return ok
}

// IsConfigError verifica si un error es de tipo ConfigError
func IsConfigError(err error) bool {
	_, ok := err.(*ConfigError)
	return ok
}

// IsServiceError verifica si un error es de tipo ServiceError
func IsServiceError(err error) bool {
	_, ok := err.(*ServiceError)
	return ok
}

// Ejemplo de uso:
/*
err := NewProviderError("aws", ErrProviderConnection, "failed to connect", originalError).
	WithDetail("attempt", 3).
	WithDetail("endpoint", "http://localhost:4566")

if IsProviderError(err) {
	// Manejar error específico del provider
}
*/
