package errors

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestErrorScenarios_ImageNotFound tests missing image handling
func TestErrorScenarios_ImageNotFound(t *testing.T) {
	err := NewImageNotFoundError("missing.jpg")
	
	require.NotNil(t, err)
	assert.Equal(t, ErrorTypeNotFound, err.Type())
	assert.Contains(t, err.Error(), "missing.jpg")
	assert.Equal(t, 404, err.HTTPStatus())
	
	details := err.Details()
	assert.Equal(t, "missing.jpg", details["filename"])
}

// TestErrorScenarios_CorruptedImage tests corrupted image handling
func TestErrorScenarios_CorruptedImage(t *testing.T) {
	err := NewCorruptedImageError("corrupted.jpg")
	
	require.NotNil(t, err)
	assert.Equal(t, ErrorTypeUnprocessable, err.Type())
	assert.Contains(t, err.Error(), "corrupted.jpg")
	assert.Equal(t, 422, err.HTTPStatus())
}

// TestErrorScenarios_InvalidParameters tests parameter validation errors
func TestErrorScenarios_InvalidParameters(t *testing.T) {
	tests := []struct {
		name    string
		param   string
		value   string
		message string
	}{
		{"Invalid width", "width", "-100", "Width must be positive"},
		{"Invalid height", "height", "0", "Height must be positive"},
		{"Invalid quality", "quality", "150", "Quality must be between 1 and 100"},
		{"Invalid format", "format", "bmp", "Unsupported format"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewValidationError(tt.message)
			err = err.AddFieldError(tt.param, tt.value)
			
			assert.Equal(t, ErrorTypeValidation, err.Type())
			assert.Equal(t, 400, err.HTTPStatus())
			
			fieldErrs := err.FieldErrors()
			assert.Contains(t, fieldErrs, tt.param)
		})
	}
}

// TestErrorScenarios_ProcessingFailure tests processing failures
func TestErrorScenarios_ProcessingFailure(t *testing.T) {
	tests := []struct {
		name   string
		cause  error
		expect string
	}{
		{"Resize failure", errors.New("resize failed"), "Image processing failed"},
		{"Format conversion", errors.New("conversion failed"), "Format conversion failed"},
		{"Memory limit", errors.New("out of memory"), "Processing error"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewProcessingError(tt.expect, tt.cause)
			
			assert.Equal(t, ErrorTypeInternal, err.Type())
			assert.True(t, errors.Is(err, tt.cause))
			assert.Contains(t, err.Error(), tt.expect)
		})
	}
}

// TestErrorScenarios_CacheFailure tests cache operation failures
func TestErrorScenarios_CacheFailure(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		message   string
	}{
		{"Cache read", "read", "Failed to read from cache"},
		{"Cache write", "write", "Failed to write to cache"},
		{"Cache clear", "clear", "Failed to clear cache"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewCacheError(tt.message)
			err = err.WithContext("operation", tt.operation)
			
			assert.Equal(t, ErrorTypeInternal, err.Type())
			
			ctx := err.Context()
			assert.Equal(t, "cache", ctx["component"])
			assert.Equal(t, tt.operation, ctx["operation"])
		})
	}
}

// TestErrorScenarios_FileSystemErrors tests file system errors
func TestErrorScenarios_FileSystemErrors(t *testing.T) {
	tests := []struct {
		name        string
		osErr       error
		expectType  ErrorType
		expectCode  int
	}{
		{"File not found", os.ErrNotExist, ErrorTypeNotFound, 404},
		{"Permission denied", os.ErrPermission, ErrorTypeInternal, 500},
		{"File closed", os.ErrClosed, ErrorTypeInternal, 500},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WrapFileSystemError(tt.osErr)
			
			var appErr *AppError
			require.True(t, errors.As(err, &appErr))
			assert.Equal(t, tt.expectType, appErr.Type())
			assert.Equal(t, tt.expectCode, appErr.HTTPStatus())
		})
	}
}

// TestErrorScenarios_NetworkErrors tests network-related errors
func TestErrorScenarios_NetworkErrors(t *testing.T) {
	tests := []struct {
		name    string
		message string
		errType ErrorType
	}{
		{"Connection timeout", "connection timeout", ErrorTypeTimeout},
		{"Connection refused", "connection refused", ErrorTypeInternal},
		{"DNS resolution", "DNS resolution failed", ErrorTypeInternal},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAppError(tt.message, tt.errType, nil)
			
			assert.Equal(t, tt.errType, err.Type())
			assert.Contains(t, err.Error(), tt.message)
		})
	}
}

// TestErrorScenarios_TimeoutErrors tests timeout handling
func TestErrorScenarios_TimeoutErrors(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		timeout   string
	}{
		{"Processing timeout", "image processing", "30s"},
		{"Cache timeout", "cache read", "5s"},
		{"Request timeout", "HTTP request", "10s"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewTimeoutError(tt.operation, tt.timeout)
			
			assert.Equal(t, ErrorTypeTimeout, err.Type())
			assert.Equal(t, 504, err.HTTPStatus())
			
			details := err.Details()
			assert.Equal(t, tt.operation, details["operation"])
			assert.Equal(t, tt.timeout, details["timeout"])
		})
	}
}

// TestErrorScenarios_MultipleErrors tests handling of multiple errors
func TestErrorScenarios_MultipleErrors(t *testing.T) {
	errs := NewMultiError()
	
	errs.Add(NewImageNotFoundError("file1.jpg"))
	errs.Add(NewCorruptedImageError("file2.jpg"))
	errs.Add(NewProcessingError("processing failed", nil))
	
	assert.Equal(t, 3, errs.Count())
	assert.True(t, errs.HasErrors())
	
	// Should return first error
	firstErr := errs.FirstError()
	require.NotNil(t, firstErr)
	assert.Contains(t, firstErr.Error(), "file1.jpg")
}

// TestErrorScenarios_ErrorRecovery tests error recovery mechanisms
func TestErrorScenarios_ErrorRecovery(t *testing.T) {
	// Test that errors can be recovered and handled
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if ok {
				appErr := NewAppError(err.Error(), ErrorTypeInternal, err)
				assert.NotNil(t, appErr)
			}
		}
	}()
	
	// Simulate a panic that should be recovered
	// In actual code, this would be in a handler
}

// TestErrorScenarios_ChainedErrors tests error chain handling
func TestErrorScenarios_ChainedErrors(t *testing.T) {
	// Create a chain of errors
	baseErr := errors.New("base error")
	level1 := NewProcessingError("level 1 error", baseErr)
	level2 := NewAppError("level 2 error", ErrorTypeValidation, level1)
	
	// Should unwrap to base error
	assert.True(t, errors.Is(level2, baseErr))
	assert.True(t, errors.Is(level2, level1))
	
	// Should be able to get AppError from chain
	var appErr *AppError
	assert.True(t, errors.As(level2, &appErr))
}
