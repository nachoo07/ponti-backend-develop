package pkgtypes

import (
	"errors"
	"fmt"
)

// ErrorType define los tipos de errores del dominio.
type ErrorType string

// Constantes para ErrorType.
const (
	ErrNotFound        ErrorType = "NOT_FOUND"
	ErrConflict        ErrorType = "CONFLICT"
	ErrInvalidInput    ErrorType = "INVALID_INPUT"
	ErrValidation      ErrorType = "VALIDATION_ERROR"
	ErrOperationFailed ErrorType = "OPERATION_FAILED"
	ErrConnection      ErrorType = "CONNECTION_ERROR"
	ErrTimeout         ErrorType = "TIMEOUT"
	ErrAuthentication  ErrorType = "AUTHENTICATION_ERROR"
	ErrAuthorization   ErrorType = "AUTHORIZATION_ERROR"
	ErrInternal        ErrorType = "INTERNAL_ERROR"
	ErrInvalidID       ErrorType = "INVALID_ID"
	ErrUnavailable     ErrorType = "SERVICE_UNAVAILABLE"
	ErrTokenNotFound   ErrorType = "TOKEN_NOT_FOUND"
	// Nuevo error para campos faltantes
	ErrMissingField ErrorType = "MISSING_FIELD"
)

// Error representa un error del dominio.
type Error struct {
	Type    ErrorType      `json:"type"`
	Message string         `json:"message"`
	Details error          `json:"-"` // No se expone en JSON.
	Context map[string]any `json:"context,omitempty"`
}

// Error devuelve la representación en string del error.
func (e *Error) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("%s: %s (details: %v)", e.Type, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap permite extraer el error interno.
func (e *Error) Unwrap() error {
	return e.Details
}

// ToJSON convierte el error a un mapa, útil para serializar a JSON.
func (e *Error) ToJSON() map[string]any {
	response := map[string]any{
		"type":    e.Type,
		"message": e.Message,
	}
	if e.Context != nil {
		response["context"] = e.Context
	}
	return response
}

// --- Constructores de errores ---

// NewError crea un nuevo error de dominio.
func NewError(errType ErrorType, message string, details error) *Error {
	return &Error{
		Type:    errType,
		Message: message,
		Details: details,
	}
}

// NewErrorWithContext crea un nuevo error de dominio con contexto adicional.
func NewErrorWithContext(errType ErrorType, message string, details error, context map[string]any) *Error {
	return &Error{
		Type:    errType,
		Message: message,
		Details: details,
		Context: context,
	}
}

// NewInvalidIDError crea un nuevo error de tipo ErrInvalidID.
func NewInvalidIDError(message string, details error) *Error {
	return NewErrorWithContext(
		ErrInvalidID,
		message,
		details,
		map[string]any{
			"field": "id",
			"error": "invalid",
		},
	)
}

// NewAuthenticationError crea un error de autenticación.
func NewAuthenticationError(message string, details error) *Error {
	return NewError(ErrAuthentication, message, details)
}

// NewAuthorizationError crea un error de autorización.
func NewAuthorizationError(message string, details error) *Error {
	return NewError(ErrAuthorization, message, details)
}

// NewTimeoutError crea un error de tipo ErrTimeout.
func NewTimeoutError(message string, details error) *Error {
	return NewError(ErrTimeout, message, details)
}

// NewTokenNotFoundError crea un error cuando el token no se encuentra.
func NewTokenNotFoundError(details error) *Error {
	return NewError(
		ErrTokenNotFound,
		"Token not found in cache",
		details,
	)
}

// NewMissingFieldError crea un error para campos faltantes.
func NewMissingFieldError(field string) *Error {
	return NewErrorWithContext(
		ErrMissingField,
		fmt.Sprintf("The field '%s' is required", field),
		nil,
		map[string]any{"field": field},
	)
}

// --- Helpers para la verificación de errores ---

// IsNotFound verifica si el error es de tipo ErrNotFound.
func IsNotFound(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Type == ErrNotFound
}

// IsConflict verifica si el error es de tipo ErrConflict.
func IsConflict(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Type == ErrConflict
}

// IsValidationError verifica si el error es de tipo ErrValidation.
func IsValidationError(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Type == ErrValidation
}

// IsAuthenticationError verifica si el error es de tipo ErrAuthentication.
func IsAuthenticationError(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Type == ErrAuthentication
}

// IsAuthorizationError verifica si el error es de tipo ErrAuthorization.
func IsAuthorizationError(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Type == ErrAuthorization
}

// IsTokenNotFoundError verifica si el error es de tipo ErrTokenNotFound.
func IsTokenNotFoundError(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Type == ErrTokenNotFound
}

// GetErrorType extrae el tipo de error del dominio.
func GetErrorType(err error) (ErrorType, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e.Type, true
	}
	return "", false
}

// GetErrorContext obtiene el contexto del error del dominio.
func GetErrorContext(err error) (map[string]any, bool) {
	var e *Error
	if errors.As(err, &e) && e.Context != nil {
		return e.Context, true
	}
	return nil, false
}
