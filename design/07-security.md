# Security

## Security Architecture

goimgserver implements multiple layers of security to protect against common web application vulnerabilities while maintaining performance and usability.

## Threat Model

### Primary Threats
- **Path Traversal**: Malicious requests attempting to access files outside {image_dir}
- **Resource Exhaustion**: Large image processing requests that could exhaust system resources
- **Cache Poisoning**: Attempts to pollute the cache with malicious content
- **Command Injection**: Malicious input to administrative endpoints
- **File System Attacks**: Attempts to access sensitive system files
- **Denial of Service**: High-volume requests to overwhelm the service

### Secondary Threats
- **Information Disclosure**: Revealing system information through error messages
- **Unauthorized Cache Access**: Accessing cached images without proper authorization
- **Git Repository Exposure**: Exposing Git repository information through gitupdate command

## Input Validation and Sanitization

### Path Traversal Prevention

#### URL Path Sanitization
```go
func sanitizePath(path string) string {
    // Clean the path to resolve . and .. elements
    cleaned := filepath.Clean(path)
    
    // Reject paths containing .. after cleaning
    if strings.Contains(cleaned, "..") {
        return ""
    }
    
    // Ensure path doesn't start with /
    cleaned = strings.TrimPrefix(cleaned, "/")
    
    // Additional security checks
    if strings.Contains(cleaned, "\x00") { // Null byte injection
        return ""
    }
    
    return cleaned
}
```

#### File Path Validation
- All file paths are validated to be within {image_dir}
- Symlinks are followed only if they point to files within {image_dir}
- Absolute paths are rejected
- Path components are validated against allowlist patterns

### Parameter Validation

#### Dimension Limits
```go
const (
    MinDimension = 10
    MaxDimension = 4000
    DefaultWidth = 1000
    DefaultHeight = 1000
)

func validateDimensions(width, height int) (int, int, error) {
    if width < MinDimension || width > MaxDimension {
        width = DefaultWidth
    }
    if height < MinDimension || height > MaxDimension {
        height = DefaultHeight
    }
    return width, height, nil
}
```

#### Quality Parameter Limits
```go
const (
    MinQuality = 1
    MaxQuality = 100
    DefaultQuality = 75
)

func validateQuality(quality int) int {
    if quality < MinQuality || quality > MaxQuality {
        return DefaultQuality
    }
    return quality
}
```

#### Format Validation
```go
var allowedFormats = map[string]bool{
    "webp": true,
    "png":  true,
    "jpeg": true,
    "jpg":  true,
}

func validateFormat(format string) string {
    if !allowedFormats[strings.ToLower(format)] {
        return "webp" // Default format
    }
    return strings.ToLower(format)
}
```

## Resource Protection

### Memory Limits
- Maximum image file size: 100MB
- Maximum processing memory: 512MB per request
- Garbage collection optimization for image processing

### Processing Limits
- Maximum concurrent image processing operations: 10
- Request timeout: 30 seconds
- Processing timeout: 60 seconds

### Disk Space Protection
- Cache size limits with automatic cleanup
- Temporary file cleanup after processing
- Disk space monitoring and alerts

## Cache Security

### Cache Isolation
- Cache files are stored with restrictive permissions (600)
- Cache directories are created with appropriate permissions (700)
- Cache file names are hashed to prevent direct access

### Cache Key Security
```go
func generateCacheKey(filename string, params ProcessingParams) string {
    h := sha256.New()
    h.Write([]byte(filename))
    h.Write([]byte(fmt.Sprintf("%dx%d", params.Width, params.Height)))
    h.Write([]byte(params.Format))
    h.Write([]byte(fmt.Sprintf("q%d", params.Quality)))
    return hex.EncodeToString(h.Sum(nil))
}
```

### Cache Access Control
- Cache files are not directly accessible via HTTP
- Cache cleanup operations are logged and monitored
- Cache directory permissions prevent unauthorized access

## File System Security

### Directory Permissions
```bash
{image_dir}/: 755 (read/execute for owner, group, and others)
{cache_dir}/: 700 (read/write/execute for owner only)
```

### File Permissions
```bash
Image files: 644 (read for owner, group, and others)
Cache files: 600 (read/write for owner only)
```

### Symlink Handling
```go
func resolveSymlink(path string, baseDir string) (string, error) {
    resolved, err := filepath.EvalSymlinks(path)
    if err != nil {
        return "", err
    }
    
    // Ensure resolved path is within baseDir
    if !strings.HasPrefix(resolved, baseDir) {
        return "", errors.New("symlink points outside allowed directory")
    }
    
    return resolved, nil
}
```

## HTTP Security

### Security Headers
```go
func setSecurityHeaders(w http.ResponseWriter) {
    w.Header().Set("X-Content-Type-Options", "nosniff")
    w.Header().Set("X-Frame-Options", "DENY")
    w.Header().Set("X-XSS-Protection", "1; mode=block")
    w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
}
```

### CORS Configuration
```go
func setupCORS(router *gin.Engine) {
    config := cors.Config{
        AllowOrigins:     []string{"*"}, // Configure appropriately for production
        AllowMethods:     []string{"GET", "POST", "OPTIONS"},
        AllowHeaders:     []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization"},
        ExposeHeaders:    []string{"Content-Length", "X-Processing-Time"},
        AllowCredentials: false,
        MaxAge:           12 * time.Hour,
    }
    router.Use(cors.New(config))
}
```

