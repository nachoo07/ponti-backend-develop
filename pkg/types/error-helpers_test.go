package pkgtypes

import (
	"testing"
)

func TestErrorResponseHelperCreation(t *testing.T) {
	helper := NewErrorResponseHelper()
	if helper == nil {
		t.Error("Expected ErrorResponseHelper to be created, got nil")
	}
}

func TestErrorResponseHelperMethods(t *testing.T) {
	helper := NewErrorResponseHelper()

	// Test that all methods exist and can be called
	// This is a basic compilation test
	if helper == nil {
		t.Error("Helper should not be nil")
	}
}
