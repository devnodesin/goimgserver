# Command Endpoints Architecture

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                          goimgserver                            │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │                      HTTP Server (Gin)                     │ │
│  │                                                            │ │
│  │  ┌─────────────────┐        ┌────────────────────────┐   │ │
│  │  │ Image Endpoints │        │  Command Endpoints      │   │ │
│  │  │                 │        │                         │   │ │
│  │  │ GET /img/*path  │        │ POST /cmd/clear         │   │ │
│  │  │                 │        │ POST /cmd/gitupdate     │   │ │
│  │  │                 │        │ POST /cmd/:name         │   │ │
│  │  └────────┬────────┘        └──────────┬──────────────┘   │ │
│  └───────────┼────────────────────────────┼──────────────────┘ │
│              │                             │                    │
│              │                             │                    │
│     ┌────────▼────────┐          ┌────────▼────────┐          │
│     │  ImageHandler   │          │ CommandHandler  │          │
│     │                 │          │                 │          │
│     │ - ServeImage()  │          │ - HandleClear() │          │
│     │                 │          │ - HandleGitUp() │          │
│     │                 │          │ - HandleCmd()   │          │
│     └────────┬────────┘          └────────┬────────┘          │
│              │                             │                    │
│              │                    ┌────────┼────────┐          │
│              │                    │        │        │          │
│              │                    │   ┌────▼─────┐ │          │
│              │                    │   │ GitOps   │ │          │
│     ┌────────▼────────┐           │   │          │ │          │
│     │  CacheManager   │◄──────────┤   │ - IsRepo │ │          │
│     │                 │           │   │ - GitPull│ │          │
│     │ - Store()       │           │   │ - Validate│           │
│     │ - Retrieve()    │           │   └──────────┘ │          │
│     │ - Clear()       │           │                │          │
│     │ - ClearAll()    │           └────────────────┘          │
│     │ - GetStats()    │                                        │
│     └─────────────────┘                                        │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │                    File System                           │  │
│  │                                                          │  │
│  │  ┌───────────────┐              ┌──────────────────┐   │  │
│  │  │  Images Dir   │              │   Cache Dir       │   │  │
│  │  │               │              │                   │   │  │
│  │  │ (May be Git)  │              │  (Auto-managed)   │   │  │
│  │  └───────────────┘              └──────────────────┘   │  │
│  └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Component Responsibilities

### CommandHandler
- **Purpose**: Handle administrative command endpoints
- **Dependencies**: Config, CacheManager, GitOperations
- **Methods**:
  - `HandleClear()` - Clear all cached files
  - `HandleGitUpdate()` - Update images from Git
  - `HandleCommand()` - Route generic commands
- **Security**: Input validation, command whitelisting

### GitOperations
- **Purpose**: Manage Git repository operations
- **Methods**:
  - `IsGitRepo()` - Detect Git repositories
  - `ExecuteGitPull()` - Execute git pull with timeout
  - `ValidatePath()` - Prevent path traversal
- **Security**: Clean environment, timeout protection

### CacheManager
- **Purpose**: Manage cached processed images
- **Methods**:
  - `Store()` - Cache processed images
  - `Retrieve()` - Get cached images
  - `Clear()` - Clear specific file cache
  - `ClearAll()` - Clear entire cache
  - `GetStats()` - Get cache statistics
- **Features**: Atomic operations, thread-safe

## Request Flow

### POST /cmd/clear
```
Client Request
    │
    ├──> Gin Router
    │
    ├──> CommandHandler.HandleClear()
    │
    ├──> CacheManager.GetStats()     (count files)
    │
    ├──> CacheManager.ClearAll()     (delete all)
    │
    └──> JSON Response
         {
           "success": true,
           "cleared_files": 1234,
           "freed_space": "2.5GB"
         }
```

### POST /cmd/gitupdate
```
Client Request
    │
    ├──> Gin Router
    │
    ├──> CommandHandler.HandleGitUpdate()
    │
    ├──> Security Validation
    │    (Check for shell metacharacters)
    │
    ├──> GitOperations.IsGitRepo()
    │    (Verify .git directory exists)
    │
    ├──> GitOperations.ExecuteGitPull()
    │    (Run git pull with timeout)
    │
    └──> JSON Response
         {
           "success": true,
           "changes": 5,
           "branch": "main",
           "last_commit": "abc123"
         }
```

## Security Layers

```
┌─────────────────────────────────────────────┐
│         Security Layer 1: Input             │
│   ┌─────────────────────────────────────┐  │
│   │ - Command whitelist validation      │  │
│   │ - Shell metacharacter detection     │  │
│   │ - Path normalization                │  │
│   └─────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────┐
│        Security Layer 2: Execution          │
│   ┌─────────────────────────────────────┐  │
│   │ - Path traversal prevention         │  │
│   │ - Clean environment variables       │  │
│   │ - Explicit command arguments        │  │
│   │ - No shell interpolation            │  │
│   └─────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────┐
│      Security Layer 3: Resource Control     │
│   ┌─────────────────────────────────────┐  │
│   │ - 30-second timeout                 │  │
│   │ - Context cancellation              │  │
│   │ - Error sanitization                │  │
│   └─────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

## Data Flow

### Cache Clearing
```
HandleClear()
    │
    ├──> GetStats() ──────────────┐
    │                              │
    │    File Count: 1234          │
    │    Total Size: 2.5GB         │
    │                              │
    ├──> ClearAll() ───────────────┤
    │                              │
    │    Delete all entries ───────┤
    │                              │
    └──> JSON Response ◄───────────┘
```

### Git Update
```
HandleGitUpdate()
    │
    ├──> Validate Path ───────────┐
    │                              │
    ├──> IsGitRepo() ──────────────┤
    │                              │
    │    Check .git directory ─────┤
    │                              │
    ├──> ExecuteGitPull() ─────────┤
    │                              │
    │    git pull with timeout ────┤
    │                              │
    ├──> Parse Results ────────────┤
    │                              │
    │    Branch: main              │
    │    Commit: abc123            │
    │    Changes: 5                │
    │                              │
    └──> JSON Response ◄───────────┘
```

## Error Handling Flow

```
Error Occurs
    │
    ├──> Log Detailed Error
    │    (Internal logs only)
    │
    ├──> Generate Generic Message
    │    (User-facing)
    │
    ├──> Set HTTP Status Code
    │    - 400 Bad Request
    │    - 500 Internal Server Error
    │
    └──> Return JSON Error
         {
           "success": false,
           "error": "Generic message",
           "code": "ERROR_CODE"
         }
```

## Testing Architecture

```
┌────────────────────────────────────────────┐
│             Test Pyramid                    │
│                                             │
│              ┌──────────┐                   │
│              │Integration│ (3 tests)        │
│              └─────┬─────┘                  │
│                    │                        │
│           ┌────────┴────────┐               │
│           │   Unit Tests    │ (18 tests)    │
│           └────────┬────────┘               │
│                    │                        │
│         ┌──────────┴──────────┐             │
│         │   Mock-Based Tests   │            │
│         └─────────────────────┘             │
│                                             │
│  ┌──────────────────────────────────────┐  │
│  │    Git Operations (10 tests)         │  │
│  │    - IsGitRepo: 3 tests              │  │
│  │    - ExecuteGitPull: 4 tests         │  │
│  │    - ValidatePath: 3 tests           │  │
│  └──────────────────────────────────────┘  │
│                                             │
│  ┌──────────────────────────────────────┐  │
│  │    Command Handlers (11 tests)       │  │
│  │    - Cache Clear: 2 tests            │  │
│  │    - Git Update: 3 tests             │  │
│  │    - Generic Command: 2 tests        │  │
│  │    - Security: 3 tests               │  │
│  │    - Integration: 1 test             │  │
│  └──────────────────────────────────────┘  │
│                                             │
│  ┌──────────────────────────────────────┐  │
│  │    Performance (1 benchmark)          │  │
│  │    - BenchmarkCommandHandler_Cache    │  │
│  └──────────────────────────────────────┘  │
└────────────────────────────────────────────┘
```

## Deployment Architecture

```
┌─────────────────────────────────────────────────┐
│                  Production                      │
│                                                  │
│  ┌────────────┐      ┌──────────────────────┐  │
│  │ Reverse    │      │   goimgserver         │  │
│  │ Proxy      │─────>│                       │  │
│  │ (nginx)    │      │   Port: 9000          │  │
│  └────────────┘      │                       │  │
│       │              │   Command Endpoints:  │  │
│       │              │   - Authenticated     │  │
│  ┌────▼──────┐      │   - Rate Limited      │  │
│  │ Firewall  │      │   - Audit Logged      │  │
│  │           │      └──────────────────────────┘  │
│  │ Rules:    │                                  │
│  │ - Allow admin IP only                       │
│  │ - Rate limit commands                       │
│  └───────────┘                                  │
└─────────────────────────────────────────────────┘
```

## Future Enhancements

### Phase 1: Authentication
```
Add API Key Middleware
    │
    ├──> Validate API Key
    ├──> Check Permissions
    └──> Allow/Deny Request
```

### Phase 2: Monitoring
```
Command Execution
    │
    ├──> Record Metrics
    │    - Execution time
    │    - Success/failure
    │    - User/IP
    │
    └──> Update Dashboard
```

### Phase 3: Advanced Features
```
- Scheduled cache clearing
- Selective cache clearing
- Git webhook integration
- Multi-repository support
```

## Integration Points

### With Cache System
- Uses `CacheManager.ClearAll()` for clearing
- Uses `CacheManager.GetStats()` for statistics
- Thread-safe operations

### With Git System
- Detects Git repositories
- Executes git pull operations
- Validates paths
- Handles timeouts

### With HTTP Server (Gin)
- Registers POST endpoints
- Handles JSON requests/responses
- Provides middleware support
- Error handling integration