### Request Size Limits
```go
func setupRequestLimits(router *gin.Engine) {
    // Limit request body size
    router.Use(gin.BodyLimit(1 * 1024 * 1024)) // 1MB limit
    
    // Limit URL length
    router.Use(func(c *gin.Context) {
        if len(c.Request.URL.String()) > 2048 {
            c.AbortWithStatus(http.StatusRequestURITooLong)
            return
        }
        c.Next()
    })
}
```

## Error Handling Security

### Secure Error Messages
```go
func handleError(c *gin.Context, err error, userMessage string) {
    // Log detailed error for debugging
    log.Error().Err(err).Msg("Internal error occurred")
    
    // Return generic error to user
    c.JSON(http.StatusInternalServerError, gin.H{
        "error": userMessage,
        "code":  "INTERNAL_ERROR",
    })
}
```

### Information Disclosure Prevention
- Stack traces are not exposed to clients
- File system paths are not revealed in error messages
- System information is not disclosed in responses
- Error responses use generic messages

## Administrative Security

### Command Endpoint Protection
```go
func validateCommand(command string) bool {
    allowedCommands := map[string]bool{
        "clear":     true,
        "gitupdate": true,
    }
    return allowedCommands[command]
}
```

### Git Operation Security
```go
func executeGitUpdate(dir string) error {
    // Validate directory is within allowed path
    if !isWithinAllowedPath(dir) {
        return errors.New("invalid directory path")
    }
    
    // Use specific git commands with limited permissions
    cmd := exec.Command("git", "pull", "origin", "main")
    cmd.Dir = dir
    cmd.Env = []string{} // Clean environment
    
    // Set timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    return cmd.Run()
}
```

### Cache Clear Security
```go
func clearCache(cacheDir string) error {
    // Validate cache directory path
    if !strings.HasPrefix(cacheDir, allowedCacheDir) {
        return errors.New("invalid cache directory")
    }
    
    // Use safe file removal
    return os.RemoveAll(cacheDir)
}
```

## Rate Limiting

### Implementation
```go
func setupRateLimit(router *gin.Engine) {
    limiter := rate.NewLimiter(rate.Every(time.Minute), 100) // 100 requests per minute
    
    router.Use(func(c *gin.Context) {
        if !limiter.Allow() {
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
                "retry_after": "60s",
            })
            return
        }
        c.Next()
    })
}
```

### Adaptive Rate Limiting
- Different limits for image processing vs. cache operations
- IP-based rate limiting with allowlists
- Exponential backoff for repeated violations

## Monitoring and Logging

### Security Event Logging
```go
func logSecurityEvent(event string, details map[string]interface{}) {
    log.Warn().
        Str("event", event).
        Interface("details", details).
        Str("timestamp", time.Now().UTC().Format(time.RFC3339)).
        Msg("Security event detected")
}
```

### Audit Trail
- All administrative operations are logged
- Failed authentication attempts are recorded
- Unusual access patterns are flagged
- File system access is monitored

### Alerting
- Real-time alerts for security events
- Integration with monitoring systems (Prometheus, Grafana)
- Automated response to certain threats

## Deployment Security

### Container Security
```dockerfile
# Use non-root user
RUN adduser --disabled-password --gecos '' appuser
USER appuser

# Set secure permissions
COPY --chown=appuser:appuser ./app /app
RUN chmod 755 /app

# Remove unnecessary packages
RUN apt-get remove --purge -y build-essential && \
    apt-get autoremove -y && \
    apt-get clean
```

### Network Security
- Bind to localhost by default
- Use reverse proxy (nginx) for production
- TLS termination at proxy level
- Firewall rules restricting access

### Process Security
```bash
# Run with limited privileges
sudo -u goimgserver ./goimgserver --port 9000

# Use systemd for process management
[Unit]
Description=Go Image Server
After=network.target

[Service]
Type=simple
User=goimgserver
Group=goimgserver
ExecStart=/usr/local/bin/goimgserver
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## Security Testing

### Automated Security Tests
- Path traversal attack tests
- Resource exhaustion tests
- Input validation bypass tests
- Cache poisoning tests

### Manual Security Review
- Regular code review for security issues
- Penetration testing of deployed instances
- Security audit of dependencies

### Vulnerability Management
- Regular dependency updates
- Security patch management
- Vulnerability disclosure process

## Production Security Checklist

### Before Deployment
- [ ] Review and configure CORS origins
- [ ] Set up proper file permissions
- [ ] Configure rate limiting
- [ ] Set up monitoring and alerting
- [ ] Review error message disclosure
- [ ] Test security controls

### Ongoing Maintenance
- [ ] Regular security updates
- [ ] Log review and analysis
- [ ] Performance monitoring
- [ ] Access control review
- [ ] Backup and recovery testing

### Incident Response
- [ ] Security incident response plan
- [ ] Contact information for security team
- [ ] Procedures for emergency shutdown
- [ ] Recovery procedures
- [ ] Communication protocols

This security framework provides comprehensive protection while maintaining the performance and usability requirements of goimgserver.