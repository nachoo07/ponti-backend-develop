package pkgtypes

import (
	"errors"
	"testing"
)

func TestErrorCreation(t *testing.T) {
	// Test basic error creation
	err := NewError(ErrNotFound, "resource not found", nil)
	if err.Type != ErrNotFound {
		t.Errorf("Expected error type %s, got %s", ErrNotFound, err.Type)
	}
	if err.Message != "resource not found" {
		t.Errorf("Expected message 'resource not found', got '%s'", err.Message)
	}
	if err.Details != nil {
		t.Errorf("Expected nil details, got %v", err.Details)
	}
}

func TestErrorWithDetails(t *testing.T) {
	// Test error with details
	details := errors.New("connection timeout")
	err := NewError(ErrInternal, "database connection failed", details)
	if err.Details != details {
		t.Errorf("Expected details %v, got %v", details, err.Details)
	}
}

func TestErrorString(t *testing.T) {
	// Test error string representation
	err := NewError(ErrNotFound, "resource not found", nil)
	expected := "NOT_FOUND: resource not found"
	if err.Error() != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
	}
}

func TestErrorWithContext(t *testing.T) {
	// Test error with context
	context := map[string]any{"field": "email", "value": "invalid"}
	err := NewErrorWithContext(ErrValidation, "validation failed", nil, context)
	if err.Context["field"] != "email" {
		t.Errorf("Expected context field 'email', got %v", err.Context["field"])
	}
}

func TestSpecializedErrors(t *testing.T) {
	// Test specialized error constructors
	details := errors.New("not a number")
	err := NewInvalidIDError("invalid id", details)
	if err.Type != ErrInvalidID {
		t.Errorf("Expected error type %s, got %s", ErrInvalidID, err.Type)
	}
	if err.Context["field"] != "id" {
		t.Errorf("Expected context field 'id', got %v", err.Context["field"])
	}
}
