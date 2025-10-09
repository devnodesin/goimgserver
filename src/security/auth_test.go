package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthentication_AdminEndpoints_TokenValidation tests token-based authentication
func TestAuthentication_AdminEndpoints_TokenValidation(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		expectedStatus int
		shouldPass     bool
	}{
		{
			name:           "valid_token",
			token:          "valid-test-token-12345",
			expectedStatus: http.StatusOK,
			shouldPass:     true,
		},
		{
			name:           "invalid_token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
			shouldPass:     false,
		},
		{
			name:           "empty_token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			shouldPass:     false,
		},
		{
			name:           "malformed_token",
			token:          "mal<script>alert('xss')</script>formed",
			expectedStatus: http.StatusUnauthorized,
			shouldPass:     false,
		},
	}

	// Create token authenticator with a valid token
	auth := NewTokenAuthenticator([]string{"valid-test-token-12345"})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			router.Use(TokenAuthMiddleware(auth))
			router.POST("/admin/clear", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "cleared"})
			})

			req := httptest.NewRequest("POST", "/admin/clear", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestAuthentication_AdminEndpoints_TokenExpiry tests token expiration
func TestAuthentication_AdminEndpoints_TokenExpiry(t *testing.T) {
	// Create auth with expiring tokens
	auth := NewExpiringTokenAuthenticator()

	// Generate a token that expires in 100ms
	token := auth.GenerateToken(100 * time.Millisecond)
	require.NotEmpty(t, token)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(TokenAuthMiddleware(auth))
	router.POST("/admin/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Should work immediately
	req1 := httptest.NewRequest("POST", "/admin/test", nil)
	req1.Header.Set("Authorization", "Bearer "+token)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code, "Token should be valid immediately")

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should fail after expiration
	req2 := httptest.NewRequest("POST", "/admin/test", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code, "Token should be expired")
}

// TestAuthentication_APIKey_Validation tests API key authentication
func TestAuthentication_APIKey_Validation(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		expectedStatus int
	}{
		{
			name:           "valid_api_key",
			apiKey:         "valid-api-key-12345",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid_api_key",
			apiKey:         "invalid-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "empty_api_key",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "sql_injection_in_key",
			apiKey:         "'; DROP TABLE keys;--",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	// Create API key authenticator
	auth := NewAPIKeyAuthenticator([]string{"valid-api-key-12345"})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			router.Use(APIKeyAuthMiddleware(auth))
			router.GET("/api/images", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"images": []string{}})
			})

			req := httptest.NewRequest("GET", "/api/images", nil)
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestAuthentication_APIKey_Storage tests secure key storage
func TestAuthentication_APIKey_Storage(t *testing.T) {
	auth := NewAPIKeyAuthenticator([]string{})

	// Add API keys
	key1 := auth.GenerateAPIKey()
	key2 := auth.GenerateAPIKey()

	require.NotEmpty(t, key1)
	require.NotEmpty(t, key2)
	require.NotEqual(t, key1, key2, "Generated keys should be unique")

	// Keys should be at least 32 characters for security
	assert.GreaterOrEqual(t, len(key1), 32, "API keys should be at least 32 characters")
	assert.GreaterOrEqual(t, len(key2), 32, "API keys should be at least 32 characters")

	// Validate generated keys
	assert.True(t, auth.ValidateAPIKey(key1), "Generated key1 should be valid")
	assert.True(t, auth.ValidateAPIKey(key2), "Generated key2 should be valid")

	// Revoke a key
	auth.RevokeAPIKey(key1)
	assert.False(t, auth.ValidateAPIKey(key1), "Revoked key should be invalid")
	assert.True(t, auth.ValidateAPIKey(key2), "Other keys should still be valid")
}

