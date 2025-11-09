# req

A semantic HTTP client written in Go that replaces traditional curl syntax with a natural, intent-based grammar.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Documentation](#documentation)
- [Examples](#examples)
- [Comparison with curl](#comparison-with-curl)
- [Current Status](#current-status)
- [Contributing](#contributing)
- [License](#license)

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

## Documentation

Comprehensive documentation is available in the [`docs/`](docs/) directory:

- **[Getting Started](docs/GETTING_STARTED.md)** - Installation and first steps
- **[Grammar Reference](docs/GRAMMAR.md)** - Complete grammar specification
- **[Verbs Reference](docs/VERBS.md)** - Detailed documentation for each verb
- **[Clauses Reference](docs/CLAUSES.md)** - Complete clause reference
- **[Examples Cookbook](docs/EXAMPLES.md)** - Comprehensive examples
- **[Architecture](docs/ARCHITECTURE.md)** - System architecture with diagrams
- **[Authentication](docs/AUTHENTICATION.md)** - All authentication methods
- **[Session Management](docs/SESSIONS.md)** - Session deep dive
- **[Error Handling](docs/ERRORS.md)** - Exit codes and troubleshooting
- **[Security Best Practices](docs/SECURITY.md)** - Security guide
- **[curl Migration Guide](docs/CURL_MIGRATION.md)** - Migrate from curl
- **[Advanced Usage](docs/ADVANCED.md)** - Advanced patterns and tips
- **[Cross-Shell Quoting](docs/QUOTING.md)** - Quoting guide
- **[Contributing](docs/CONTRIBUTING.md)** - Contribution guidelines

See the [Documentation Index](docs/README.md) for complete navigation.

## Examples

### Basic Requests

```bash
# GET request
req read https://api.example.com/users as=json

# POST with JSON
req send https://api.example.com/users with='{"name":"Alice"}'

# With authentication
req read https://api.example.com/users \
  include='header: Authorization: Bearer $TOKEN' \
  as=json
```

### File Operations

```bash
# Download file
req save https://example.com/file.zip to=file.zip

# Upload file
req upload https://api.example.com/upload \
  attach='part: name=file, file=@./document.pdf' \
  as=json
```

### Sessions

```bash
# Authenticate and store session
req authenticate https://api.example.com/login \
  using=POST \
  with='{"username":"user","password":"pass"}'

# Session automatically used
req read https://api.example.com/me as=json
```

For more examples, see the [Examples Cookbook](docs/EXAMPLES.md).

## Comparison with curl

| Task | curl | req |
|------|------|-----|
| **Basic GET** | `curl https://api.example.com/users` | `req read https://api.example.com/users` |
| **GET with headers** | `curl -H "Authorization: Bearer $TOKEN" https://api.example.com/users` | `req read https://api.example.com/users include="header: Authorization: Bearer $TOKEN"` |
| **POST JSON** | `curl -X POST -H "Content-Type: application/json" -d '{"name":"Adam"}' https://api.example.com/users` | `req send https://api.example.com/users with='{"name":"Adam"}'` |
| **Multipart upload** | `curl -F "file=@avatar.png" https://api.example.com/upload` | `req upload https://api.example.com/upload attach='part: name=file, file=@avatar.png'` |
| **Basic Auth** | `curl -u user:pass https://api.example.com` | `req read https://api.example.com include='basic: user:pass'` |
| **Follow redirects** | `curl -L https://example.com` | `req read https://example.com` (default) |
| **Ignore SSL** | `curl -k https://self-signed.example.com` | `req read https://self-signed.example.com insecure=true` |

See the [curl Migration Guide](docs/CURL_MIGRATION.md) for detailed comparisons and migration tips.

## Current Status

**v0.1** - Core functionality complete

- ✅ Command parsing with full grammar validation
- ✅ All clauses implemented
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
- ✅ Basic Auth support

## Roadmap

- **v0.1** ✅ - Core functionality (current)
- **v0.2** - Watch mode with SSE and polling
- **v0.3** - JSONPath selection and filtering
- **v0.4** - Advanced retry and backoff strategies
- **v1.0** - Stability hardening and release candidates

## Contributing

Contributions are welcome! Please see our [Contributing Guidelines](docs/CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) file for details.
