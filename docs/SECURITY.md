# Security Best Practices

This document covers security considerations and best practices when using `req`.

## Credential Handling

### Never Put Credentials in Commands

**Bad**: Credentials appear in shell history
```bash
req read https://api.example.com/users \
  include='header: Authorization: Bearer secret-token-123'
```

**Good**: Use environment variables
```bash
export API_TOKEN="secret-token-123"
req read https://api.example.com/users \
  include="header: Authorization: Bearer $API_TOKEN"
```

### Environment Variables

Always use environment variables for sensitive data:

```bash
# Set in shell
export API_TOKEN="your-token"
export USERNAME="user"
export PASSWORD="pass"

# Use in commands
req read https://api.example.com/users \
  include="header: Authorization: Bearer $API_TOKEN" \
  as=json

req read https://api.example.com/users \
  include="basic: $USERNAME:$PASSWORD" \
  as=json
```

### Shell History Protection

#### Bash/Zsh

Enable `HISTCONTROL=ignorespace`:
```bash
export HISTCONTROL=ignorespace
# Commands starting with space are not saved to history
 req read https://api.example.com/users include='header: Authorization: Bearer secret'
```

#### Fish

Delete from history after running:
```fish
req read https://api.example.com/users include='header: Authorization: Bearer secret'
history --delete
```

#### PowerShell

Use `Set-PSReadlineOption`:
```powershell
Set-PSReadlineOption -HistoryNoDuplicates
# Manually edit history if needed
```

## Session File Security

### File Permissions

Session files are stored with strict permissions (`0600`):

```bash
# Check permissions
ls -l ~/.config/req/session_*.json
# Should show: -rw------- (0600)

# Fix if needed
chmod 600 ~/.config/req/session_*.json
```

### Permission Enforcement

`req` refuses to load session files with insecure permissions:

```bash
# If file has group/world readable permissions
req read https://api.example.com/users
# Error: session file ... has insecure permissions (0644): group or world readable, refusing to load
```

**Fix**:
```bash
chmod 600 ~/.config/req/session_*.json
```

### Session File Location

- **Path**: `~/.config/req/session_<host>.json`
- **Permissions**: `0600` (owner read/write only)
- **Content**: Cookies and authorization tokens (sensitive)

### Never Commit Session Files

**Never** commit session files to version control:

```bash
# Add to .gitignore
echo "~/.config/req/session_*.json" >> .gitignore
```

## TLS/SSL Security

### Always Use HTTPS

**Bad**: Credentials sent in plain text
```bash
req authenticate http://api.example.com/login ...
```

**Good**: Use HTTPS
```bash
req authenticate https://api.example.com/login ...
```

### Certificate Verification

By default, `req` verifies TLS certificates. Only disable verification for:
- Development/testing environments
- Self-signed certificates in trusted networks
- Troubleshooting TLS issues

**Never** use `insecure=true` in production:

```bash
# Development only
req read https://self-signed-dev.example.com/users insecure=true

# Production - verify certificates
req read https://api.example.com/users
```

### Warning Message

When `insecure=true` is used, a warning is printed:
```
Warning: TLS verification disabled
```

## Proxy Security

### Use Trusted Proxies

Only use proxies you trust:

```bash
# Trusted proxy
req read https://api.example.com/users via=http://trusted-proxy.example.com:8080

# Untrusted proxy - avoid
req read https://api.example.com/users via=http://unknown-proxy.com:8080
```

### Proxy Credentials

If proxy requires authentication, include credentials securely:

```bash
# Use environment variables
export PROXY_USER="user"
export PROXY_PASS="pass"
# Note: req doesn't currently support proxy auth directly
# Consider using environment variables for proxy URL with credentials
```

## Basic Auth Security

### Base64 Encoding is Not Encryption

Basic Auth uses Base64 encoding, which is **not encryption**:

- Credentials are easily decoded if intercepted
- Always use HTTPS with Basic Auth
- Consider using tokens instead of passwords

### Per-Request Only

Basic Auth credentials are **not stored** in sessions (by design):

```bash
# Credentials sent with request, not stored
req read https://api.example.com/users include='basic: user:pass'
```

This is a security feature - credentials are not persisted.

## Bearer Token Security

### Token Storage

- Store tokens in environment variables or secure vaults
- Never hardcode tokens in scripts
- Rotate tokens regularly
- Use short-lived tokens when possible

### Token Exposure

Tokens in commands may appear in:
- Shell history
- Process lists (`ps aux`)
- Log files
- Screen sharing

Always use environment variables:

```bash
# Bad: Token visible in process list
req read https://api.example.com/users include='header: Authorization: Bearer token123'

# Good: Token from environment
export TOKEN="token123"
req read https://api.example.com/users include="header: Authorization: Bearer $TOKEN"
```

## Redaction Behavior

### Stderr Output

Sensitive data is redacted in stderr output:

```
HTTP 200
URL: https://api.example.com/users
Size: 1024 bytes
Content-Type: application/json
Authorization: Bearer ***
```

### Session Display

Sessions shown with `session show` are redacted:

```bash
req session show api.example.com
# Output:
# Host: api.example.com
# Authorization: Bearer ***
# Cookies: session_id=***
```

Use `as=json` for full details (use with caution):

```bash
req session show api.example.com as=json
# Shows full token (use carefully)
```

## Best Practices Summary

### ✅ Do

- Use environment variables for credentials
- Always use HTTPS
- Verify TLS certificates in production
- Protect session files (`chmod 600`)
- Rotate credentials regularly
- Use short-lived tokens
- Check session file permissions
- Use `req explain` to verify commands before execution

### ❌ Don't

- Put credentials directly in commands
- Commit session files to version control
- Use `insecure=true` in production
- Share session files
- Use HTTP for authentication
- Store credentials in scripts
- Ignore permission warnings

## Security Checklist

Before using `req` in production:

- [ ] All credentials use environment variables
- [ ] HTTPS is used for all requests
- [ ] TLS verification is enabled (`insecure=false` or omitted)
- [ ] Session files have `0600` permissions
- [ ] Session files are in `.gitignore`
- [ ] Shell history protection is configured
- [ ] Credentials are rotated regularly
- [ ] Proxy is trusted (if used)
- [ ] Error messages don't expose sensitive data
- [ ] Logs don't contain credentials

## Incident Response

If credentials are exposed:

1. **Immediately revoke** exposed credentials
2. **Clear session files**: `req session clear <host>`
3. **Rotate** all related credentials
4. **Review** shell history and logs
5. **Check** for unauthorized access
6. **Update** security practices

## See Also

- [Authentication Guide](AUTHENTICATION.md) - Authentication methods
- [Session Management](SESSIONS.md) - Session security
- [Error Handling](ERRORS.md) - Security-related errors
- [Cross-Shell Quoting](QUOTING.md) - Secure quoting practices

