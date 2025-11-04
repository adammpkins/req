# req

A semantic HTTP client written in Go that replaces traditional curl syntax with a natural, intent-based grammar.

## Overview

`req` is an HTTP client tool that focuses on:

- **Human-readable commands** (verbs + clauses)
- **Sensible defaults** (follow redirects, TLS verify, retries)
- **JSON/CSV/text awareness** with intelligent output
- **Watch mode** (poll or stream)
- **Auth profiles and sessions**
- **Pretty diagnostics** and dry-run transparency

## Quick Start

```bash
# Read JSON from an API
req read https://api.example.com/users as=json

# Send JSON data
req send https://api.example.com/users with=json:'{"name":"Ada"}'

# Save a file
req save https://example.com/file.zip to=file.zip
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

- `read` - Read a resource (defaults to GET)
- `save` - Save a resource to a file (defaults to GET)
- `send` - Send data (defaults to POST if `with=` is present)
- `upload` - Upload a file (defaults to POST with multipart)
- `watch` - Watch a resource for changes (SSE or polling)
- `inspect` - Inspect headers/cookies (defaults to HEAD)
- `auth` - Manage authentication profiles
- `session` - Manage session cookie jars
- `profile` - Manage request profiles

### Clauses

- `with=<type>:<value>` - Request body (e.g., `with=json:'{"key":"value"}'`)
- `headers=<object>` - HTTP headers
- `params=<object>` - Query parameters
- `as=<format>` - Output format (json, csv, text, raw)
- `to=<path>` - Destination file or directory
- `method=<method>` - HTTP method override
- `retry=<count>` - Retry count
- `backoff=<min>..<max>` - Backoff range (e.g., `backoff=200ms..5s`)
- `timeout=<duration>` - Request timeout
- `proxy=<url>` - Proxy URL
- `pick=<jsonpath>` - JSONPath selector
- `every=<duration>` - Polling interval for watch
- `until=<predicate>` - Stop condition for watch
- `field=<name>=<value>` - Multipart form field
- `insecure` - Skip TLS verification
- `verbose` - Verbose output with timing
- `resume` - Resume partial downloads

## Examples

### Read JSON

```bash
req read https://api.example.com/users as=json
```

### Send JSON Data

```bash
req send https://api.example.com/users with=json:'{"name":"Ada","email":"ada@example.com"}'
```

### Save a File

```bash
req save https://example.com/file.zip to=file.zip
```

### With Headers

```bash
req read https://api.example.com/users headers='{"Authorization":"Bearer token"}' as=json
```

### With Retry and Timeout

```bash
req read https://api.example.com/users retry=3 timeout=10s as=json
```

### Dry Run

```bash
req read https://api.example.com/users as=json --dry-run
```

## Current Status

**v0.1.0** - Parser and planner with dry-run output

- ✅ Command parsing with grammar validation
- ✅ Execution plan generation
- ✅ Dry-run mode that prints JSON plans
- ✅ Helpful error messages with suggestions

## Roadmap

- **v0.2.0** - Read and save with headers, params, as, to, timeout, retry
- **v0.3.0** - Send with json and form, upload with multipart
- **v0.4.0** - Watch SSE and polling with until
- **v0.5.0** - Pick on JSON responses
- **v0.6.0** - Auth profiles and session cookie jars
- **v0.7.0** - Inspect and verbose
- **v0.8.0** - Cross platform polish and docs
- **v1.0.0** - Stability hardening and release candidates

## Contributing

Contributions are welcome! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) file for details.

