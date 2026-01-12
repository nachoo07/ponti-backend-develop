package pkgtypes

import (
	"errors"
	"fmt"
	"net/http"
)

// APIErrorType define los tipos de errores de API.
type APIErrorType string

// Constantes para APIErrorType.
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

// APIError representa un error de API.
type APIError struct {
	Type    APIErrorType   `json:"type"`
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Details string         `json:"details,omitempty"`
	Context map[string]any `json:"context,omitempty"`
}

// Error devuelve la representación en string del APIError.
func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Type, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// APIErrorResponse representa la estructura de respuesta de error para JSON.
type APIErrorResponse struct {
	Type    APIErrorType   `json:"type"`
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Details string         `json:"details,omitempty"`
	Context map[string]any `json:"context,omitempty"`
}

// IsType comprueba si el error es de un determinado tipo.
func (e *APIErrorResponse) IsType(t APIErrorType) bool {
	return e.Type == t
}

// HasCode comprueba si el error tiene el código HTTP especificado.
func (e *APIErrorResponse) HasCode(code int) bool {
	return e.Code == code
}

// Mapeo entre errores de dominio y errores de API.
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
}

// Mapear APIErrorType a códigos HTTP.
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

// NewAPIError convierte un error de dominio a un APIError junto con el código HTTP.
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

	// Para errores no manejados, se considera error interno.
	return &APIError{
		Type:    APIErrInternal,
		Code:    http.StatusInternalServerError,
		Message: "Internal server error",
		Details: err.Error(),
	}, http.StatusInternalServerError
}

// ToResponse convierte un APIError a un APIErrorResponse.
func (e *APIError) ToResponse() *APIErrorResponse {
	return &APIErrorResponse{
		Type:    e.Type,
		Code:    e.Code,
		Message: e.Message,
		Details: e.Details,
		Context: e.Context,
	}
}
