# curl Migration Guide

This guide helps you migrate from `curl` to `req` with side-by-side comparisons.

## Quick Reference

| Task | curl | req |
|------|------|-----|
| Basic GET | `curl https://api.example.com/users` | `req read https://api.example.com/users` |
| GET with headers | `curl -H "Authorization: Bearer $TOKEN" https://api.example.com/users` | `req read https://api.example.com/users include="header: Authorization: Bearer $TOKEN"` |
| POST JSON | `curl -X POST -H "Content-Type: application/json" -d '{"name":"Alice"}' https://api.example.com/users` | `req send https://api.example.com/users with='{"name":"Alice"}'` |
| Multipart upload | `curl -F "file=@avatar.png" https://api.example.com/upload` | `req upload https://api.example.com/upload attach='part: name=file, file=@avatar.png'` |
| Save file | `curl -O https://example.com/file.zip` | `req save https://example.com/file.zip` |
| Follow redirects | `curl -L https://example.com` | `req read https://example.com` (default) |
| Ignore SSL | `curl -k https://self-signed.example.com` | `req read https://self-signed.example.com insecure=true` |
| Proxy | `curl --proxy http://proxy:8080 https://api.example.com` | `req read https://api.example.com via=http://proxy:8080` |
| Cookies | `curl -b "session=abc123" https://api.example.com` | `req read https://api.example.com include='cookie: session=abc123'` |
| Basic Auth | `curl -u user:pass https://api.example.com` | `req read https://api.example.com include='basic: user:pass'` |

## Common Patterns

### Basic GET Request

**curl**:
```bash
curl https://api.example.com/users
```

**req**:
```bash
req read https://api.example.com/users
```

### GET with Headers

**curl**:
```bash
curl -H "Authorization: Bearer $TOKEN" \
     -H "Accept: application/json" \
     https://api.example.com/users
```

**req**:
```bash
req read https://api.example.com/users \
  include='header: Authorization: Bearer $TOKEN; header: Accept: application/json' \
  as=json
```

### POST with JSON Body

**curl**:
```bash
curl -X POST \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer $TOKEN" \
     -d '{"name":"Alice","email":"alice@example.com"}' \
     https://api.example.com/users
```

**req**:
```bash
req send https://api.example.com/users \
  using=POST \
  include="header: Authorization: Bearer $TOKEN" \
  with='{"name":"Alice","email":"alice@example.com"}' \
  as=json
```

**Note**: `req` automatically infers `Content-Type: application/json` when body starts with `{` or `[`.

### PUT Request

**curl**:
```bash
curl -X PUT \
     -H "Content-Type: application/json" \
     -d '{"name":"Bob"}' \
     https://api.example.com/users/1
```

**req**:
```bash
req send https://api.example.com/users/1 \
  using=PUT \
  with='{"name":"Bob"}' \
  as=json
```

### DELETE Request

**curl**:
```bash
curl -X DELETE \
     -H "Authorization: Bearer $TOKEN" \
     https://api.example.com/users/1
```

**req**:
```bash
req send https://api.example.com/users/1 \
  using=DELETE \
  include="header: Authorization: Bearer $TOKEN"
```

### Query Parameters

**curl**:
```bash
curl "https://api.example.com/search?q=test&page=1"
```

**req**:
```bash
req read https://api.example.com/search \
  include='param: q=test; param: page=1' \
  as=json
```

Or include in URL:
```bash
req read "https://api.example.com/search?q=test&page=1" as=json
```

### Cookies

**curl**:
```bash
curl -b "session=abc123; csrf=xyz789" https://api.example.com/users
```

**req**:
```bash
req read https://api.example.com/users \
  include='cookie: session=abc123; cookie: csrf=xyz789' \
  as=json
```

### Basic Authentication

**curl**:
```bash
curl -u user:pass https://api.example.com/users
```

**req**:
```bash
req read https://api.example.com/users \
  include='basic: user:pass' \
  as=json
```

### File Upload (Multipart)

**curl**:
```bash
curl -X POST \
     -F "file=@avatar.png" \
     -F "name=My Avatar" \
     https://api.example.com/upload
```

**req**:
```bash
req upload https://api.example.com/upload \
  attach='part: name=file, file=@avatar.png; part: name=name, value=My Avatar' \
  as=json
```

### File Upload with Metadata

**curl**:
```bash
curl -X POST \
     -F "file=@document.pdf" \
     -F "title=Document" \
     -F "description=My document" \
     https://api.example.com/upload
```

**req**:
```bash
req upload https://api.example.com/upload \
  attach='part: name=file, file=@document.pdf; part: name=title, value=Document; part: name=description, value=My document' \
  as=json
```

### Download File

**curl**:
```bash
curl -O https://example.com/file.zip
# or
curl -o file.zip https://example.com/file.zip
```

**req**:
```bash
req save https://example.com/file.zip to=file.zip
# or (auto-detect filename)
req save https://example.com/file.zip
```

### Follow Redirects

**curl**:
```bash
curl -L https://example.com/redirect
```