// TestAuthorization_PermissionBased_Access tests permission checking
func TestAuthorization_PermissionBased_Access(t *testing.T) {
	tests := []struct {
		name           string
		userRole       string
		requiredPerm   string
		expectedStatus int
	}{
		{
			name:           "admin_has_all_permissions",
			userRole:       "admin",
			requiredPerm:   "cache:clear",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user_has_read_permission",
			userRole:       "user",
			requiredPerm:   "image:read",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user_lacks_admin_permission",
			userRole:       "user",
			requiredPerm:   "cache:clear",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "guest_has_no_permissions",
			userRole:       "guest",
			requiredPerm:   "image:read",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			// Setup authorization with roles
			authz := NewAuthorizer()
			authz.AddRole("admin", []string{"cache:clear", "image:read", "image:write"})
			authz.AddRole("user", []string{"image:read"})
			authz.AddRole("guest", []string{})

			// Middleware that sets the role
			router.Use(func(c *gin.Context) {
				c.Set("role", tt.userRole)
				c.Next()
			})

			router.Use(RequirePermission(authz, tt.requiredPerm))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestAuthorization_PermissionBased_Denial tests access denial
func TestAuthorization_PermissionBased_Denial(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	authz := NewAuthorizer()
	authz.AddRole("user", []string{"image:read"})

	// Try to access admin endpoint without proper role
	router.Use(func(c *gin.Context) {
		c.Set("role", "user")
		c.Next()
	})

	router.Use(RequirePermission(authz, "admin:full"))
	router.POST("/admin/dangerous", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
	})

	req := httptest.NewRequest("POST", "/admin/dangerous", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.NotContains(t, w.Body.String(), "should not reach here")
}

// TestAuthentication_TokenGeneration_Entropy tests token randomness
func TestAuthentication_TokenGeneration_Entropy(t *testing.T) {
	auth := NewExpiringTokenAuthenticator()

	// Generate multiple tokens
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token := auth.GenerateToken(time.Hour)
		require.NotEmpty(t, token)

		// Check uniqueness
		assert.False(t, tokens[token], "Token should be unique")
		tokens[token] = true

		// Check length (should be at least 32 chars for good entropy)
		assert.GreaterOrEqual(t, len(token), 32)
	}
}

// TestAuthentication_BearerToken_Extraction tests bearer token extraction
func TestAuthentication_BearerToken_Extraction(t *testing.T) {
	tests := []struct {
		name          string
		authHeader    string
		expectedToken string
		shouldExtract bool
	}{
		{
			name:          "valid_bearer_token",
			authHeader:    "Bearer valid-token-123",
			expectedToken: "valid-token-123",
			shouldExtract: true,
		},
		{
			name:          "bearer_lowercase",
			authHeader:    "bearer valid-token-123",
			expectedToken: "valid-token-123",
			shouldExtract: true,
		},
		{
			name:          "no_bearer_prefix",
			authHeader:    "valid-token-123",
			expectedToken: "",
			shouldExtract: false,
		},
		{
			name:          "empty_header",
			authHeader:    "",
			expectedToken: "",
			shouldExtract: false,
		},
		{
			name:          "malformed_header",
			authHeader:    "Bearer",
			expectedToken: "",
			shouldExtract: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, ok := ExtractBearerToken(tt.authHeader)
			assert.Equal(t, tt.shouldExtract, ok)
			if tt.shouldExtract {
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}
}

// TestAuthentication_CombinedAuth tests combining multiple auth methods
func TestAuthentication_CombinedAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	tokenAuth := NewTokenAuthenticator([]string{"valid-token"})
	apiKeyAuth := NewAPIKeyAuthenticator([]string{"valid-api-key"})

	// Use OR authentication (either token or API key works)
	router.Use(CombinedAuthMiddleware(tokenAuth, apiKeyAuth))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	tests := []struct {
		name           string
		tokenHeader    string
		apiKeyHeader   string
		expectedStatus int
	}{
		{
			name:           "valid_token",
			tokenHeader:    "Bearer valid-token",
			apiKeyHeader:   "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid_api_key",
			tokenHeader:    "",
			apiKeyHeader:   "valid-api-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "both_invalid",
			tokenHeader:    "Bearer invalid",
			apiKeyHeader:   "invalid",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "no_credentials",
			tokenHeader:    "",
			apiKeyHeader:   "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/protected", nil)
			if tt.tokenHeader != "" {
				req.Header.Set("Authorization", tt.tokenHeader)
			}
			if tt.apiKeyHeader != "" {
				req.Header.Set("X-API-Key", tt.apiKeyHeader)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestAuthentication_TokenAuth_EdgeCases tests edge cases in token auth
func TestAuthentication_TokenAuth_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Test with regular token auth
	regularAuth := NewTokenAuthenticator([]string{"valid"})
	router1 := gin.New()
	router1.Use(TokenAuthMiddleware(regularAuth))
	router1.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	
	// Test bearer token extraction edge cases
	tests := []struct {
		name   string
		header string
		expectOK bool
	}{
		{"multiple_spaces", "Bearer  token", true},
		{"trailing_space", "Bearer token ", true},
		{"extra_bearer", "Bearer Bearer token", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, ok := ExtractBearerToken(tt.header)
			assert.True(t, ok)
			assert.NotEmpty(t, token)
		})
	}
}

// TestAuthorization_EdgeCases tests authorization edge cases
func TestAuthorization_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authz := NewAuthorizer()
	authz.AddRole("admin", []string{"read", "write"})
	
	// Test permission check for non-existent role
	assert.False(t, authz.HasPermission("nonexistent", "read"))
	
	// Test permission check for valid role without permission
	assert.False(t, authz.HasPermission("admin", "delete"))
	
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Don't set role - test missing role
		c.Next()
	})
	router.Use(RequirePermission(authz, "read"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
	
	// Test with wrong role type
	router2 := gin.New()
	router2.Use(func(c *gin.Context) {
		c.Set("role", 123) // Wrong type
		c.Next()
	})
	router2.Use(RequirePermission(authz, "read"))
	router2.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	router2.ServeHTTP(w2, req2)
	
	assert.Equal(t, http.StatusForbidden, w2.Code)
}

// TestValidation_EdgeCases tests validation edge cases
func TestValidation_EdgeCases(t *testing.T) {
	// Test file extension validation edge cases
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"dotfile_with_ext", ".hidden.jpg", false},
		{"multiple_dots_hidden", "...test.png", false},
		{"just_extension", ".jpg", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileExtension(tt.filename)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
	
	// Test dimension validation with aspect ratio
	w, h, err := ValidateDimensions(800, 0)
	assert.NoError(t, err)
	assert.Equal(t, 800, w)
	assert.Equal(t, 0, h)
}

// TestResourceProtection_EdgeCases tests resource protection edge cases
func TestResourceProtection_EdgeCases(t *testing.T) {
	// Test disk space monitor with invalid path
	monitor := NewDiskSpaceMonitor("/nonexistent/path", 1000)
	_, err := monitor.GetUsage()
	assert.Error(t, err)
	
	// IsThresholdExceeded should handle errors gracefully
	exceeded := monitor.IsThresholdExceeded()
	assert.False(t, exceeded)
	
	// Test with very high threshold (should not be exceeded)
	monitor2 := NewDiskSpaceMonitor("/tmp", 1000*1024*1024*1024*1024) // 1 PB
	usage, err := monitor2.GetUsage()
	assert.NoError(t, err)
	assert.NotNil(t, usage)
	exceeded2 := monitor2.IsThresholdExceeded()
	assert.False(t, exceeded2)
}

// TestParseAndValidateParameters_FullCoverage tests all parameter parsing paths
func TestParseAndValidateParameters_FullCoverage(t *testing.T) {
	tests := []struct {
		name     string
		segments []string
		checkFn  func(t *testing.T, params ProcessingParams)
	}{
		{
			name:     "quality_first_wins",
			segments: []string{"q80", "q90"},
			checkFn: func(t *testing.T, params ProcessingParams) {
				assert.Equal(t, 80, params.Quality)
			},
		},
		{
			name:     "format_first_wins",
			segments: []string{"png", "jpeg"},
			checkFn: func(t *testing.T, params ProcessingParams) {
				assert.Equal(t, "png", params.Format)
			},
		},
		{
			name:     "width_only",
			segments: []string{"600"},
			checkFn: func(t *testing.T, params ProcessingParams) {
				assert.Equal(t, 600, params.Width)
				assert.Equal(t, 0, params.Height)
			},
		},
		{
			name:     "invalid_quality_uses_default",
			segments: []string{"q999"},
			checkFn: func(t *testing.T, params ProcessingParams) {
				assert.Equal(t, DefaultQuality, params.Quality)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := ParseAndValidateParameters(tt.segments)
			tt.checkFn(t, params)
		})
	}
}
