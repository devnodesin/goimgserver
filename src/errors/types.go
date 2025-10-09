package errors

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"sync"
)

// ErrorType represents the category of an error
type ErrorType string

const (
	ErrorTypeValidation      ErrorType = "validation"
	ErrorTypeNotFound        ErrorType = "not_found"
	ErrorTypeUnsupportedMedia ErrorType = "unsupported_media"
	ErrorTypeUnprocessable   ErrorType = "unprocessable"
	ErrorTypeInternal        ErrorType = "internal"
	ErrorTypeTimeout         ErrorType = "timeout"
	ErrorTypeConflict        ErrorType = "conflict"
)

var (
	developmentMode bool
	modeMutex       sync.RWMutex
)

// SetDevelopmentMode sets the development mode flag
func SetDevelopmentMode(enabled bool) {
	modeMutex.Lock()
	defer modeMutex.Unlock()
	developmentMode = enabled
}

// IsDevelopmentMode returns whether development mode is enabled
func IsDevelopmentMode() bool {
	modeMutex.RLock()
	defer modeMutex.RUnlock()
	return developmentMode
}

// AppError represents an application error with context
type AppError struct {
	message    string
	errorType  ErrorType
	cause      error
	ctx        map[string]interface{}
	metadata   map[string]interface{}
	requestID  string
	stackTrace string
	mu         sync.RWMutex
}

// NewAppError creates a new application error
func NewAppError(message string, errorType ErrorType, cause error) *AppError {
	err := &AppError{
		message:   message,
		errorType: errorType,
		cause:     cause,
		ctx:       make(map[string]interface{}),
		metadata:  make(map[string]interface{}),
	}
	
	if IsDevelopmentMode() {
		err.stackTrace = string(debug.Stack())
	}
	
	return err
}

// NewAppErrorWithContext creates an error with request context
func NewAppErrorWithContext(ctx context.Context, message string, errorType ErrorType, cause error) *AppError {
	err := NewAppError(message, errorType, cause)
	
	// Extract common values from context
	if reqID, ok := ctx.Value("request_id").(string); ok {
		err.requestID = reqID
	}
	if userID, ok := ctx.Value("user_id").(string); ok {
		err.ctx["user_id"] = userID
	}
	
	return err
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

// Unwrap implements error unwrapping
func (e *AppError) Unwrap() error {
	return e.cause
}

// Type returns the error type
func (e *AppError) Type() ErrorType {
	return e.errorType
}

// IsType checks if error is of a specific type
func (e *AppError) IsType(t ErrorType) bool {
	return e.errorType == t
}

// HTTPStatus returns the appropriate HTTP status code
func (e *AppError) HTTPStatus() int {
	switch e.errorType {
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeUnsupportedMedia:
		return http.StatusUnsupportedMediaType
	case ErrorTypeUnprocessable:
		return http.StatusUnprocessableEntity
	case ErrorTypeTimeout:
		return http.StatusGatewayTimeout
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeInternal:
		fallthrough
	default:
		return http.StatusInternalServerError
	}
}

// Code returns the error code
func (e *AppError) Code() string {
	switch e.errorType {
	case ErrorTypeValidation:
		return "VALIDATION_ERROR"
	case ErrorTypeNotFound:
		return "NOT_FOUND"
	case ErrorTypeUnsupportedMedia:
		return "UNSUPPORTED_MEDIA_TYPE"
	case ErrorTypeUnprocessable:
		return "UNPROCESSABLE_ENTITY"
	case ErrorTypeTimeout:
		return "TIMEOUT"
	case ErrorTypeConflict:
		return "CONFLICT"
	case ErrorTypeInternal:
		return "INTERNAL_ERROR"
	default:
		return "UNKNOWN_ERROR"
	}
}

// UserMessage returns a user-friendly error message
func (e *AppError) UserMessage() string {
	if e.message != "" {
		// In production mode, sanitize internal error messages
		if !IsDevelopmentMode() && e.errorType == ErrorTypeInternal {
			return "Internal server error"
		}
		return e.message
	}
	
	// Default messages based on error type
	switch e.errorType {
	case ErrorTypeValidation:
		return "Invalid request parameters"
	case ErrorTypeNotFound:
		return "Resource not found"
	case ErrorTypeUnsupportedMedia:
		return "Unsupported media type"
	case ErrorTypeUnprocessable:
		return "Unable to process the request"
	case ErrorTypeTimeout:
		return "Request timeout"
	case ErrorTypeConflict:
		return "Resource conflict"
	case ErrorTypeInternal:
		return "Internal server error"
	default:
		return "An error occurred"
	}
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if e.metadata == nil {
		e.metadata = make(map[string]interface{})
	}
	for k, v := range details {
		e.metadata[k] = v
	}
	return e
}

// Details returns the error details
func (e *AppError) Details() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	result := make(map[string]interface{})
	for k, v := range e.metadata {
		result[k] = v
	}
	return result
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if e.ctx == nil {
		e.ctx = make(map[string]interface{})
	}
	e.ctx[key] = value
	return e
}