**req**:
```bash
req read https://example.com/redirect
# Follows redirects by default for read/save verbs
```

### Ignore SSL Certificate

**curl**:
```bash
curl -k https://self-signed.example.com
```

**req**:
```bash
req read https://self-signed.example.com insecure=true
# Warning: TLS verification disabled
```

### Proxy

**curl**:
```bash
curl --proxy http://proxy.example.com:8080 https://api.example.com
```

**req**:
```bash
req read https://api.example.com via=http://proxy.example.com:8080
```

### Timeout

**curl**:
```bash
curl --max-time 30 https://api.example.com
```

**req**:
```bash
req read https://api.example.com under=30s
```

### Retry

**curl**:
```bash
curl --retry 3 https://api.example.com
```

**req**:
```bash
req read https://api.example.com retry=3
```

### Verbose Output

**curl**:
```bash
curl -v https://api.example.com
```

**req**:
```bash
req read https://api.example.com verbose
# Shows request/response details
```

### Save Headers

**curl**:
```bash
curl -D headers.txt https://api.example.com
```

**req**:
```bash
# Headers are in stderr, redirect if needed
req read https://api.example.com 2> headers.txt
```

### Custom User-Agent

**curl**:
```bash
curl -H "User-Agent: MyApp/1.0" https://api.example.com
```

**req**:
```bash
req read https://api.example.com \
  include='header: User-Agent: MyApp/1.0' \
  as=json
```

### POST from File

**curl**:
```bash
curl -X POST \
     -H "Content-Type: application/json" \
     --data-binary @data.json \
     https://api.example.com/users
```

**req**:
```bash
req send https://api.example.com/users \
  using=POST \
  with=@data.json \
  as=json
```

### POST from Stdin

**curl**:
```bash
echo '{"name":"Alice"}' | curl -X POST \
     -H "Content-Type: application/json" \
     --data-binary @- \
     https://api.example.com/users
```

**req**:
```bash
echo '{"name":"Alice"}' | req send https://api.example.com/users with=@- as=json
```

## Key Differences

### 1. Default Behavior

**curl**: Minimal defaults, explicit flags required
**req**: Sensible defaults (follow redirects, TLS verify, JSON inference)

### 2. Command Structure

**curl**: Flag-based (`-X`, `-H`, `-d`, etc.)
**req**: Verb + clauses (`read`, `send`, `include=`, `with=`)

### 3. Redirect Following

**curl**: Must use `-L` flag
**req**: Follows by default for `read`/`save` verbs

### 4. JSON Handling

**curl**: Must set `Content-Type` header
**req**: Auto-infers `application/json` from body content

### 5. Output Formatting

**curl**: Raw output by default
**req**: Auto-formats JSON, CSV, etc. with `as=` clause

### 6. Error Handling

**curl**: Exit codes vary, less structured
**req**: Consistent exit codes (0, 3, 4, 5)

### 7. Session Management

**curl**: Manual cookie handling (`-b`, `-c`)
**req**: Automatic session management with `authenticate` verb

## Migration Checklist

- [ ] Replace `curl` with `req read` for GET requests
- [ ] Replace `curl -X POST` with `req send` for POST requests
- [ ] Convert `-H` headers to `include='header: ...'`
- [ ] Convert `-d` data to `with='...'`
- [ ] Convert `-F` multipart to `attach='part: ...'`
- [ ] Convert `-b` cookies to `include='cookie: ...'`
- [ ] Convert `-u` Basic Auth to `include='basic: ...'`
- [ ] Convert `-L` redirects (default in req)
- [ ] Convert `-k` to `insecure=true`
- [ ] Convert `--proxy` to `via=`
- [ ] Convert `--max-time` to `under=`
- [ ] Convert `-O`/`-o` to `save` verb with `to=`
- [ ] Test all commands with `req explain` first

## Feature Parity

### ✅ Supported in req

- GET, POST, PUT, PATCH, DELETE requests
- Headers, cookies, query parameters
- JSON and form data bodies
- Multipart file uploads
- Redirect following
- Proxy support
- Timeout and retry
- Basic Auth and Bearer tokens
- File downloads
- Stdin input

### ⚠️ Different in req

- Session management (automatic vs manual)
- Output formatting (structured vs raw)
- Error codes (structured vs varied)
- JSON inference (automatic vs manual)

### ❌ Not Yet Supported

- SOCKS proxy
- Custom CA certificates
- Client certificates
- HTTP/2 specific features
- Some advanced curl options

## Tips for Migration

1. **Start Simple**: Migrate basic GET requests first
2. **Use `req explain`**: Verify commands before executing
3. **Test Incrementally**: Migrate one command at a time
4. **Check Exit Codes**: Update scripts to handle req's exit codes
5. **Leverage Defaults**: Take advantage of req's sensible defaults

## See Also

- [Getting Started](GETTING_STARTED.md) - Installation and basics
- [Grammar Reference](GRAMMAR.md) - Complete syntax
- [Examples Cookbook](EXAMPLES.md) - More examples
- [Verbs Reference](VERBS.md) - Verb details

