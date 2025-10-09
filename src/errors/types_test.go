package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestErrorType_Basic tests basic error type functionality
func TestErrorType_Basic(t *testing.T) {
	err := NewAppError("test error", ErrorTypeValidation, nil)
	
	assert.NotNil(t, err)
	assert.Equal(t, "test error", err.Error())
	assert.Equal(t, ErrorTypeValidation, err.Type())
}

// TestErrorType_WithCause tests error wrapping
func TestErrorType_WithCause(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewAppError("wrapped error", ErrorTypeInternal, cause)
	
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "wrapped error")
	assert.True(t, errors.Is(err, cause))
}

// TestErrorType_HTTPStatus tests HTTP status code mapping
func TestErrorType_HTTPStatus(t *testing.T) {
	tests := []struct {
		name       string
		errorType  ErrorType
		wantStatus int
	}{
		{"Validation error", ErrorTypeValidation, http.StatusBadRequest},
		{"Not found error", ErrorTypeNotFound, http.StatusNotFound},
		{"Unsupported media", ErrorTypeUnsupportedMedia, http.StatusUnsupportedMediaType},
		{"Unprocessable entity", ErrorTypeUnprocessable, http.StatusUnprocessableEntity},
		{"Internal error", ErrorTypeInternal, http.StatusInternalServerError},
		{"Timeout error", ErrorTypeTimeout, http.StatusGatewayTimeout},
		{"Conflict error", ErrorTypeConflict, http.StatusConflict},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAppError("test", tt.errorType, nil)
			status := err.HTTPStatus()
			assert.Equal(t, tt.wantStatus, status)
		})
	}
}

// TestErrorType_UserFriendlyMessage tests user-facing messages
func TestErrorType_UserFriendlyMessage(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		errorType   ErrorType
		wantMessage string
	}{
		{"Custom message", "custom error", ErrorTypeValidation, "custom error"},
		{"Generic validation", "", ErrorTypeValidation, "Invalid request parameters"},
		{"Generic not found", "", ErrorTypeNotFound, "Resource not found"},
		{"Generic internal", "", ErrorTypeInternal, "Internal server error"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAppError(tt.message, tt.errorType, nil)
			msg := err.UserMessage()
			if tt.message != "" {
				assert.Equal(t, tt.message, msg)
			} else {
				assert.NotEmpty(t, msg)
				assert.Contains(t, msg, tt.wantMessage)
			}
		})
	}
}

// TestErrorType_Code tests error code generation
func TestErrorType_Code(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		wantCode  string
	}{
		{"Validation", ErrorTypeValidation, "VALIDATION_ERROR"},
		{"Not found", ErrorTypeNotFound, "NOT_FOUND"},
		{"Unsupported media", ErrorTypeUnsupportedMedia, "UNSUPPORTED_MEDIA_TYPE"},
		{"Unprocessable", ErrorTypeUnprocessable, "UNPROCESSABLE_ENTITY"},
		{"Internal", ErrorTypeInternal, "INTERNAL_ERROR"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAppError("test", tt.errorType, nil)
			assert.Equal(t, tt.wantCode, err.Code())
		})
	}
}

// TestErrorType_IsType tests error type checking
func TestErrorType_IsType(t *testing.T) {
	err := NewAppError("test", ErrorTypeValidation, nil)
	
	assert.True(t, err.IsType(ErrorTypeValidation))
	assert.False(t, err.IsType(ErrorTypeNotFound))
	assert.False(t, err.IsType(ErrorTypeInternal))
}

// TestErrorType_WithDetails tests adding details to errors
func TestErrorType_WithDetails(t *testing.T) {
	err := NewAppError("test", ErrorTypeValidation, nil)
	err = err.WithDetails(map[string]interface{}{
		"field": "email",
		"value": "invalid",
	})
	
	details := err.Details()
	require.NotNil(t, details)
	assert.Equal(t, "email", details["field"])
	assert.Equal(t, "invalid", details["value"])
}

// TestErrorType_Chain tests error chain unwrapping
func TestErrorType_Chain(t *testing.T) {
	original := errors.New("original error")
	wrapped1 := NewAppError("wrapped level 1", ErrorTypeInternal, original)
	wrapped2 := NewAppError("wrapped level 2", ErrorTypeValidation, wrapped1)
	
	// Should be able to unwrap to original
	assert.True(t, errors.Is(wrapped2, original))
	assert.True(t, errors.Is(wrapped2, wrapped1))
}

// TestErrorType_AsAppError tests type assertion
func TestErrorType_AsAppError(t *testing.T) {
	appErr := NewAppError("test", ErrorTypeValidation, nil)
	
	var target *AppError
	assert.True(t, errors.As(appErr, &target))
	assert.Equal(t, ErrorTypeValidation, target.Type())
}

// TestErrorType_StandardError tests compatibility with standard error interface
func TestErrorType_StandardError(t *testing.T) {
	err := NewAppError("test error", ErrorTypeInternal, nil)
	
	// Should work with error interface
	var e error = err
	assert.Equal(t, "test error", e.Error())
}

// TestErrorResponse_Structure tests error response JSON structure
func TestErrorResponse_Structure(t *testing.T) {
	err := NewAppError("test error", ErrorTypeValidation, nil)
	response := err.ToResponse()
	
	assert.NotNil(t, response)
	assert.Equal(t, "test error", response.Error)
	assert.Equal(t, "VALIDATION_ERROR", response.Code)
	assert.Equal(t, http.StatusBadRequest, response.Status)
}

// TestErrorResponse_WithRequestID tests request ID in error response
func TestErrorResponse_WithRequestID(t *testing.T) {
	err := NewAppError("test error", ErrorTypeInternal, nil)
	err = err.WithRequestID("req-12345")
	response := err.ToResponse()
	
	assert.Equal(t, "req-12345", response.RequestID)
}

// TestValidationError_FieldErrors tests field-specific validation errors
func TestValidationError_FieldErrors(t *testing.T) {
	err := NewValidationError("validation failed")
	err = err.AddFieldError("email", "invalid format")
	err = err.AddFieldError("age", "must be positive")
	
	assert.Equal(t, 2, len(err.FieldErrors()))
	assert.Contains(t, err.FieldErrors(), "email")
	assert.Contains(t, err.FieldErrors(), "age")
}

// TestImageError_ImageSpecific tests image-specific error types
func TestImageError_ImageSpecific(t *testing.T) {
	tests := []struct {
		name      string
		createErr func() error
		wantType  ErrorType
	}{
		{"Image not found", func() error { return NewImageNotFoundError("test.jpg") }, ErrorTypeNotFound},
		{"Corrupted image", func() error { return NewCorruptedImageError("test.jpg") }, ErrorTypeUnprocessable},
		{"Unsupported format", func() error { return NewUnsupportedFormatError("bmp") }, ErrorTypeUnsupportedMedia},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.createErr()
			require.NotNil(t, err)
			
			var appErr *AppError
			require.True(t, errors.As(err, &appErr))
			assert.Equal(t, tt.wantType, appErr.Type())
		})
	}
}

// TestProcessingError_Processing tests processing-specific errors
func TestProcessingError_Processing(t *testing.T) {
	cause := errors.New("resize failed")
	err := NewProcessingError("image processing failed", cause)
	
	var appErr *AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, ErrorTypeInternal, appErr.Type())
	assert.True(t, errors.Is(err, cause))
}

// TestCacheError_Cache tests cache-specific errors
func TestCacheError_Cache(t *testing.T) {
	err := NewCacheError("cache write failed")
	
	var appErr *AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, ErrorTypeInternal, appErr.Type())
}
