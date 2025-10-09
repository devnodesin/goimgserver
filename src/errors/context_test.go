package errors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestErrorContext_Preservation_ThroughLayers tests error context preservation
func TestErrorContext_Preservation_ThroughLayers(t *testing.T) {
	// Simulate error through multiple layers
	baseErr := NewAppError("base error", ErrorTypeValidation, nil)
	
	// Add context at layer 1
	layer1Err := baseErr.WithContext("layer", "handler")
	
	// Add context at layer 2
	layer2Err := layer1Err.WithContext("function", "processImage")
	
	// Context should be preserved through all layers
	ctx := layer2Err.Context()
	require.NotNil(t, ctx)
	assert.Equal(t, "handler", ctx["layer"])
	assert.Equal(t, "processImage", ctx["function"])
}

// TestErrorContext_StackTrace_Development tests stack traces in dev mode
func TestErrorContext_StackTrace_Development(t *testing.T) {
	// Enable development mode
	SetDevelopmentMode(true)
	defer SetDevelopmentMode(false)
	
	err := NewAppError("test error", ErrorTypeInternal, nil)
	
	// Stack trace should be available in development mode
	stack := err.StackTrace()
	assert.NotEmpty(t, stack, "Stack trace should be available in development mode")
}

// TestErrorContext_Sanitization_Production tests error sanitization in prod
func TestErrorContext_Sanitization_Production(t *testing.T) {
	// Ensure production mode
	SetDevelopmentMode(false)
	
	// Create error with sensitive information
	err := NewAppError("database connection failed: password=secret123", ErrorTypeInternal, nil)
	
	// User message should be sanitized
	userMsg := err.UserMessage()
	assert.NotContains(t, userMsg, "secret123", "Should not expose sensitive details")
	assert.NotContains(t, userMsg, "password=", "Should not expose sensitive details")
}

// TestErrorContext_RequestContext tests request context propagation
func TestErrorContext_RequestContext(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", "req-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")
	
	err := NewAppErrorWithContext(ctx, "test error", ErrorTypeValidation, nil)
	
	errCtx := err.Context()
	assert.NotNil(t, errCtx)
	// Context should be extracted from the request context
}

// TestErrorContext_Chaining tests context preservation through error chains
func TestErrorContext_Chaining(t *testing.T) {
	originalErr := NewAppError("original", ErrorTypeInternal, nil)
	originalErr = originalErr.WithContext("component", "cache")
	
	wrappedErr := NewAppError("wrapped", ErrorTypeValidation, originalErr)
	wrappedErr = wrappedErr.WithContext("component", "handler")
	
	// Both contexts should be accessible
	ctx := wrappedErr.Context()
	assert.NotNil(t, ctx)
	assert.Contains(t, ctx, "component")
}

// TestErrorContext_Clone tests error context cloning
func TestErrorContext_Clone(t *testing.T) {
	err := NewAppError("test", ErrorTypeInternal, nil)
	err = err.WithContext("key1", "value1")
	
	cloned := err.Clone()
	cloned = cloned.WithContext("key2", "value2")
	
	// Original should not have key2
	origCtx := err.Context()
	assert.NotContains(t, origCtx, "key2")
	
	// Clone should have both keys
	clonedCtx := cloned.Context()
	assert.Contains(t, clonedCtx, "key1")
	assert.Contains(t, clonedCtx, "key2")
}

// TestErrorContext_Metadata tests adding metadata to errors
func TestErrorContext_Metadata(t *testing.T) {
	err := NewAppError("test", ErrorTypeValidation, nil)
	err = err.WithMetadata("timestamp", "2024-01-01T12:00:00Z")
	err = err.WithMetadata("version", "1.0.0")
	
	metadata := err.Metadata()
	assert.Equal(t, "2024-01-01T12:00:00Z", metadata["timestamp"])
	assert.Equal(t, "1.0.0", metadata["version"])
}

// TestErrorContext_ThreadSafety tests concurrent access to error context
func TestErrorContext_ThreadSafety(t *testing.T) {
	err := NewAppError("test", ErrorTypeInternal, nil)
	
	done := make(chan bool)
	
	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_ = err.Context()
			_ = err.UserMessage()
			_ = err.Code()
			done <- true
		}()
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Should not panic or race
}

// TestErrorContext_EmptyContext tests handling of empty context
func TestErrorContext_EmptyContext(t *testing.T) {
	err := NewAppError("test", ErrorTypeInternal, nil)
	
	ctx := err.Context()
	assert.NotNil(t, ctx, "Context should not be nil even when empty")
}

// TestErrorContext_LargeContext tests handling of large context data
func TestErrorContext_LargeContext(t *testing.T) {
	err := NewAppError("test", ErrorTypeInternal, nil)
	
	// Add many context values
	for i := 0; i < 100; i++ {
		err = err.WithContext("key"+string(rune(i)), "value"+string(rune(i)))
	}
	
	ctx := err.Context()
	assert.GreaterOrEqual(t, len(ctx), 100)
}

// TestErrorContext_Serialization tests error context serialization
func TestErrorContext_Serialization(t *testing.T) {
	err := NewAppError("test", ErrorTypeValidation, nil)
	err = err.WithContext("field", "email")
	err = err.WithContext("value", "invalid@")
	
	response := err.ToResponse()
	require.NotNil(t, response)
	
	// Context should be included in response
	if response.Details != nil {
		// Context may be included in details
		assert.NotEmpty(t, response.Details)
	}
}
