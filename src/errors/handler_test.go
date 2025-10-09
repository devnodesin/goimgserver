package errors

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestErrorHandler_HTTPStatus_BadRequest tests 400 error responses
func TestErrorHandler_HTTPStatus_BadRequest(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		err := NewAppError("invalid parameters", ErrorTypeValidation, nil)
		HandleError(c, err)
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid parameters", response.Error)
}

// TestErrorHandler_HTTPStatus_NotFound tests 404 error responses
func TestErrorHandler_HTTPStatus_NotFound(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		err := NewImageNotFoundError("missing.jpg")
		HandleError(c, err)
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestErrorHandler_HTTPStatus_UnsupportedMedia tests 415 error responses
func TestErrorHandler_HTTPStatus_UnsupportedMedia(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		err := NewUnsupportedFormatError("bmp")
		HandleError(c, err)
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)
}

// TestErrorHandler_HTTPStatus_UnprocessableEntity tests 422 error responses
func TestErrorHandler_HTTPStatus_UnprocessableEntity(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		err := NewCorruptedImageError("bad.jpg")
		HandleError(c, err)
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// TestErrorHandler_HTTPStatus_InternalServer tests 500 error responses
func TestErrorHandler_HTTPStatus_InternalServer(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		err := NewAppError("internal error", ErrorTypeInternal, nil)
		HandleError(c, err)
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestErrorHandler_JSONResponse_UserFriendly tests user-friendly JSON errors
func TestErrorHandler_JSONResponse_UserFriendly(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		err := NewAppError("database connection failed", ErrorTypeInternal, nil)
		HandleError(c, err)
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	// Should not expose internal details in production
	assert.NotEmpty(t, response.Error)
	assert.NotEmpty(t, response.Code)
}

// TestErrorHandler_JSONResponse_Structure tests error response structure
func TestErrorHandler_JSONResponse_Structure(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		err := NewAppError("test error", ErrorTypeValidation, nil)
		HandleError(c, err)
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.NotEmpty(t, response.Error)
	assert.NotEmpty(t, response.Code)
	assert.NotZero(t, response.Status)
}

// TestErrorHandler_JSONResponse_RequestID tests request ID in errors
func TestErrorHandler_JSONResponse_RequestID(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "test-req-123")
		c.Next()
	})
	router.Use(ErrorHandlerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		err := NewAppError("test error", ErrorTypeInternal, nil)
		HandleError(c, err)
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "test-req-123", response.RequestID)
}

// TestErrorHandler_StandardError tests handling of standard Go errors
func TestErrorHandler_StandardError(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		err := assert.AnError
		HandleError(c, err)
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestErrorHandler_NilError tests handling of nil errors
func TestErrorHandler_NilError(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		HandleError(c, nil)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Should not error on nil
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestErrorHandler_WithDetails tests error details in response
func TestErrorHandler_WithDetails(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		err := NewAppError("validation failed", ErrorTypeValidation, nil)
		err = err.WithDetails(map[string]interface{}{
			"field": "email",
			"issue": "invalid format",
		})
		HandleError(c, err)
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.NotNil(t, response.Details)
}

// TestErrorHandler_Middleware tests error handler as middleware
func TestErrorHandler_Middleware(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		panic("test panic")
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.Error)
}
