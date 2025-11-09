# Session Management

This document provides a deep dive into session management in `req`.

## Overview

Sessions allow you to authenticate once and have credentials automatically applied to subsequent requests to the same host. This eliminates the need to include authentication headers or cookies in every request.

## How Sessions Work

### 1. Authentication Flow

When you use the `authenticate` verb, `req`:

1. Executes the authentication request
2. Follows redirects (to capture Set-Cookie headers from redirect responses)
3. Captures all `Set-Cookie` headers
4. Extracts `access_token` from JSON response body (if present)
5. Stores session data per host

```mermaid
sequenceDiagram
    participant User
    participant req
    participant Server
    
    User->>req: authenticate https://api.example.com/login
    req->>Server: POST /login
    Server->>req: 302 Redirect + Set-Cookie
    req->>Server: GET /dashboard (with cookie)
    Server->>req: 200 OK + Set-Cookie + JSON
    req->>req: Extract cookies and access_token
    req->>req: Save to ~/.config/req/session_api.example.com.json
    req->>User: Session saved
```

### 2. Auto-Application Flow

For subsequent requests to the same host:

1. `req` checks if a session exists for the host
2. If session exists and no explicit Authorization/Cookie headers are provided, applies session
3. Prints "Using session for <host>" to stderr
4. If explicit headers are provided, uses those instead (session override)

```mermaid
graph TD
    A[Request to api.example.com] --> B{Session exists?}
    B -->|No| C[No session applied]
    B -->|Yes| D{Explicit auth?}
    D -->|Yes| E[Use explicit auth]
    D -->|No| F[Apply session]
    F --> G[Print: Using session for api.example.com]
    E --> H[Execute request]
    G --> H
    C --> H
```

## Session Storage

### Location

Sessions are stored in:
```
~/.config/req/session_<host>.json
```

The host name is sanitized for filename safety:
- Colons (`:`) become underscores (`_`)
- Slashes (`/`) become underscores (`_`)

Examples:
- `api.example.com` → `session_api.example.com.json`
- `localhost:8080` → `session_localhost_8080.json`

### File Format

Session files are JSON:

