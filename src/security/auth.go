package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// TokenAuthenticator handles token-based authentication
type TokenAuthenticator struct {
	validTokens map[string]bool
	mu          sync.RWMutex
}

// NewTokenAuthenticator creates a new token authenticator
func NewTokenAuthenticator(tokens []string) *TokenAuthenticator {
	auth := &TokenAuthenticator{
		validTokens: make(map[string]bool),
	}
	for _, token := range tokens {
		auth.validTokens[token] = true
	}
	return auth
}

// ValidateToken validates a token
func (a *TokenAuthenticator) ValidateToken(token string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.validTokens[token]
}

// ExpiringTokenAuthenticator handles tokens with expiration
type ExpiringTokenAuthenticator struct {
	tokens map[string]time.Time
	mu     sync.RWMutex
}

// NewExpiringTokenAuthenticator creates a new expiring token authenticator
func NewExpiringTokenAuthenticator() *ExpiringTokenAuthenticator {
	return &ExpiringTokenAuthenticator{
		tokens: make(map[string]time.Time),
	}
}

// GenerateToken generates a new random token with expiration
func (a *ExpiringTokenAuthenticator) GenerateToken(duration time.Duration) string {
	// Generate 32 bytes of random data
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}

	token := hex.EncodeToString(b)
	expiresAt := time.Now().Add(duration)

	a.mu.Lock()
	a.tokens[token] = expiresAt
	a.mu.Unlock()

	return token
}

// ValidateToken validates a token and checks expiration
func (a *ExpiringTokenAuthenticator) ValidateToken(token string) bool {
	a.mu.RLock()
	expiresAt, exists := a.tokens[token]
	a.mu.RUnlock()

	if !exists {
		return false
	}

	return time.Now().Before(expiresAt)
}

// TokenAuthMiddleware creates middleware for token authentication
func TokenAuthMiddleware(auth interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract bearer token
		authHeader := c.GetHeader("Authorization")
		token, ok := ExtractBearerToken(authHeader)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing or invalid authorization header",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		// Validate token
		var valid bool
		switch a := auth.(type) {
		case *TokenAuthenticator:
			valid = a.ValidateToken(token)
		case *ExpiringTokenAuthenticator:
			valid = a.ValidateToken(token)
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid authenticator type",
				"code":  "INTERNAL_ERROR",
			})
			return
		}

		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		c.Next()
	}
}

// ExtractBearerToken extracts a bearer token from an Authorization header
func ExtractBearerToken(authHeader string) (string, bool) {
	if authHeader == "" {
		return "", false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return "", false
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}

	return token, true
}

// APIKeyAuthenticator handles API key authentication
type APIKeyAuthenticator struct {
	keys map[string]bool
	mu   sync.RWMutex
}

// NewAPIKeyAuthenticator creates a new API key authenticator
func NewAPIKeyAuthenticator(keys []string) *APIKeyAuthenticator {
	auth := &APIKeyAuthenticator{
		keys: make(map[string]bool),
	}
	for _, key := range keys {
		auth.keys[key] = true
	}
	return auth
}

// GenerateAPIKey generates a new random API key
func (a *APIKeyAuthenticator) GenerateAPIKey() string {
	// Generate 32 bytes of random data
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}

	key := hex.EncodeToString(b)

	a.mu.Lock()
	a.keys[key] = true
	a.mu.Unlock()

	return key
}

// ValidateAPIKey validates an API key using constant-time comparison
func (a *APIKeyAuthenticator) ValidateAPIKey(key string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Use constant-time comparison to prevent timing attacks
	for validKey := range a.keys {
		if subtle.ConstantTimeCompare([]byte(key), []byte(validKey)) == 1 {
			return true
		}
	}
	return false
}

// RevokeAPIKey revokes an API key
func (a *APIKeyAuthenticator) RevokeAPIKey(key string) {
	a.mu.Lock()
	delete(a.keys, key)
	a.mu.Unlock()
}

// APIKeyAuthMiddleware creates middleware for API key authentication
func APIKeyAuthMiddleware(auth *APIKeyAuthenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing API key",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		if !auth.ValidateAPIKey(apiKey) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		c.Next()
	}
}

// Authorizer handles role-based authorization
type Authorizer struct {
	roles map[string][]string
	mu    sync.RWMutex
}

// NewAuthorizer creates a new authorizer
func NewAuthorizer() *Authorizer {
	return &Authorizer{
		roles: make(map[string][]string),
	}
}

// AddRole adds a role with permissions
func (a *Authorizer) AddRole(role string, permissions []string) {
	a.mu.Lock()
	a.roles[role] = permissions
	a.mu.Unlock()
}

// HasPermission checks if a role has a specific permission
func (a *Authorizer) HasPermission(role string, permission string) bool {
	a.mu.RLock()
	permissions, exists := a.roles[role]
	a.mu.RUnlock()

	if !exists {
		return false
	}

	for _, perm := range permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// RequirePermission creates middleware that requires a specific permission
func RequirePermission(authz *Authorizer, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get role from context (set by previous auth middleware)
		roleValue, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "No role assigned",
				"code":  "FORBIDDEN",
			})
			return
		}

		role, ok := roleValue.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Invalid role type",
				"code":  "FORBIDDEN",
			})
			return
		}

		if !authz.HasPermission(role, permission) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"code":  "FORBIDDEN",
			})
			return
		}

		c.Next()
	}
}

// CombinedAuthMiddleware creates middleware that accepts either token or API key
func CombinedAuthMiddleware(tokenAuth *TokenAuthenticator, apiKeyAuth *APIKeyAuthenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try token auth first
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			token, ok := ExtractBearerToken(authHeader)
			if ok && tokenAuth.ValidateToken(token) {
				c.Next()
				return
			}
		}

		// Try API key auth
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" && apiKeyAuth.ValidateAPIKey(apiKey) {
			c.Next()
			return
		}

		// Both failed
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
			"code":  "UNAUTHORIZED",
		})
	}
}
