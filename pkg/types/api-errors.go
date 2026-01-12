package pkgtypes

import (
	"errors"
	"fmt"
	"net/http"
)

// APIErrorType define el tipo de error de la API.
type APIErrorType string

const (
	APIErrNotFound     APIErrorType = "NOT_FOUND"
	APIErrConflict     APIErrorType = "CONFLICT"
	APIErrBadRequest   APIErrorType = "BAD_REQUEST"
	APIErrInternal     APIErrorType = "INTERNAL_ERROR"
	APIErrValidation   APIErrorType = "VALIDATION_ERROR"
	APIErrUnauthorized APIErrorType = "UNAUTHORIZED"
	APIErrTimeout      APIErrorType = "TIMEOUT"
	APIErrUnavailable  APIErrorType = "SERVICE_UNAVAILABLE"
	APIErrForbidden    APIErrorType = "FORBIDDEN"
)

// APIError representa un error de la API con información adicional.
type APIError struct {
	Type    APIErrorType   `json:"type"`
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Details string         `json:"details,omitempty"`
	Context map[string]any `json:"context,omitempty"`
}

func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Type, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// APIErrorResponse es la respuesta estándar de error para la API.
type APIErrorResponse struct {
	Type    APIErrorType   `json:"type"`
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Details string         `json:"details,omitempty"`
	Context map[string]any `json:"context,omitempty"`
}

func (e *APIErrorResponse) IsType(t APIErrorType) bool { return e.Type == t }
func (e *APIErrorResponse) HasCode(code int) bool      { return e.Code == code }

// errorToAPIError mapea errores de dominio a tipos de error de la API.
var errorToAPIError = map[ErrorType]APIErrorType{
	ErrNotFound:        APIErrNotFound,
	ErrConflict:        APIErrConflict,
	ErrInvalidInput:    APIErrBadRequest,
	ErrValidation:      APIErrValidation,
	ErrOperationFailed: APIErrInternal,
	ErrConnection:      APIErrUnavailable,
	ErrTimeout:         APIErrTimeout,
	ErrAuthentication:  APIErrUnauthorized,
	ErrAuthorization:   APIErrForbidden,
	ErrInvalidID:       APIErrBadRequest,
	ErrUnavailable:     APIErrUnavailable,
	ErrTokenNotFound:   APIErrUnauthorized,
	ErrMissingField:    APIErrBadRequest,
	ErrBadRequest:      APIErrBadRequest,
}

var httpStatus = map[APIErrorType]int{
	APIErrBadRequest:   http.StatusBadRequest,
	APIErrNotFound:     http.StatusNotFound,
	APIErrConflict:     http.StatusConflict,
	APIErrInternal:     http.StatusInternalServerError,
	APIErrValidation:   http.StatusBadRequest,
	APIErrUnauthorized: http.StatusUnauthorized,
	APIErrTimeout:      http.StatusGatewayTimeout,
	APIErrUnavailable:  http.StatusServiceUnavailable,
	APIErrForbidden:    http.StatusForbidden,
}

// NewAPIError convierte errores de dominio a APIError con el código HTTP apropiado.
func NewAPIError(err error) (*APIError, int) {
	var domainErr *Error
	if errors.As(err, &domainErr) {
		apiType, exists := errorToAPIError[domainErr.Type]
		if !exists {
			apiType = APIErrInternal
		}
		code := httpStatus[apiType]
		apiError := &APIError{
			Type:    apiType,
			Code:    code,
			Message: domainErr.Message,
			Context: domainErr.Context,
		}
		if domainErr.Details != nil {
			apiError.Details = domainErr.Details.Error()
		}
		return apiError, code
	}
	return &APIError{
		Type:    APIErrInternal,
		Code:    http.StatusInternalServerError,
		Message: "Error interno del servidor",
		Details: err.Error(),
	}, http.StatusInternalServerError
}

// ToResponse convierte el APIError a la estructura JSON.
func (e *APIError) ToResponse() *APIErrorResponse {
	return &APIErrorResponse{
		Type:    e.Type,
		Code:    e.Code,
		Message: e.Message,
		Details: e.Details,
		Context: e.Context,
	}
}
