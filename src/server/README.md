# Server Package

Enhanced HTTP server with comprehensive middleware, health checks, and graceful shutdown for goimgserver.

## Features

- **CORS Support**: Configurable cross-origin resource sharing
- **Security Headers**: Automatic security headers (X-Content-Type-Options, X-Frame-Options, etc.)
- **Request ID**: Automatic request ID generation and propagation for tracing
- **Structured Logging**: Request/response logging with correlation IDs
- **Error Handling**: Panic recovery and standardized error responses
- **Rate Limiting**: Configurable global or per-IP rate limiting
- **Health Checks**: Liveness, readiness, and detailed health endpoints
- **Graceful Shutdown**: Proper connection draining on shutdown

## Quick Start

```go
package main

import (
    "goimgserver/server"
    "time"
)

func main() {
    // Configure server
    config := &server.Config{
        Port:            8080,
        ReadTimeout:     30 * time.Second,
        WriteTimeout:    30 * time.Second,
        ShutdownTimeout: 10 * time.Second,
        EnableCORS:      true,
        EnableRateLimit: true,
        RateLimit:       100,
        RatePer:         time.Minute,
        Production:      false,
    }
    
    // Create server
    srv := server.New(config)
    
    // Add health checks
    srv.AddHealthCheck("database", func() bool {
        return true // Check database connectivity
    })
    
    // Register routes
    srv.Router.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello!"})
    })
    
    // Run with graceful shutdown
    srv.Run()
}
```

## Configuration

### Server.Config

```go
type Config struct {
    Port            int           // Server port
    ReadTimeout     time.Duration // Read timeout
    WriteTimeout    time.Duration // Write timeout
    ShutdownTimeout time.Duration // Graceful shutdown timeout
    EnableCORS      bool          // Enable CORS middleware
    EnableRateLimit bool          // Enable rate limiting
    RateLimit       int           // Number of requests
    RatePer         time.Duration // Per time period
    Production      bool          // Production mode (disables debug logs)
}
```

## Middleware

The server automatically applies the following middleware in order:

1. **Request ID** - Generates unique request IDs
2. **Security Headers** - Adds security headers
3. **CORS** - Handles cross-origin requests (if enabled)
4. **Error Handler** - Catches panics and formats errors
5. **Logging** - Logs requests and responses
6. **Rate Limiter** - Limits request rate (if enabled)

## Health Endpoints

### GET /health
Detailed health check with dependency status.

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2025-01-15T10:30:00Z",
  "uptime": "1h30m0s",
  "checks": {
    "database": "ok",
    "cache": "ok"
  }
}
```

### GET /live
Liveness probe for Kubernetes.

**Response:**
```json
{
  "status": "alive"
}
```

### GET /ready
Readiness probe for Kubernetes.

**Response:**
```json
{
  "status": "ready"
}
```

## Middleware Details

### CORS Middleware

```go
// Use default CORS (allows all origins)
srv.Router.Use(middleware.CORS())

// Use specific origins
srv.Router.Use(middleware.CORSWithOrigins([]string{
    "https://example.com",
    "https://app.example.com",
}))
```

### Rate Limiting

```go
// Global rate limit
srv.Router.Use(middleware.RateLimit(100, time.Minute))

// Per-IP rate limit
srv.Router.Use(middleware.RateLimitPerIP(10, time.Second))

// With burst capacity
srv.Router.Use(middleware.RateLimitWithBurst(50, time.Minute, 100))
```

### Request ID

Request IDs are automatically generated and added to:
- Response header: `X-Request-ID`
- Context: `c.GetString("request_id")`
- Logs: `[request-id] ...`

### Security Headers

Automatically adds:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Referrer-Policy: strict-origin-when-cross-origin`

### Error Handling

Automatically catches panics and returns standardized error responses:

```json
{
  "error": "Internal server error",
  "code": "INTERNAL_ERROR",
  "request_id": "abc123..."
}
```

## Graceful Shutdown

The server handles SIGINT and SIGTERM signals for graceful shutdown:

1. Stops accepting new connections
2. Waits for active requests to complete
3. Shuts down within the configured timeout
4. Logs shutdown status

```go
// Manual shutdown
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
srv.Shutdown(ctx)
```

## Testing

The package includes comprehensive tests:

```bash
# Run all tests
go test ./server/...

# Run with coverage
go test ./server/... -cover

# Run benchmarks
go test ./server/middleware/... -bench=.
```

### Test Coverage

- Server: 65.1%
- Health: 90.6%
- Middleware: 92.9%
- **Overall: 86.6%**

## Examples

See `server/example/main.go` for a complete example.

Run the example:
```bash
cd src/server/example
go run main.go
```

## Architecture

```
server/
├── server.go           # Main server implementation
├── server_test.go      # Server integration tests
├── middleware/
│   ├── cors.go         # CORS middleware
│   ├── cors_test.go
│   ├── logging.go      # Logging middleware
│   ├── logging_test.go
│   ├── error.go        # Error handling middleware
│   ├── error_test.go
│   ├── ratelimit.go    # Rate limiting middleware
│   ├── ratelimit_test.go
│   ├── requestid.go    # Request ID middleware
│   ├── requestid_test.go
│   ├── security.go     # Security headers middleware
│   ├── security_test.go
│   ├── integration_test.go
│   └── benchmark_test.go
└── health/
    ├── checker.go      # Health check implementation
    └── checker_test.go
```

## Best Practices

1. **Always use graceful shutdown** in production
2. **Add health checks** for all critical dependencies
3. **Enable rate limiting** to prevent abuse
4. **Use production mode** in production deployments
5. **Configure appropriate timeouts** based on your use case
6. **Monitor health endpoints** for alerting

## Performance

Benchmark results show minimal overhead:

```
BenchmarkMiddleware_CORS_Overhead-4              	 1000000	      1168 ns/op
BenchmarkMiddleware_Logging_Overhead-4           	  500000	      2871 ns/op
BenchmarkMiddleware_FullStack-4                  	  151726	      9913 ns/op
BenchmarkServer_RequestThroughput-4              	  549658	      2751 ns/op
```

The full middleware stack adds approximately 4-5x overhead compared to a bare router, which is acceptable for production use.

## Integration with goimgserver

This server package is designed to replace the basic Gin setup in the main goimgserver application, providing production-ready middleware and operational features without changing the existing handlers.

The integration maintains backward compatibility with existing image and command handlers while adding enterprise-grade features like health checks, structured logging, and graceful shutdown.
