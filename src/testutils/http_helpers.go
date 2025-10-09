package testutils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// RequestBuilder helps build HTTP requests for testing
type RequestBuilder struct {
	method  string
	path    string
	body    []byte
	headers map[string]string
	query   url.Values
}

// NewRequestBuilder creates a new request builder
func NewRequestBuilder(method, path string) *RequestBuilder {
	return &RequestBuilder{
		method:  method,
		path:    path,
		headers: make(map[string]string),
		query:   make(url.Values),
	}
}

// WithBody sets the request body
func (rb *RequestBuilder) WithBody(body []byte) *RequestBuilder {
	rb.body = body
	return rb
}

// WithHeader adds a header to the request
func (rb *RequestBuilder) WithHeader(key, value string) *RequestBuilder {
	rb.headers[key] = value
	return rb
}

// WithQueryParam adds a query parameter
func (rb *RequestBuilder) WithQueryParam(key, value string) *RequestBuilder {
	rb.query.Add(key, value)
	return rb
}

// Build constructs the HTTP request
func (rb *RequestBuilder) Build() *http.Request {
	var body io.Reader
	if rb.body != nil {
		body = bytes.NewReader(rb.body)
	}
	
	req := httptest.NewRequest(rb.method, rb.path, body)
	
	// Add headers
	for key, value := range rb.headers {
		req.Header.Set(key, value)
	}
	
	// Add query parameters
	if len(rb.query) > 0 {
		req.URL.RawQuery = rb.query.Encode()
	}
	
	return req
}

// NewResponseRecorder creates a new response recorder
func NewResponseRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// MakeTestRequest executes a test request and returns the response recorder
func MakeTestRequest(router *gin.Engine, method, path string, body []byte, headers map[string]string) *httptest.ResponseRecorder {
	builder := NewRequestBuilder(method, path)
	
	if body != nil {
		builder.WithBody(body)
	}
	
	for key, value := range headers {
		builder.WithHeader(key, value)
	}
	
	req := builder.Build()
	rec := NewResponseRecorder()
	
	router.ServeHTTP(rec, req)
	
	return rec
}

// MakeJSONRequest executes a JSON request
func MakeJSONRequest(router *gin.Engine, method, path string, payload interface{}) *httptest.ResponseRecorder {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	
	return MakeTestRequest(router, method, path, jsonBytes, headers)
}

// AssertJSONResponse parses JSON response and performs basic validation
func AssertJSONResponse(t *testing.T, rec *httptest.ResponseRecorder, target interface{}) error {
	t.Helper()
	
	return ParseJSONResponse(rec, target)
}

// ParseJSONResponse parses JSON from response recorder
func ParseJSONResponse(rec *httptest.ResponseRecorder, target interface{}) error {
	return json.Unmarshal(rec.Body.Bytes(), target)
}

// CreateTestRouter creates a new Gin router in test mode
func CreateTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// CreateTestRouterWithMiddleware creates a router with middleware
func CreateTestRouterWithMiddleware(middleware ...gin.HandlerFunc) *gin.Engine {
	router := CreateTestRouter()
	router.Use(middleware...)
	return router
}

// AssertResponseCode asserts the response status code
func AssertResponseCode(t *testing.T, rec *httptest.ResponseRecorder, expectedCode int) {
	t.Helper()
	require.Equal(t, expectedCode, rec.Code, "Response code mismatch")
}

// AssertResponseBody asserts the response body contains expected string
func AssertResponseBody(t *testing.T, rec *httptest.ResponseRecorder, expected string) {
	t.Helper()
	require.Contains(t, rec.Body.String(), expected, "Response body doesn't contain expected string")
}

// AssertResponseHeader asserts a response header value
func AssertResponseHeader(t *testing.T, rec *httptest.ResponseRecorder, header, expected string) {
	t.Helper()
	actual := rec.Header().Get(header)
	require.Equal(t, expected, actual, "Response header mismatch for %s", header)
}
