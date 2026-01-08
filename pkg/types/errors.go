// Package pkgtypes provides common types and utilities used across the ponti-backend project,
// including domain error types, API error handling, and common data structures.
package pkgtypes

import (
	"errors"
	"fmt"
)

// ErrorType defines the main domain error types.
type ErrorType string

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
	ErrMissingField    ErrorType = "MISSING_FIELD"
	ErrBadRequest      ErrorType = "BAD_REQUEST"
)

// Error is a domain-level error with type, message, details and optional context.
type Error struct {
	Type    ErrorType      `json:"type"`
	Message string         `json:"message"`
	Details error          `json:"-"` // Not marshaled to JSON
	Context map[string]any `json:"context,omitempty"`
}

func (e *Error) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("%s: %s (details: %v)", e.Type, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Details
}

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

// NewError creates a new domain error with the specified type, message, and details.
func NewError(errType ErrorType, message string, details error) *Error {
	return &Error{
		Type:    errType,
		Message: message,
		Details: details,
	}
}

func NewErrorWithContext(errType ErrorType, message string, details error, context map[string]any) *Error {
	return &Error{
		Type:    errType,
		Message: message,
		Details: details,
		Context: context,
	}
}

func NewInvalidIDError(message string, details error) *Error {
	return NewErrorWithContext(
		ErrInvalidID,
		message,
		details,
		map[string]any{"field": "id", "error": "invalid"},
	)
}

func NewAuthenticationError(message string, details error) *Error {
	return NewError(ErrAuthentication, message, details)
}

func NewAuthorizationError(message string, details error) *Error {
	return NewError(ErrAuthorization, message, details)
}

func NewTimeoutError(message string, details error) *Error {
	return NewError(ErrTimeout, message, details)
}

func NewTokenNotFoundError(details error) *Error {
	return NewError(ErrTokenNotFound, "Token not found in cache", details)
}

func NewMissingFieldError(field string) *Error {
	return NewErrorWithContext(
		ErrMissingField,
		fmt.Sprintf("The field '%s' is required", field),
		nil,
		map[string]any{"field": field},
	)
}

// IsNotFound checks if the error is a "not found" domain error.
func IsNotFound(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Type == ErrNotFound
}

func IsConflict(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Type == ErrConflict
}

func IsValidationError(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Type == ErrValidation
}

func IsAuthenticationError(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Type == ErrAuthentication
}

func IsAuthorizationError(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Type == ErrAuthorization
}

func IsTokenNotFoundError(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Type == ErrTokenNotFound
}

// GetErrorType extracts the error type from a domain error, returning the type and whether it was found.
func GetErrorType(err error) (ErrorType, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e.Type, true
	}
	return "", false
}

func GetErrorContext(err error) (map[string]any, bool) {
	var e *Error
	if errors.As(err, &e) && e.Context != nil {
		return e.Context, true
	}
	return nil, false
}

func IsErrInvalidInput(err error) bool {
	var e *Error
	return errors.As(err, &e) &&
		(e.Type == ErrInvalidInput || e.Type == ErrInvalidID || e.Type == ErrBadRequest)
}