// Context returns the error context
func (e *AppError) Context() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	result := make(map[string]interface{})
	for k, v := range e.ctx {
		result[k] = v
	}
	return result
}

// WithMetadata adds metadata to the error
func (e *AppError) WithMetadata(key string, value interface{}) *AppError {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if e.metadata == nil {
		e.metadata = make(map[string]interface{})
	}
	e.metadata[key] = value
	return e
}

// Metadata returns the error metadata
func (e *AppError) Metadata() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	result := make(map[string]interface{})
	for k, v := range e.metadata {
		result[k] = v
	}
	return result
}

// WithRequestID sets the request ID
func (e *AppError) WithRequestID(requestID string) *AppError {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.requestID = requestID
	return e
}

// StackTrace returns the stack trace if available
func (e *AppError) StackTrace() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stackTrace
}

// Clone creates a copy of the error
func (e *AppError) Clone() *AppError {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	cloned := &AppError{
		message:    e.message,
		errorType:  e.errorType,
		cause:      e.cause,
		requestID:  e.requestID,
		stackTrace: e.stackTrace,
		ctx:        make(map[string]interface{}),
		metadata:   make(map[string]interface{}),
	}
	
	for k, v := range e.ctx {
		cloned.ctx[k] = v
	}
	for k, v := range e.metadata {
		cloned.metadata[k] = v
	}
	
	return cloned
}

// ErrorResponse represents the JSON error response structure
type ErrorResponse struct {
	Error     string                 `json:"error"`
	Code      string                 `json:"code"`
	Status    int                    `json:"status"`
	RequestID string                 `json:"request_id,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// ToResponse converts the error to an HTTP response structure
func (e *AppError) ToResponse() ErrorResponse {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	response := ErrorResponse{
		Error:     e.UserMessage(),
		Code:      e.Code(),
		Status:    e.HTTPStatus(),
		RequestID: e.requestID,
	}
	
	// Include details/metadata if present
	if len(e.metadata) > 0 {
		response.Details = e.Details()
	}
	
	return response
}

// ValidationError represents a validation error with field-specific errors
type ValidationError struct {
	*AppError
	fieldErrors map[string]string
	mu          sync.RWMutex
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		AppError:    NewAppError(message, ErrorTypeValidation, nil),
		fieldErrors: make(map[string]string),
	}
}

// AddFieldError adds a field-specific error
func (v *ValidationError) AddFieldError(field, message string) *ValidationError {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.fieldErrors[field] = message
	return v
}

// FieldErrors returns all field errors
func (v *ValidationError) FieldErrors() map[string]string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	result := make(map[string]string)
	for k, v := range v.fieldErrors {
		result[k] = v
	}
	return result
}

// Image-specific errors

// NewImageNotFoundError creates an image not found error
func NewImageNotFoundError(filename string) *AppError {
	return NewAppError(
		fmt.Sprintf("Image not found: %s", filename),
		ErrorTypeNotFound,
		nil,
	).WithDetails(map[string]interface{}{
		"filename": filename,
	})
}

// NewCorruptedImageError creates a corrupted image error
func NewCorruptedImageError(filename string) *AppError {
	return NewAppError(
		fmt.Sprintf("Image is corrupted or invalid: %s", filename),
		ErrorTypeUnprocessable,
		nil,
	).WithDetails(map[string]interface{}{
		"filename": filename,
	})
}

// NewUnsupportedFormatError creates an unsupported format error
func NewUnsupportedFormatError(format string) *AppError {
	return NewAppError(
		fmt.Sprintf("Unsupported image format: %s", format),
		ErrorTypeUnsupportedMedia,
		nil,
	).WithDetails(map[string]interface{}{
		"format": format,
	})
}

// NewProcessingError creates a processing error
func NewProcessingError(message string, cause error) *AppError {
	return NewAppError(message, ErrorTypeInternal, cause)
}

// NewCacheError creates a cache error
func NewCacheError(message string) *AppError {
	return NewAppError(message, ErrorTypeInternal, nil).
		WithContext("component", "cache")
}
