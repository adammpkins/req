# req

A semantic HTTP client written in Go that replaces traditional curl syntax with a natural, intent-based grammar.

## Overview

`req` is an HTTP client tool that focuses on:

- **Human-readable commands** (verbs + clauses)
- **Sensible defaults** (follow redirects, TLS verify, retries)
- **JSON/CSV/text awareness** with intelligent output
- **Watch mode** (poll or stream)
- **Session management** (authenticate and auto-apply)
- **Pretty diagnostics** and dry-run transparency

## Quick Start

```bash
# Read JSON from an API
req read https://api.example.com/users as=json

# Send JSON data
req send https://api.example.com/users with='{"name":"Adam"}'

# Send with headers and assertions
req send https://api.example.com/users \
  using=POST \
  include='header: Authorization: Bearer $TOKEN' \
  with='{"name":"Adam"}' \
  expect=status:201, header:Content-Type=application/json \
  as=json

# Save a file
req save https://example.com/file.zip to=file.zip

# Upload multipart form data
req upload https://api.example.com/upload \
  attach='part: name=file, file=@./avatar.png, type=image/png' \
  as=json

# Authenticate and store session
req authenticate https://api.example.com/login \
  using=POST \
  with='{"user":"adam","pass":"xyz"}'

# Use stored session automatically
req read https://api.example.com/me as=json
```

## Installation

```bash
go install github.com/adammpkins/req/cmd/req@latest
```