```json
{
  "host": "api.example.com",
  "cookies": {
    "session_id": "abc123",
    "csrf_token": "xyz789"
  },
  "authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Permissions

Session files are created with strict permissions:
- **Mode**: `0600` (owner read/write only)
- **Security**: Group and world readable files are rejected

If a session file has insecure permissions, `req` will refuse to load it:
```
Error: session file /home/user/.config/req/session_api.example.com.json has insecure permissions (0644): group or world readable, refusing to load
```

## Session Capture

### Set-Cookie Headers

All `Set-Cookie` headers from the authentication response are captured, including those from redirect responses. This is important because many authentication flows use redirects.

Example:
```bash
req authenticate https://httpbin.org/cookies/set?session_id=abc123 using=GET
```

This request:
1. Gets redirected to `/cookies` with `Set-Cookie: session_id=abc123`
2. Captures the cookie from the redirect response
3. Stores it in the session file

### Access Token Extraction

If the response body is JSON and contains a top-level `access_token` field, it's extracted and stored as a Bearer token:

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

The token is stored as:
```json
{
  "authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## Auto-Application Rules

### When Session is Applied

A session is automatically applied when:
1. A request is made to a host with a stored session
2. No explicit `Authorization` header is provided via `include=`
3. No explicit cookies are provided via `include=`

### When Session is NOT Applied

A session is NOT applied when:
1. Explicit `Authorization` header is provided:
   ```bash
   req read https://api.example.com/users \
     include='header: Authorization: Bearer explicit-token' \
     as=json
   # Session NOT used, explicit token used instead
   ```

2. Explicit cookies are provided:
   ```bash
   req read https://api.example.com/users \
     include='cookie: session=explicit-session' \
     as=json
   # Session NOT used, explicit cookie used instead
   ```

3. No session exists for the host

### Override Behavior

Explicit `include=` clauses always override session data. This allows you to:
- Test with different credentials
- Use multiple accounts
- Temporarily override session

## Session Management Commands

### Show Session

Display stored session information (redacted by default):

```bash
# Human-readable (redacted)
req session show api.example.com

# JSON format (full details)
req session show api.example.com as=json
```

**Redaction**: Authorization tokens are shown as `Bearer ***` in human-readable format.

### Clear Session

Delete a stored session:

```bash
req session clear api.example.com
```

This permanently deletes the session file.

### Use Session (Export)

Print environment variable stub for shell scoping:

```bash
req session use api.example.com
# Output: export REQ_SESSION_HOST="api.example.com"
```

Useful in scripts:
```bash
eval $(req session use api.example.com)
# Now REQ_SESSION_HOST is set
```

## Examples

### Complete Authentication Flow

```bash
# 1. Authenticate
req authenticate https://api.example.com/login \
  using=POST \
  with='{"username":"user","password":"pass"}'
# Output: Session saved for api.example.com

# 2. Use session automatically
req read https://api.example.com/me as=json
# Output: Using session for api.example.com
#         { "user": {...} }

# 3. Override session
req read https://api.example.com/me \
  include='header: Authorization: Bearer different-token' \
  as=json
# Output: { "user": {...} }
# (No "Using session" message)

# 4. Check session
req session show api.example.com
# Output: Host: api.example.com
#         Authorization: Bearer ***
#         Cookies: session_id=***

# 5. Clear session
req session clear api.example.com
```

### Multiple Hosts

Sessions are stored per host, so you can have multiple sessions:

```bash
# Authenticate to API 1
req authenticate https://api1.example.com/login \
  using=POST \
  with='{"user":"user1","pass":"pass1"}'
# Session saved for api1.example.com

# Authenticate to API 2
req authenticate https://api2.example.com/login \
  using=POST \
  with='{"user":"user2","pass":"pass2"}'
# Session saved for api2.example.com

# Each request uses the appropriate session
req read https://api1.example.com/users as=json
# Uses api1.example.com session

req read https://api2.example.com/users as=json
# Uses api2.example.com session
```

## Security Considerations

### File Permissions

- Session files must have `0600` permissions
- Files with group or world readable permissions are rejected
- Always check permissions: `ls -l ~/.config/req/session_*.json`

### Credential Storage

- Sessions contain sensitive data (cookies, tokens)
- Never commit session files to version control
- Consider encrypting session files for additional security
- Rotate credentials regularly

### Override Safety

- Explicit `include=` clauses override sessions
- This is intentional for security and flexibility
- Always verify which credentials are being used

## Troubleshooting

### Session Not Applied

**Problem**: Session exists but not being used.

**Solutions**:
1. Check if explicit auth is provided (it overrides session)
2. Verify host matches exactly (including port)
3. Check session file permissions: `ls -l ~/.config/req/session_*.json`
4. Verify session exists: `req session show <host>`

### Session File Permission Error

**Problem**: `Error: session file has insecure permissions`

**Solution**:
```bash
chmod 600 ~/.config/req/session_*.json
```

### Session Not Captured

**Problem**: Authentication succeeds but session not saved.

**Solutions**:
1. Check if `authenticate` verb was used (not `send` or `read`)
2. Verify redirects are being followed (check stderr for redirect trace)
3. Check if Set-Cookie headers are present in response
4. Verify response is JSON if expecting `access_token`

### Multiple Sessions Conflict

**Problem**: Wrong session being used.

**Solution**: Sessions are per-host. Use explicit `include=` to override:
```bash
req read https://api.example.com/users \
  include='header: Authorization: Bearer correct-token' \
  as=json
```

## See Also

- [Authentication Guide](AUTHENTICATION.md) - Authentication methods
- [Security Best Practices](SECURITY.md) - Security considerations
- [Verbs Reference](VERBS.md) - authenticate verb details
- [Error Handling](ERRORS.md) - Session-related errors

