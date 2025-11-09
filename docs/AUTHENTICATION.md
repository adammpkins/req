# Authentication

This document covers all authentication methods supported by `req`.

## Authentication Methods

`req` supports three authentication methods:

1. **Basic Auth** - Username/password via `include='basic: user:pass'`
2. **Bearer Token** - Token via `include='header: Authorization: Bearer token'`
3. **Session-Based** - Automatic via `authenticate` verb

## Basic Auth

Basic Auth sends credentials with every request using HTTP Basic Authentication.

### Syntax

```bash
include='basic: username:password'
```

### How It Works

1. Credentials are Base64-encoded automatically
2. `Authorization: Basic <encoded>` header is set
3. Credentials are sent with every request
4. **Not stored** in sessions (per-request only)

### Examples

```bash
# Simple Basic Auth
req read https://httpbin.org/basic-auth/user/passwd \
  include='basic: user:passwd' \
  expect=status:200 \
  as=json

# Basic Auth with other headers
req read https://api.example.com/users \
  include='basic: user:pass; header: Accept: application/json' \
  as=json

# Basic Auth in POST request
req send https://api.example.com/data \
  using=POST \
  include='basic: admin:secret' \
  with='{"data":"value"}' \
  as=json
```

### Security Considerations

- **Per-request only**: Credentials are not stored
- **Base64 encoding**: Not encryption (credentials are visible if intercepted)
- **Shell history**: Credentials appear in command history
- **Best practice**: Use environment variables

```bash
# Bad: Credentials in command history
req read https://api.example.com/users include='basic: user:secret123'

# Good: Use environment variables
USERNAME="user"
PASSWORD="secret123"
req read https://api.example.com/users \
  include="basic: $USERNAME:$PASSWORD" \
  as=json
```

## Bearer Token

Bearer tokens are sent via the `Authorization` header.

### Syntax

```bash
include='header: Authorization: Bearer <token>'
```

### Examples

```bash
# Simple Bearer token
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
req read https://api.example.com/users \
  include="header: Authorization: Bearer $TOKEN" \
  as=json

# Bearer token with other headers
req read https://api.example.com/users \
  include='header: Authorization: Bearer token; header: Accept: application/json' \
  as=json

# Bearer token in POST
req send https://api.example.com/users \
  using=POST \
  include="header: Authorization: Bearer $TOKEN" \
  with='{"name":"Alice"}' \
  as=json
```

### Token Sources

**Environment Variable** (Recommended):
```bash
export API_TOKEN="your-token-here"
req read https://api.example.com/users \
  include="header: Authorization: Bearer $API_TOKEN" \
  as=json
```

**Script Variable**:
```bash
#!/bin/bash
TOKEN=$(get_token_from_vault)
req read https://api.example.com/users \
  include="header: Authorization: Bearer $TOKEN" \
  as=json
```

## Session-Based Authentication

Session-based auth uses the `authenticate` verb to store credentials automatically.

### Flow

1. **Authenticate**: Use `authenticate` verb to login
2. **Capture**: Cookies and tokens are captured automatically
3. **Store**: Session saved per host
4. **Auto-apply**: Subsequent requests use session automatically

### Examples

```bash
# Step 1: Authenticate
req authenticate https://api.example.com/login \
  using=POST \
  with='{"username":"user","password":"pass"}'
# Output: Session saved for api.example.com

# Step 2: Use session automatically
req read https://api.example.com/me as=json
# Output: Using session for api.example.com
#         { "user": {...} }
```

### What Gets Captured

- **Set-Cookie headers**: All cookies from response and redirects
- **access_token**: If response JSON contains `access_token` field

### Session Storage

- **Location**: `~/.config/req/session_<host>.json`
- **Permissions**: `0600` (owner read/write only)
- **Format**: JSON with cookies and authorization

See [Session Management](SESSIONS.md) for detailed information.

## Choosing an Authentication Method

### Use Basic Auth When:
- ✅ Simple username/password authentication
- ✅ Credentials change frequently
- ✅ Don't want credentials stored
- ✅ Testing or one-off requests

### Use Bearer Token When:
- ✅ API uses token-based auth
- ✅ Token is long-lived
- ✅ Want explicit control per request
- ✅ Token comes from external source (vault, env)

### Use Session-Based When:
- ✅ Login flow returns cookies/tokens
- ✅ Want automatic credential management
- ✅ Multiple requests to same API
- ✅ Credentials are session-scoped

## Authentication Precedence

When multiple auth methods are specified, precedence is:

1. **Explicit `include=`** (highest priority)
2. **Session auto-apply** (if no explicit auth)
3. **No auth** (lowest priority)

### Examples

```bash
# Session exists, but explicit header overrides
req read https://api.example.com/users \
  include='header: Authorization: Bearer explicit-token' \
  as=json
# Uses explicit-token, NOT session

# Basic Auth overrides session
req read https://api.example.com/users \
  include='basic: user:pass' \
  as=json
# Uses Basic Auth, NOT session

# No explicit auth, session used
req read https://api.example.com/users as=json
# Uses session automatically
```

## Security Best Practices

### 1. Use Environment Variables

```bash
# Bad
req read https://api.example.com/users \
  include='header: Authorization: Bearer secret-token-123'

# Good
export API_TOKEN="secret-token-123"
req read https://api.example.com/users \
  include="header: Authorization: Bearer $API_TOKEN" \
  as=json
```

### 2. Protect Session Files

```bash
# Check permissions
ls -l ~/.config/req/session_*.json

# Should be: -rw------- (0600)
# If not: chmod 600 ~/.config/req/session_*.json
```

### 3. Avoid Shell History

```bash
# Bash/Zsh: Prefix with space (requires HISTCONTROL=ignorespace)
 req read https://api.example.com/users include='header: Authorization: Bearer secret'

# Or use environment variables (recommended)
export TOKEN="secret"
req read https://api.example.com/users include="header: Authorization: Bearer $TOKEN"
```

### 4. Rotate Credentials

- Regularly rotate API tokens
- Clear old sessions: `req session clear <host>`
- Use short-lived tokens when possible

### 5. Use HTTPS

Always use HTTPS for authentication:
```bash
# Good
req authenticate https://api.example.com/login ...

# Bad (credentials sent in plain text)
req authenticate http://api.example.com/login ...
```

## Troubleshooting

### Basic Auth Not Working

**Problem**: Getting 401 Unauthorized.

**Solutions**:
1. Verify credentials are correct
2. Check format: `basic: username:password` (colon separator)
3. Ensure URL uses HTTPS
4. Check if server requires additional headers

### Bearer Token Not Working

**Problem**: Getting 401 Unauthorized.

**Solutions**:
1. Verify token is valid and not expired
2. Check header format: `Bearer <token>` (space after Bearer)
3. Ensure token is not quoted in header value
4. Check if token needs to be refreshed

### Session Not Captured

**Problem**: Authentication succeeds but session not saved.

**Solutions**:
1. Verify `authenticate` verb is used (not `send`)
2. Check if Set-Cookie headers are present
3. Verify redirects are being followed
4. Check session file permissions

### Wrong Credentials Used

**Problem**: Unexpected authentication.

**Solutions**:
1. Check for explicit `include=` clauses (they override session)
2. Verify which session is being used: `req session show <host>`
3. Check host matching (including port)
4. Use `req explain` to see parsed command

## See Also

- [Session Management](SESSIONS.md) - Session details
- [Security Best Practices](SECURITY.md) - Security guide
- [Clauses Reference](CLAUSES.md) - include= clause details
- [Verbs Reference](VERBS.md) - authenticate verb details