Or download a pre-built binary from the [Releases](https://github.com/adammpkins/req/releases) page.

## Grammar

The `req` command follows this grammar:

```
req <verb> <target> [clauses...]
```

### Verbs

- `read` - GET, print to stdout
- `save` - GET, write to file via to=
- `send` - default GET, POST if `with=` is present
- `upload` - POST when `attach=` or `with=` present, else error
- `watch` - GET with SSE or polling
- `inspect` - HEAD only
- `authenticate` - login and store session state
- `session` - session management (show, clear, use)

### Clauses

- `using=<method>` - HTTP method override
- `include=<items>` - Add headers, params, cookies (repeatable)
  - Format: `include='header: Name: Value; param: key=value; cookie: key=value'`
- `with=<body>` - Request body
  - Format: `with=@user.json` or `with='{"name":"Adam"}'`
  - JSON inference: Automatically sets Content-Type for JSON when inline starts with `{` or `[`
- `expect=<checks>` - Assertions on response
  - Format: `expect=status:200, header:Content-Type=application/json, contains:"ok"`
- `as=<format>` - Output format for stdout
- `to=<path>` - Destination path
- `retry=<count>` - Retry attempts for transient errors
- `under=<limit>` - Timeout or size limit
  - Format: `under=30s` or `under=10MB`
- `via=<url>` - Proxy URL
- `attach=<parts>` - Multipart parts for upload or send (repeatable)
  - Format: `attach='part: name=avatar, file=@me.png; part: name=meta, value=xyz'`
- `follow=<policy>` - Redirect policy for write verbs
  - Format: `follow=smart`
- `insecure=<bool>` - Disable TLS verification for this request
  - Format: `insecure=true`

## Examples

### Read JSON

```bash
req read https://api.example.com/users as=json
```

### Send JSON Data

```bash
req send https://api.example.com/users with='{"name":"Ada","email":"ada@example.com"}'
```

### Save a File

```bash
# Save with explicit filename
req save https://example.com/file.zip to=file.zip

# Save with auto-detected filename (extracts from URL)
req save https://example.com/file.zip

# Save to directory path
req save https://example.com/file.zip to=/tmp/file.zip
```

### With Headers, Params, and Cookies

```bash
# Using include clause
req read https://api.example.com/search \
  include='header: Authorization: Bearer $TOKEN; param: q=search query; cookie: session=abc123' \
  as=json
```

### With Assertions

```bash
req send https://api.example.com/users \
  using=POST \
  with='{"name":"Adam"}' \
  expect=status:201, header:Content-Type=application/json, contains:"id" \
  as=json
```

### Sessions

```bash
# Authenticate and store session
req authenticate https://api.example.com/login \
  using=POST \
  with='{"user":"adam","pass":"xyz"}'

# Session is automatically used for subsequent requests
req read https://api.example.com/me as=json

# Show stored session (redacted)
req session show api.example.com

# Show session in JSON format
req session show api.example.com as=json

# Clear session
req session clear api.example.com
```

### Redirects

```bash
# Read and save follow redirects by default (up to 5)
req read https://example.com/redirect

# Write verbs don't follow by default
req send https://api.example.com/create using=POST with='{"data":"value"}'

# Use smart follow for write verbs (only follows 307/308)
req send https://api.example.com/create \
  using=POST \
  with='{"data":"value"}' \
  follow=smart
```

### With Retry and Timeout

```bash
req read https://api.example.com/users retry=3 under=10s as=json
```

### Edge Cases

```bash
# Header with commas and q values (must be quoted)
req read https://api.example.com/search \
  include='header: Accept: application/json, application/problem+json; q=0.9' \
  as=json

# Cookie value containing semicolons (must be quoted)
req read https://api.example.com/search \
  include='cookie: prefs="a=1; b=2; c=3"' \
  as=json

# Multipart upload with file and text parts (Content-Type automatically overridden)
req upload https://api.example.com/upload \
  include='header: Content-Type: application/json' \
  attach='part: name=file, file=@avatar.png; part: name=meta, value={"name":"test"}' \
  as=json
# Note: Content-Type will be overridden to multipart/form-data

# Smart redirect on write verb (only follows 307/308)
req send https://api.example.com/create \
  using=POST \
  with='{"data":"value"}' \
  follow=smart
# Will follow 307/308 redirects, but not 301/302/303

# Write verb with 303 redirect (advisory printed, not followed)
req send https://api.example.com/create \
  using=POST \
  with='{"data":"value"}'
# If server returns 303, advisory message printed but redirect not followed
```

### Method Override

```bash
# Use PUT instead of POST
req send https://api.example.com/users/1 using=PUT with='{"name":"Updated"}'

# Use PATCH for partial updates
req send https://api.example.com/users/1 using=PATCH with='{"email":"new@example.com"}'

# Use HEAD to check headers without body
req read https://api.example.com/users using=HEAD
```

**Note:** The `using=` clause validates method-verb compatibility. For example, `read using=POST` will fail as `read` only allows GET, HEAD, or OPTIONS.

### Comparison with curl

| Task | curl | req |
|------|------|-----|
| **Basic GET with headers** | `curl -H "Authorization: Bearer $TOKEN" https://api.example.com/users` | `req read https://api.example.com/users include='header: Authorization: Bearer $TOKEN'` |
| **Multipart upload** | `curl -F "file=@avatar.png" -F "name=test" https://api.example.com/upload` | `req upload https://api.example.com/upload attach='part: name=file, file=@avatar.png; part: name=name, value=test'` |
| **Authenticated POST** | `curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"name":"Adam"}' https://api.example.com/users` | `req send https://api.example.com/users using=POST include='header: Authorization: Bearer $TOKEN' with='{"name":"Adam"}'` |

### Dry Run

```bash
req read https://api.example.com/users as=json --dry-run
```

### Interactive TUI Mode

```bash
# Launch interactive TUI mode
req --tui

# Or run without arguments to launch TUI
req
```

The TUI mode provides:
- Interactive command builder with form-based input
- **Syntax-highlighted JSON output** with color-coded keys, values, and punctuation
- **Scrollable viewport** for long responses with keyboard navigation
- Pretty-printed JSON with automatic indentation
- Real-time command execution and response display

**Keyboard Controls:**
- `↑` / `↓` or `k` / `j` - Scroll line by line
- `pgup` / `pgdown` - Page scrolling
- `home` - Jump to top
- `end` - Jump to bottom
- `ctrl+u` / `ctrl+d` - Half-page scrolling
- `esc` - Quit TUI

## Security

### Shell History

**Warning:** Commands containing secrets (tokens, passwords) are stored in your shell history by default. Use environment variables to avoid exposing secrets:

```bash
# Bad: Token appears in shell history
req read https://api.example.com/users include='header: Authorization: Bearer secret-token-123'

# Good: Use environment variable
TOKEN="secret-token-123"
req read https://api.example.com/users include="header: Authorization: Bearer $TOKEN"
```

To prevent secrets from being saved to history:
- **Bash/Zsh:** Prefix command with a space (requires `HISTCONTROL=ignorespace` or `setopt HIST_IGNORE_SPACE`)
- **Fish:** Use `history --delete` after running commands with secrets
- **PowerShell:** Use `Set-PSReadlineOption -HistoryNoDuplicates` and manually edit history

### Session Files

Session files are stored in `~/.config/req/session_<host>.json` with permissions `0600` (owner read/write only).

**Security rules:**
- Session files are created with strict permissions (`0600`)
- If a session file has group or world readable permissions, `req` will refuse to load it
- Session files contain sensitive data (cookies, tokens) and should be protected
- Never commit session files to version control

To check session file permissions:
```bash
ls -l ~/.config/req/session_*.json
```

## Current Status

**v0.1** - Core functionality complete

- ✅ Command parsing with full grammar validation
- ✅ All clauses implemented (include, attach, expect, follow, insecure, etc.)
- ✅ Execution plan generation with verb defaults
- ✅ HTTP request execution with redirect handling
- ✅ Transparent compression (gzip, br)
- ✅ Session management (authenticate, session show/clear/use)
- ✅ Auto-apply sessions for matching hosts
- ✅ File downloads with automatic filename extraction
- ✅ Multipart form data support
- ✅ Response assertions (expect clause)
- ✅ Proper exit codes (0 success, 3 expect fail, 4 network, 5 grammar)
- ✅ Helpful error messages with suggestions
- ✅ Help and explain commands
- ✅ Interactive TUI mode
- ✅ JSON output formatting
- ✅ Stderr meta output with redaction

## Roadmap

- **v0.1** ✅ - Core functionality (current)
- **v0.2** - Watch mode with SSE and polling
- **v0.3** - JSONPath selection and filtering
- **v0.4** - Advanced retry and backoff strategies
- **v1.0** - Stability hardening and release candidates

## Contributing

Contributions are welcome! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) file for details.

