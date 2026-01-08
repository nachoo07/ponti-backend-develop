package pkgtypes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponseHelper proporciona funciones helper para manejar errores en handlers
type ErrorResponseHelper struct{}

// NewErrorResponseHelper crea una nueva instancia del helper
func NewErrorResponseHelper() *ErrorResponseHelper {
	return &ErrorResponseHelper{}
}

// BadRequest responde con un error 400 (Bad Request)
func (h *ErrorResponseHelper) BadRequest(c *gin.Context, message string, details error) {
	apiErr := &APIError{
		Type:    APIErrBadRequest,
		Code:    http.StatusBadRequest,
		Message: message,
	}
	if details != nil {
		apiErr.Details = details.Error()
	}
	c.JSON(http.StatusBadRequest, apiErr.ToResponse())
}

// NotFound responde con un error 404 (Not Found)
func (h *ErrorResponseHelper) NotFound(c *gin.Context, message string, details error) {
	apiErr := &APIError{
		Type:    APIErrNotFound,
		Code:    http.StatusNotFound,
		Message: message,
	}
	if details != nil {
		apiErr.Details = details.Error()
	}
	c.JSON(http.StatusNotFound, apiErr.ToResponse())
}

// Conflict responde con un error 409 (Conflict)
func (h *ErrorResponseHelper) Conflict(c *gin.Context, message string, details error) {
	apiErr := &APIError{
		Type:    APIErrConflict,
		Code:    http.StatusConflict,
		Message: message,
	}
	if details != nil {
		apiErr.Details = details.Error()
	}
	c.JSON(http.StatusConflict, apiErr.ToResponse())
}

// InternalError responde con un error 500 (Internal Server Error)
func (h *ErrorResponseHelper) InternalError(c *gin.Context, message string, details error) {
	apiErr := &APIError{
		Type:    APIErrInternal,
		Code:    http.StatusInternalServerError,
		Message: message,
	}
	if details != nil {
		apiErr.Details = details.Error()
	}
	c.JSON(http.StatusInternalServerError, apiErr.ToResponse())
}

// ValidationError responde con un error 400 (Validation Error)
func (h *ErrorResponseHelper) ValidationError(c *gin.Context, message string, details error) {
	apiErr := &APIError{
		Type:    APIErrValidation,
		Code:    http.StatusBadRequest,
		Message: message,
	}
	if details != nil {
		apiErr.Details = details.Error()
	}
	c.JSON(http.StatusBadRequest, apiErr.ToResponse())
}

// Unauthorized responde con un error 401 (Unauthorized)
func (h *ErrorResponseHelper) Unauthorized(c *gin.Context, message string, details error) {
	apiErr := &APIError{
		Type:    APIErrUnauthorized,
		Code:    http.StatusUnauthorized,
		Message: message,
	}
	if details != nil {
		apiErr.Details = details.Error()
	}
	c.JSON(http.StatusUnauthorized, apiErr.ToResponse())
}

// Forbidden responde con un error 403 (Forbidden)
func (h *ErrorResponseHelper) Forbidden(c *gin.Context, message string, details error) {
	apiErr := &APIError{
		Type:    APIErrForbidden,
		Code:    http.StatusForbidden,
		Message: message,
	}
	if details != nil {
		apiErr.Details = details.Error()
	}
	c.JSON(http.StatusForbidden, apiErr.ToResponse())
}

// Timeout responde con un error 504 (Gateway Timeout)
func (h *ErrorResponseHelper) Timeout(c *gin.Context, message string, details error) {
	apiErr := &APIError{
		Type:    APIErrTimeout,
		Code:    http.StatusGatewayTimeout,
		Message: message,
	}
	if details != nil {
		apiErr.Details = details.Error()
	}
	c.JSON(http.StatusGatewayTimeout, apiErr.ToResponse())
}

// ServiceUnavailable responde con un error 503 (Service Unavailable)
func (h *ErrorResponseHelper) ServiceUnavailable(c *gin.Context, message string, details error) {
	apiErr := &APIError{
		Type:    APIErrUnavailable,
		Code:    http.StatusServiceUnavailable,
		Message: message,
	}
	if details != nil {
		apiErr.Details = details.Error()
	}
	c.JSON(http.StatusServiceUnavailable, apiErr.ToResponse())
}

// HandleError maneja cualquier error convirtiéndolo a API error apropiado
func (h *ErrorResponseHelper) HandleError(c *gin.Context, err error) {
	apiErr, status := NewAPIError(err)
	c.JSON(status, apiErr.ToResponse())
}

// HandleDomainError maneja errores de dominio específicos
func (h *ErrorResponseHelper) HandleDomainError(c *gin.Context, err error) {
	apiErr, status := NewAPIError(err)
	c.JSON(status, apiErr.ToResponse())
}

// InvalidID responde con un error 400 para IDs inválidos
func (h *ErrorResponseHelper) InvalidID(c *gin.Context, fieldName string) {
	apiErr := &APIError{
		Type:    APIErrBadRequest,
		Code:    http.StatusBadRequest,
		Message: "Invalid " + fieldName,
		Details: fieldName + " is not a valid identifier",
		Context: map[string]any{"field": fieldName},
	}
	c.JSON(http.StatusBadRequest, apiErr.ToResponse())
}

// MissingField responde con un error 400 para campos faltantes
func (h *ErrorResponseHelper) MissingField(c *gin.Context, fieldName string) {
	apiErr := &APIError{
		Type:    APIErrBadRequest,
		Code:    http.StatusBadRequest,
		Message: "Missing required field",
		Details: "The field '" + fieldName + "' is required",
		Context: map[string]any{"field": fieldName},
	}
	c.JSON(http.StatusBadRequest, apiErr.ToResponse())
}

// InvalidPayload responde con un error 400 para payloads inválidos
func (h *ErrorResponseHelper) InvalidPayload(c *gin.Context, details error) {
	apiErr := &APIError{
		Type:    APIErrBadRequest,
		Code:    http.StatusBadRequest,
		Message: "Invalid request payload",
	}
	if details != nil {
		apiErr.Details = details.Error()
	}
	c.JSON(http.StatusBadRequest, apiErr.ToResponse())
}
