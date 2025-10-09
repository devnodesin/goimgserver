# Command Endpoints Quick Reference

## Overview
Administrative command endpoints for cache management and Git operations.

## Endpoints

### Clear Cache
**Endpoint:** `POST /cmd/clear`  
**Description:** Clears the entire cache directory  
**Authentication:** None (should be added in production)

**Request:**
```bash
curl -X POST http://localhost:9000/cmd/clear
```

**Response (Success):**
```json
{
  "success": true,
  "message": "Cache cleared successfully",
  "cleared_files": 1234,
  "freed_space": "2.5GB"
}
```

**Response (Error):**
```json
{
  "success": false,
  "error": "Failed to clear cache"
}
```

---

### Git Update
**Endpoint:** `POST /cmd/gitupdate`  
**Description:** Updates images directory via git pull  
**Requirements:** Images directory must be a git repository  
**Authentication:** None (should be added in production)

**Request:**
```bash
curl -X POST http://localhost:9000/cmd/gitupdate
```

**Response (Success):**
```json
{
  "success": true,
  "message": "Git update completed",
  "changes": 5,
  "branch": "main",
  "last_commit": "abc123..."
}
```

**Response (Not a Git Repo):**
```json
{
  "success": false,
  "error": "Images directory is not a git repository",
  "code": "GIT_NOT_FOUND"
}
```

**Response (Git Error):**
```json
{
  "success": false,
  "error": "git update failed: [details]",
  "code": "GIT_UPDATE_FAILED"
}
```

---

### Generic Command
**Endpoint:** `POST /cmd/:name`  
**Description:** Generic command router  
**Valid Commands:** `clear`, `gitupdate`  
**Authentication:** None (should be added in production)

**Request:**
```bash
# Valid command
curl -X POST http://localhost:9000/cmd/clear

# Invalid command
curl -X POST http://localhost:9000/cmd/invalid
```

**Response (Invalid Command):**
```json
{
  "success": false,
  "error": "invalid command: invalid",
  "code": "INVALID_COMMAND"
}
```

## Security Considerations

### Current Implementation
- ✅ Command validation (whitelist only)
- ✅ Path traversal prevention
- ✅ Command injection prevention
- ✅ Timeout protection (30 seconds)
- ✅ Clean environment variables

### Production Recommendations
- ⚠️  Add authentication (API keys or tokens)
- ⚠️  Add authorization (role-based access)
- ⚠️  Add rate limiting
- ⚠️  Add audit logging
- ⚠️  Restrict to admin network/IPs

## Usage Examples

### Shell Script
```bash
#!/bin/bash

# Clear cache
response=$(curl -s -X POST http://localhost:9000/cmd/clear)
echo "Clear cache: $response"

# Update from git
response=$(curl -s -X POST http://localhost:9000/cmd/gitupdate)
echo "Git update: $response"
```

### Python
```python
import requests

# Clear cache
response = requests.post('http://localhost:9000/cmd/clear')
print(f"Clear cache: {response.json()}")

# Update from git
response = requests.post('http://localhost:9000/cmd/gitupdate')
print(f"Git update: {response.json()}")
```

### JavaScript/Node.js
```javascript
const fetch = require('node-fetch');

// Clear cache
fetch('http://localhost:9000/cmd/clear', { method: 'POST' })
  .then(res => res.json())
  .then(data => console.log('Clear cache:', data));

// Update from git
fetch('http://localhost:9000/cmd/gitupdate', { method: 'POST' })
  .then(res => res.json())
  .then(data => console.log('Git update:', data));
```

## Error Codes

| Code | Description | HTTP Status |
|------|-------------|-------------|
| `INVALID_COMMAND` | Command name not in whitelist | 400 |
| `GIT_NOT_FOUND` | Directory is not a git repository | 400 |
| `GIT_UPDATE_FAILED` | Git pull operation failed | 500 |
| `INVALID_PATH` | Path contains invalid characters | 400 |

## Testing

Test the endpoints with the provided script:
```bash
./test_commands.sh
```

Or manually:
```bash
# Test clear cache
curl -v -X POST http://localhost:9000/cmd/clear

# Test git update
curl -v -X POST http://localhost:9000/cmd/gitupdate

# Test invalid command
curl -v -X POST http://localhost:9000/cmd/invalid
```

## Monitoring

Recommended monitoring points:
- Command execution frequency
- Command success/failure rates
- Git operation duration
- Cache clear frequency
- Error rates by error code

## Further Reading

- [Full Implementation Documentation](COMMAND_IMPLEMENTATION.md)
- [Security Documentation](design/07-security.md)
- [API Specification](design/06-api-specification.md)
