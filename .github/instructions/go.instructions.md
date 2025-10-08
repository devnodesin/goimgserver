---
description: 'Instructions for writing Go code following idiomatic Go practices and community standards'
applyTo: '**/*.go,**/go.mod,**/go.sum'
---

# Go Development Instructions

Follow idiomatic Go practices and community standards when writing Go code. These instructions are based on [Effective Go](https://go.dev/doc/effective_go), [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments), and [Google's Go Style Guide](https://google.github.io/styleguide/go/).

## General Instructions

- Write simple, clear, and idiomatic Go code
- Favor clarity and simplicity over cleverness
- Follow the principle of least surprise
- Keep the happy path left-aligned (minimize indentation)
- Return early to reduce nesting
- Prefer early return over if-else chains; use `if condition { return }` pattern to avoid else blocks
- Make the zero value useful
- Write self-documenting code with clear, descriptive names
- Document exported types, functions, methods, and packages
- Use Go modules for dependency management
- Leverage the Go standard library instead of reinventing the wheel (e.g., use `strings.Builder` for string concatenation, `filepath.Join` for path construction)
- **Use Gin Web Framework**: Leverage Gin's features (routing, middleware, parameter binding, validation, rendering) instead of implementing custom HTTP handling
- Prefer standard library solutions over custom implementations when functionality exists
- **Prefer established frameworks**: Use Gin for HTTP servers, testify for testing assertions, and other well-maintained packages over custom implementations
- Write comments in English by default; translate only upon user request
- Avoid using emoji in code and comments

## Test-Driven Development (TDD) Principles

**MANDATORY**: Follow Test-Driven Development (TDD) methodology for all code implementation:

### TDD Cycle (Red-Green-Refactor)
1. **Red**: Write a failing test first
   - Write the minimal test that defines the desired behavior
   - Ensure the test fails for the right reason
   - Tests must be specific and focused on one behavior
2. **Green**: Write the minimal code to make the test pass
   - Implement only what's needed to make the test pass
   - Don't over-engineer or add unnecessary features
   - Focus on making tests pass quickly
3. **Refactor**: Improve the code while keeping tests green
   - Clean up code structure and design
   - Extract common functionality
   - Ensure all tests continue to pass

### TDD Implementation Rules
- **Always write tests before implementation code**
- **Never write production code without a failing test**
- **Write the smallest possible test that fails**
- **Write the smallest amount of production code to make the test pass**
- **Run tests frequently** (after every small change)
- **Keep test cycles short** (minutes, not hours)
- **One failing test at a time** - fix before moving to the next

### Test Structure and Organization
- **Test file naming**: Use `*_test.go` suffix
- **Test function naming**: `Test_<FunctionName>_<Scenario>` format
- **Table-driven tests**: Use for multiple test cases with same logic
- **Test helpers**: Mark with `t.Helper()` and keep them focused
- **Test fixtures**: Create reusable test data and setup functions
- **Arrange-Act-Assert (AAA)**: Structure tests clearly with setup, execution, and verification phases

## Naming Conventions

### Packages

- Use lowercase, single-word package names
- Avoid underscores, hyphens, or mixedCaps
- Choose names that describe what the package provides, not what it contains
- Avoid generic names like `util`, `common`, or `base`
- Package names should be singular, not plural

#### Package Declaration Rules (CRITICAL):
- **NEVER duplicate `package` declarations** - each Go file must have exactly ONE `package` line
- When editing an existing `.go` file:
  - **PRESERVE** the existing `package` declaration - do not add another one
  - If you need to replace the entire file content, start with the existing package name
- When creating a new `.go` file:
  - **BEFORE writing any code**, check what package name other `.go` files in the same directory use
  - Use the SAME package name as existing files in that directory
  - If it's a new directory, use the directory name as the package name
  - Write **exactly one** `package <name>` line at the very top of the file
- When using file creation or replacement tools:
  - **ALWAYS verify** the target file doesn't already have a `package` declaration before adding one
  - If replacing file content, include only ONE `package` declaration in the new content
  - **NEVER** create files with multiple `package` lines or duplicate declarations

### Variables and Functions

- Use mixedCaps or MixedCaps (camelCase) rather than underscores
- Keep names short but descriptive
- Use single-letter variables only for very short scopes (like loop indices)
- Exported names start with a capital letter
- Unexported names start with a lowercase letter
- Avoid stuttering (e.g., avoid `http.HTTPServer`, prefer `http.Server`)

### Interfaces

- Name interfaces with -er suffix when possible (e.g., `Reader`, `Writer`, `Formatter`)
- Single-method interfaces should be named after the method (e.g., `Read` → `Reader`)
- Keep interfaces small and focused

### Constants

- Use MixedCaps for exported constants
- Use mixedCaps for unexported constants
- Group related constants using `const` blocks
- Consider using typed constants for better type safety

## Code Style and Formatting

### Formatting

- Always use `gofmt` to format code
- Use `goimports` to manage imports automatically
- Keep line length reasonable (no hard limit, but consider readability)
- Add blank lines to separate logical groups of code

### Comments

- Strive for self-documenting code; prefer clear variable names, function names, and code structure over comments
- Write comments only when necessary to explain complex logic, business rules, or non-obvious behavior
- Write comments in complete sentences in English by default
- Translate comments to other languages only upon specific user request
- Start sentences with the name of the thing being described
- Package comments should start with "Package [name]"
- Use line comments (`//`) for most comments
- Use block comments (`/* */`) sparingly, mainly for package documentation
- Document why, not what, unless the what is complex
- Avoid emoji in comments and code

### Error Handling

- Check errors immediately after the function call
- Don't ignore errors using `_` unless you have a good reason (document why)
- Wrap errors with context using `fmt.Errorf` with `%w` verb
- Create custom error types when you need to check for specific errors
- Place error returns as the last return value
- Name error variables `err`
- Keep error messages lowercase and don't end with punctuation

## Architecture and Project Structure

### Package Organization

- Follow standard Go project layout conventions
- Keep `main` packages in `cmd/` directory
- Put reusable packages in `pkg/` or `internal/`
- Use `internal/` for packages that shouldn't be imported by external projects
- Group related functionality into packages
- Avoid circular dependencies

### Dependency Management

- Use Go modules (`go.mod` and `go.sum`)
- **Prioritize Gin Web Framework**: Use `github.com/gin-gonic/gin` for HTTP server functionality
- Keep dependencies minimal and prefer well-established packages
- **Leverage Gin ecosystem**: Use Gin-compatible middleware and extensions when available
- Regularly update dependencies for security patches
- Use `go mod tidy` to clean up unused dependencies
- Vendor dependencies only when necessary

## Type Safety and Language Features

### Type Definitions

- Define types to add meaning and type safety
- Use struct tags for JSON, XML, database mappings
- Prefer explicit type conversions
- Use type assertions carefully and check the second return value
- Prefer generics over unconstrained types; when an unconstrained type is truly needed, use the predeclared alias `any` instead of `interface{}` (Go 1.18+)

### Pointers vs Values

- Use pointer receivers for large structs or when you need to modify the receiver
- Use value receivers for small structs and when immutability is desired
- Use pointer parameters when you need to modify the argument or for large structs
- Use value parameters for small structs and when you want to prevent modification
- Be consistent within a type's method set
- Consider the zero value when choosing pointer vs value receivers

### Interfaces and Composition

- Accept interfaces, return concrete types
- Keep interfaces small (1-3 methods is ideal)
- Use embedding for composition
- Define interfaces close to where they're used, not where they're implemented
- Don't export interfaces unless necessary

## Concurrency

### Goroutines

- Be cautious about creating goroutines in libraries; prefer letting the caller control concurrency
- If you must create goroutines in libraries, provide clear documentation and cleanup mechanisms
- Always know how a goroutine will exit
- Use `sync.WaitGroup` or channels to wait for goroutines
- Avoid goroutine leaks by ensuring cleanup

### Channels

- Use channels to communicate between goroutines
- Don't communicate by sharing memory; share memory by communicating
- Close channels from the sender side, not the receiver
- Use buffered channels when you know the capacity
- Use `select` for non-blocking operations

### Synchronization

- Use `sync.Mutex` for protecting shared state
- Keep critical sections small
- Use `sync.RWMutex` when you have many readers
- Choose between channels and mutexes based on the use case: use channels for communication, mutexes for protecting state
- Use `sync.Once` for one-time initialization
- WaitGroup usage by Go version:
	- If `go >= 1.25` in `go.mod`, use the new `WaitGroup.Go` method ([documentation](https://pkg.go.dev/sync#WaitGroup)):
		```go
		var wg sync.WaitGroup
		wg.Go(task1)
		wg.Go(task2)
		wg.Wait()
		```
	- If `go < 1.25`, use the classic `Add`/`Done` pattern

## Error Handling Patterns

### Creating Errors

- Use `errors.New` for simple static errors
- Use `fmt.Errorf` for dynamic errors
- Create custom error types for domain-specific errors
- Export error variables for sentinel errors
- Use `errors.Is` and `errors.As` for error checking

### Error Propagation

- Add context when propagating errors up the stack
- Don't log and return errors (choose one)
- Handle errors at the appropriate level
- Consider using structured errors for better debugging

## API Design

### Gin Web Framework - https://gin-gonic.com/

**MANDATORY**: Use Gin Web Framework for all HTTP server implementations. Leverage Gin's features and best practices instead of reinventing the wheel.

#### Core Gin Principles
- **Use Gin's router**: `gin.Default()` or `gin.New()` for custom middleware setup
- **Leverage Gin middleware**: Built-in and custom middleware for cross-cutting concerns
- **Use Gin's context**: `*gin.Context` for request/response handling
- **Follow Gin patterns**: Handler functions, middleware chains, route groups
- **Utilize Gin features**: Parameter binding, validation, rendering, error handling

#### Gin Router Setup
```go
// Use gin.Default() for development (includes logging and recovery middleware)
router := gin.Default()

// Use gin.New() for production with custom middleware
router := gin.New()
router.Use(gin.Logger())
router.Use(gin.Recovery())
```

#### Gin Handler Functions
```go
// Proper Gin handler signature
func handleImageRequest(c *gin.Context) {
    // Use Gin's parameter binding
    filename := c.Param("filename")
    
    // Use Gin's query parameter handling
    quality := c.DefaultQuery("quality", "75")
    
    // Use Gin's JSON response
    c.JSON(http.StatusOK, gin.H{
        "filename": filename,
        "quality":  quality,
    })
}

// Register routes using Gin's router
router.GET("/img/:filename", handleImageRequest)
```

#### Gin Middleware Best Practices
```go
// Use built-in middleware when possible
router.Use(gin.Logger())
router.Use(gin.Recovery())

// Custom middleware following Gin patterns
func customMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Pre-processing
        start := time.Now()
        
        // Process request
        c.Next()
        
        // Post-processing
        latency := time.Since(start)
        log.Printf("Request processed in %v", latency)
    }
}
```

#### Gin Route Groups
```go
// Use route groups for organizing endpoints
api := router.Group("/api/v1")
{
    api.GET("/images/:filename", handleImage)
    api.POST("/images/clear", clearCache)
}

// Middleware can be applied to route groups
auth := router.Group("/admin")
auth.Use(authMiddleware())
{
    auth.POST("/cmd/clear", clearAllCache)
    auth.POST("/cmd/gitupdate", gitUpdate)
}
```

#### Gin Parameter Binding and Validation
```go
// Use Gin's ShouldBind for automatic binding
type ImageRequest struct {
    Width   int    `form:"width" binding:"min=10,max=4000"`
    Height  int    `form:"height" binding:"min=10,max=4000"`
    Quality int    `form:"quality" binding:"min=1,max=100"`
    Format  string `form:"format" binding:"oneof=webp png jpeg jpg"`
}

func handleImageProcess(c *gin.Context) {
    var req ImageRequest
    if err := c.ShouldBind(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Process with validated parameters
    // ...
}
```

#### Gin Error Handling
```go
// Use Gin's error handling mechanisms
func handleWithError(c *gin.Context) {
    result, err := processImage()
    if err != nil {
        // Use Gin's AbortWithStatusJSON for errors
        c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
            "error": "Processing failed",
            "code":  "PROCESSING_ERROR",
        })
        return
    }
    
    c.JSON(http.StatusOK, result)
}

// Global error middleware
func errorMiddleware() gin.HandlerFunc {
    return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
        if err, ok := recovered.(string); ok {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Internal server error",
                "code":  "INTERNAL_ERROR",
            })
        }
        c.AbortWithStatus(http.StatusInternalServerError)
    })
}
```

#### Gin Response Rendering
```go
// Use Gin's built-in response methods
c.JSON(http.StatusOK, data)           // JSON response
c.XML(http.StatusOK, data)            // XML response
c.String(http.StatusOK, "Hello")      // String response
c.Data(http.StatusOK, "image/jpeg", imageData) // Binary data
c.File("/path/to/file")               // File response
c.Redirect(http.StatusMovedPermanently, "/new-url") // Redirect
```

#### Gin Testing Patterns
```go
// Test Gin handlers using httptest
func TestImageHandler(t *testing.T) {
    // Setup Gin in test mode
    gin.SetMode(gin.TestMode)
    router := gin.New()
    router.GET("/img/:filename", handleImageRequest)
    
    // Create test request
    req := httptest.NewRequest("GET", "/img/test.jpg", nil)
    w := httptest.NewRecorder()
    
    // Execute request
    router.ServeHTTP(w, req)
    
    // Assert response
    assert.Equal(t, http.StatusOK, w.Code)
}
```

#### Gin Configuration Best Practices
```go
// Production setup
if gin.Mode() == gin.ReleaseMode {
    gin.DisableConsoleColor()
}

// Custom log format
gin.DefaultWriter = io.MultiWriter(os.Stdout, logFile)
router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
    return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
        param.ClientIP,
        param.TimeStamp.Format(time.RFC1123),
        param.Method,
        param.Path,
        param.Request.Proto,
        param.StatusCode,
        param.Latency,
        param.Request.UserAgent(),
        param.ErrorMessage,
    )
}))
```

### HTTP Handlers (Legacy - Use Gin Instead)

⚠️ **DEPRECATED**: Use Gin Web Framework instead of standard `http.HandlerFunc`

- ~~Use `http.HandlerFunc` for simple handlers~~
- ~~Implement `http.Handler` for handlers that need state~~
- ~~Use middleware for cross-cutting concerns~~
- ~~Set appropriate status codes and headers~~
- ~~Handle errors gracefully and return appropriate error responses~~

**Instead**: Use Gin's handler functions, middleware, and context for all HTTP operations.

### JSON APIs

- Use struct tags to control JSON marshaling
- Validate input data
- Use pointers for optional fields
- Consider using `json.RawMessage` for delayed parsing
- Handle JSON errors appropriately

### HTTP Clients

- Keep the client struct focused on configuration and dependencies only (e.g., base URL, `*http.Client`, auth, default headers). It must not store per-request state
- Do not store or cache `*http.Request` inside the client struct, and do not persist request-specific state across calls; instead, construct a fresh request per method invocation
- Methods should accept `context.Context` and input parameters, assemble the `*http.Request` locally (or via a short-lived builder/helper created per call), then call `c.httpClient.Do(req)`
- If request-building logic is reused, factor it into unexported helper functions or a per-call builder type; never keep `http.Request` (URL params, body, headers) as fields on the long-lived client
- Ensure the underlying `*http.Client` is configured (timeouts, transport) and is safe for concurrent use; avoid mutating `Transport` after first use
- Always set headers on the request instance you’re sending, and close response bodies (`defer resp.Body.Close()`), handling errors appropriately

## Performance Optimization

### Memory Management

- Minimize allocations in hot paths
- Reuse objects when possible (consider `sync.Pool`)
- Use value receivers for small structs
- Preallocate slices when size is known
- Avoid unnecessary string conversions

### I/O: Readers and Buffers

- Most `io.Reader` streams are consumable once; reading advances state. Do not assume a reader can be re-read without special handling
- If you must read data multiple times, buffer it once and recreate readers on demand:
	- Use `io.ReadAll` (or a limited read) to obtain `[]byte`, then create fresh readers via `bytes.NewReader(buf)` or `bytes.NewBuffer(buf)` for each reuse
	- For strings, use `strings.NewReader(s)`; you can `Seek(0, io.SeekStart)` on `*bytes.Reader` to rewind
- For HTTP requests, do not reuse a consumed `req.Body`. Instead:
	- Keep the original payload as `[]byte` and set `req.Body = io.NopCloser(bytes.NewReader(buf))` before each send
	- Prefer configuring `req.GetBody` so the transport can recreate the body for redirects/retries: `req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(buf)), nil }`
- To duplicate a stream while reading, use `io.TeeReader` (copy to a buffer while passing through) or write to multiple sinks with `io.MultiWriter`
- Reusing buffered readers: call `(*bufio.Reader).Reset(r)` to attach to a new underlying reader; do not expect it to “rewind” unless the source supports seeking
- For large payloads, avoid unbounded buffering; consider streaming, `io.LimitReader`, or on-disk temporary storage to control memory

- Use `io.Pipe` to stream without buffering the whole payload:
	- Write to `*io.PipeWriter` in a separate goroutine while the reader consumes
	- Always close the writer; use `CloseWithError(err)` on failures
	- `io.Pipe` is for streaming, not rewinding or making readers reusable

- **Warning:** When using `io.Pipe` (especially with multipart writers), all writes must be performed in strict, sequential order. Do not write concurrently or out of order—multipart boundaries and chunk order must be preserved. Out-of-order or parallel writes can corrupt the stream and result in errors.

- Streaming multipart/form-data with `io.Pipe`:
	- `pr, pw := io.Pipe()`; `mw := multipart.NewWriter(pw)`; use `pr` as the HTTP request body
	- Set `Content-Type` to `mw.FormDataContentType()`
	- In a goroutine: write all parts to `mw` in the correct order; on error `pw.CloseWithError(err)`; on success `mw.Close()` then `pw.Close()`
	- Do not store request/in-flight form state on a long-lived client; build per call
	- Streamed bodies are not rewindable; for retries/redirects, buffer small payloads or provide `GetBody`

### Profiling

- Use built-in profiling tools (`pprof`)
- Benchmark critical code paths
- Profile before optimizing
- Focus on algorithmic improvements first
- Consider using `testing.B` for benchmarks

## Testing

**CRITICAL**: All code must be developed using Test-Driven Development (TDD). Tests are not optional - they drive the design and implementation.

### Test Organization

- Keep tests in the same package (white-box testing)
- Use `_test` package suffix for black-box testing
- Name test files with `_test.go` suffix
- Place test files next to the code they test
- **Write tests BEFORE writing implementation code**

### TDD Test Writing Process

1. **Start with a failing test**:
   - Write the test that describes the behavior you want
   - Run the test to confirm it fails for the right reason
   - The test should fail because the functionality doesn't exist yet

2. **Write minimal implementation**:
   - Write only enough code to make the test pass
   - Don't write extra functionality that isn't tested
   - Keep the implementation simple and focused

3. **Refactor with confidence**:
   - Improve the code structure while keeping tests green
   - Tests provide safety net for refactoring
   - Clean up both test and production code

### Test Writing Guidelines

- **Use table-driven tests** for multiple test cases with similar logic
- **Name tests descriptively** using `Test_functionName_scenario` format
- **Use subtests** with `t.Run` for better organization and parallel execution
- **Test both success and error cases** - error cases are especially important
- **Test edge cases and boundary conditions** (empty inputs, nil values, large datasets)
- **One assertion per test** when possible for clarity
- **Arrange-Act-Assert structure**: Setup, execute, verify in clear sections
- **Use testify for assertions**: `github.com/stretchr/testify/assert` for cleaner test code
- **Test Gin handlers**: Use `gin.SetMode(gin.TestMode)` and `httptest` for HTTP handler testing
- Consider using `testify` or similar libraries when they add value, but don't over-complicate simple tests

### Test-First Examples

```go
// Step 1: Write failing test first for Gin handler
func TestImageHandler_GET_ValidRequest(t *testing.T) {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    
    // This will fail because handleImageRequest doesn't exist yet
    router.GET("/img/:filename", handleImageRequest)
    
    req := httptest.NewRequest("GET", "/img/test.jpg", nil)
    w := httptest.NewRecorder()
    
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    // Test will fail because handler doesn't exist yet
}

// Step 2: Write minimal Gin handler implementation to make test pass
func handleImageRequest(c *gin.Context) {
    // Minimal implementation - just return OK to make test pass
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Step 3: Add more specific tests and refactor implementation
func TestImageHandler_GET_ReturnsImageData(t *testing.T) {
    // More specific test that will drive better implementation
    gin.SetMode(gin.TestMode)
    router := gin.New()
    router.GET("/img/:filename", handleImageRequest)
    
    req := httptest.NewRequest("GET", "/img/test.jpg", nil)
    w := httptest.NewRecorder()
    
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Header().Get("Content-Type"), "image/")
}
```

#### Gin-Specific TDD Patterns
```go
// Test Gin parameter binding
func TestImageHandler_ParameterBinding(t *testing.T) {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    router.GET("/img/:filename/:width/:height", handleImageResize)
    
    req := httptest.NewRequest("GET", "/img/test.jpg/800/600", nil)
    w := httptest.NewRecorder()
    
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
}

// Test Gin middleware
func TestRateLimitMiddleware(t *testing.T) {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    router.Use(rateLimitMiddleware())
    router.GET("/test", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })
    
    // Test multiple requests to trigger rate limit
    for i := 0; i < 10; i++ {
        req := httptest.NewRequest("GET", "/test", nil)
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        if i < 5 {
            assert.Equal(t, http.StatusOK, w.Code)
        } else {
            assert.Equal(t, http.StatusTooManyRequests, w.Code)
        }
    }
}
```

### Test Helpers

- Mark helper functions with `t.Helper()`
- Create test fixtures for complex setup
- Use `testing.TB` interface for functions used in tests and benchmarks
- Clean up resources using `t.Cleanup()`

## Security Best Practices

### Input Validation

- Validate all external input
- Use strong typing to prevent invalid states
- Sanitize data before using in SQL queries
- Be careful with file paths from user input
- Validate and escape data for different contexts (HTML, SQL, shell)

### Cryptography

- Use standard library crypto packages
- Don't implement your own cryptography
- Use crypto/rand for random number generation
- Store passwords using bcrypt, scrypt, or argon2 (consider golang.org/x/crypto for additional options)
- Use TLS for network communication

## Documentation

### Code Documentation

- Prioritize self-documenting code through clear naming and structure
- Document all exported symbols with clear, concise explanations
- Start documentation with the symbol name
- Write documentation in English by default
- Use examples in documentation when helpful
- Keep documentation close to code
- Update documentation when code changes
- Avoid emoji in documentation and comments

### README and Documentation Files

- Include clear setup instructions
- Document dependencies and requirements
- Provide usage examples
- Document configuration options
- Include troubleshooting section

## Tools and Development Workflow

### Essential Tools

- `go fmt`: Format code
- `go vet`: Find suspicious constructs
- `golangci-lint`: Additional linting (golint is deprecated)
- `go test`: Run tests
- `go mod`: Manage dependencies
- `go generate`: Code generation
- **Gin Web Framework**: `github.com/gin-gonic/gin` for HTTP server development

### Development Practices

- Run tests before committing
- Use pre-commit hooks for formatting and linting
- Keep commits focused and atomic
- Write meaningful commit messages
- Review diffs before committing

## Common Pitfalls to Avoid

- Not checking errors
- Ignoring race conditions
- Creating goroutine leaks
- Not using defer for cleanup
- Modifying maps concurrently
- Not understanding nil interfaces vs nil pointers
- Forgetting to close resources (files, connections)
- Using global variables unnecessarily
- Over-using unconstrained types (e.g., `any`); prefer specific types or generic type parameters with constraints. If an unconstrained type is required, use `any` rather than `interface{}`
- Not considering the zero value of types
- **Creating duplicate `package` declarations** - this is a compile error; always check existing files before adding package declarations
- **Reinventing HTTP handling** - use Gin Web Framework instead of custom HTTP routers, middleware, or parameter parsing
- **Not leveraging Gin features** - use Gin's built-in parameter binding, validation, middleware, and rendering instead of custom implementations
