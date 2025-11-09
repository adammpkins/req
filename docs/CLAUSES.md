# Clauses Reference

This document provides detailed documentation for each clause in `req`.

## Overview

Clauses are key=value pairs that modify the behavior of a `req` command. They can appear in any order and are optional (except where required by specific verbs).

## Clause Categories

- **Request Modification**: `using=`, `include=`, `with=`, `attach=`, `via=`, `insecure=`
- **Output Control**: `as=`, `to=`
- **Validation**: `expect=`
- **Behavior**: `follow=`, `retry=`, `under=`

## Request Modification Clauses

### using=

**Purpose**: Override the default HTTP method for the verb.

**Format**: `using=<method>`

**Values**: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD`, `OPTIONS`

**Repeatable**: No

**Validation**: Method must be compatible with the verb (see [Verbs Reference](VERBS.md))

**Examples**:
```bash
# Use PUT instead of POST
req send https://api.example.com/users/1 using=PUT with='{"name":"Updated"}'

# Use HEAD instead of GET
req read https://api.example.com/users using=HEAD

# Use PATCH for partial update
req send https://api.example.com/users/1 using=PATCH with='{"email":"new@example.com"}'
```

**Common Mistakes**:
- Using incompatible methods (e.g., `read using=POST`) results in validation error
- Method names are case-insensitive but normalized to uppercase

### include=

**Purpose**: Add headers, query parameters, cookies, or Basic Auth credentials.

**Format**: `include='<items>'`

**Repeatable**: Yes (multiple include clauses are merged)

**Item Types**:
- `header: Name: Value` - HTTP header
- `param: key=value` - Query parameter
- `cookie: key=value` - Cookie
- `basic: username:password` - Basic Auth (automatically encoded)

**Merging Rules**:
- **Headers**: Last value wins (except multi-valued headers keep all values)
- **Params**: Repeated keys become repeated query parameters in order
- **Cookies**: Last value wins per cookie name
- **Basic Auth**: Sets Authorization header, overrides existing

**Examples**:
```bash
# Single header
req read https://api.example.com/users \
  include='header: Authorization: Bearer $TOKEN' \
  as=json

# Multiple items in one clause
req read https://api.example.com/search \
  include='header: Accept: application/json; param: q=search; cookie: session=abc' \
  as=json

# Multiple include clauses
req read https://api.example.com/users \
  include='header: Accept: application/json' \
  include='header: X-Trace: 1' \
  include='param: page=1' \
  as=json

# Basic Auth
req read https://httpbin.org/basic-auth/user/passwd \
  include='basic: user:passwd' \
  expect=status:200

# Header with commas and q values (must be quoted)
req read https://api.example.com/search \
  include='header: Accept: application/json, application/problem+json; q=0.9' \
  as=json

# Cookie with semicolons (must be quoted)
req read https://api.example.com/search \
  include='cookie: prefs="a=1; b=2; c=3"' \
  as=json
```

**Edge Cases**:
- Values containing semicolons must be quoted
- Header values with commas and quality values need quoting
- Basic Auth automatically encodes credentials as Base64

**See Also**: [Authentication Guide](AUTHENTICATION.md) for auth examples

### with=

**Purpose**: Specify the request body.

**Format**: `with=<body>` or `with=@<file>` or `with=@-`

**Repeatable**: No

**Modes**:
1. **Inline**: `with='{"name":"Alice"}'` - Direct value
2. **File**: `with=@file.json` - Read from file
3. **Stdin**: `with=@-` - Read from stdin

**Content-Type Inference**:
- If Content-Type header is not set and inline value starts with `{` or `[`, infer `application/json`
- A note is printed to stderr when inference occurs
- Explicit Content-Type header always overrides inference

**Examples**:
```bash
# Inline JSON (inference occurs)
req send https://api.example.com/users \
  with='{"name":"Alice","email":"alice@example.com"}' \
  as=json
# stderr: Inferred Content-Type: application/json

# From file
req send https://api.example.com/users \
  with=@user.json \
  as=json

# From stdin
echo '{"name":"Bob"}' | req send https://api.example.com/users with=@- as=json

# Explicit Content-Type (no inference)
req send https://api.example.com/users \
  include='header: Content-Type: application/xml' \
  with='<user><name>Alice</name></user>' \
  as=json
```

**Common Mistakes**:
- File path must exist (error if not found)
- Stdin mode (`@-`) reads entire stdin until EOF

### attach=

**Purpose**: Add multipart/form-data parts for file uploads.

**Format**: `attach='<parts>'`

**Repeatable**: Yes (multiple attach clauses are combined)

**Part Format**:
```
part: name=<name>, file=@<path> | value=<value> [, filename=<filename>] [, type=<mime-type>]
```

**Required**:
- `name=` - Form field name

**Exactly one of**:
- `file=@<path>` - File path (must exist)
- `value=<value>` - Text value

**Optional**:
- `filename=` - Filename for file parts
- `type=` - MIME type

**Boundary**:
- `boundary: <token>` - Optional explicit boundary (rarely needed)

**Behavior**:
- Automatically sets `Content-Type: multipart/form-data` with generated boundary
- Overrides manual Content-Type header with a note

**Examples**:
```bash
# Single file
req upload https://api.example.com/upload \
  attach='part: name=file, file=@./avatar.png' \
  as=json

# File with filename and type
req upload https://api.example.com/upload \
  attach='part: name=avatar, file=@./me.png, filename=avatar.png, type=image/png' \
  as=json

# Text part
req upload https://api.example.com/upload \
  attach='part: name=description, value=My file description' \
  as=json

# Multiple parts
req upload https://api.example.com/upload \
  attach='part: name=file, file=@./document.pdf; part: name=meta, value={"title":"Doc"}' \
  as=json

# Content-Type override note
req upload https://api.example.com/upload \
  include='header: Content-Type: application/json' \
  attach='part: name=file, file=@test.png' \
  as=json
# stderr: Note: Content-Type overridden for multipart
```

**Validation Errors**:
- Missing `name=` → Error: "attach part missing name"
- Missing both `file=` and `value=` → Error: "attach part missing both file and value"
- Both `file=` and `value=` present → Error: "attach part cannot have both file and value"
- File not found → Error: "file not found: <path>"

### via=

**Purpose**: Specify a proxy server.

**Format**: `via=<proxy-url>`

**Repeatable**: No

**Values**: HTTP proxy URL (e.g., `http://proxy.example.com:8080`)

**Examples**:
```bash
# HTTP proxy
req read https://api.example.com/users via=http://proxy.example.com:8080 as=json

# HTTPS proxy
req read https://api.example.com/users via=https://proxy.example.com:8080 as=json
```

**Note**: Only HTTP proxies are currently supported. SOCKS proxies are not supported.

### insecure=

**Purpose**: Disable TLS certificate verification.

**Format**: `insecure=true` or `insecure=false`

**Repeatable**: No

**Default**: `false` (verification enabled)

**Security Warning**: A warning is printed to stderr when TLS verification is disabled.

**Examples**:
```bash
# Disable TLS verification (for self-signed certificates)
req read https://self-signed.example.com insecure=true as=json
# stderr: Warning: TLS verification disabled
```

**Use Cases**:
- Testing with self-signed certificates
- Internal development environments
- Troubleshooting TLS issues

**Security**: Never use `insecure=true` in production or with sensitive data.

## Output Control Clauses

### as=

**Purpose**: Specify output format for stdout.

**Format**: `as=<format>`

**Repeatable**: No

**Values**: `json`, `csv`, `text`, `raw`, `auto`

**Formats**:
- `json` - Pretty-printed JSON
- `csv` - CSV format (if applicable)
- `text` - Plain text
- `raw` - Raw response body
- `auto` - Auto-detect based on Content-Type

**Examples**:
```bash
# JSON output
req read https://api.example.com/users as=json

# Raw output
req read https://api.example.com/users as=raw

# Auto-detect
req read https://api.example.com/users as=auto
```

**Default by Verb**:
- `read`: `auto`
- `save`: `raw` (writes to file, stdout empty)
- `send`: `auto`
- `upload`: `auto`
- `watch`: `auto`
- `inspect`: `json`

### to=

**Purpose**: Specify destination path for file output.

**Format**: `to=<path>`

**Repeatable**: No

**Values**: File path or directory path

**Behavior**:
- If path is a directory, filename is extracted from URL
- If path is a file, uses that filename
- If not specified for `save` verb, filename extracted from URL

**Examples**:
```bash
# Explicit filename
req save https://example.com/file.zip to=file.zip

# Directory (filename from URL)
req save https://example.com/document.pdf to=/tmp/

# Full path
req save https://example.com/file.zip to=/tmp/archive.zip
```

## Validation Clauses

### expect=

**Purpose**: Assert conditions on the response.

**Format**: `expect=<checks>`

**Repeatable**: No

**Check Types**:
- `status:<code>` - HTTP status code
- `header:<name>=<value>` - Header value
- `contains:"<text>"` - Body contains text
- `jsonpath:"<path>"` - JSONPath expression matches
- `matches:"<regex>"` - Regex pattern matches

**Exit Code**: 3 if any check fails

**Examples**:
```bash
# Status check
req read https://api.example.com/users expect=status:200 as=json

# Multiple checks
req send https://api.example.com/users \
  using=POST \
  with='{"name":"Alice"}' \
  expect=status:201, header:Content-Type=application/json, contains:"id" \
  as=json

# JSONPath check
req read https://api.example.com/users \
  expect=jsonpath:"$.items[0].id" \
  as=json

# Regex match
req read https://api.example.com/status \
  expect=matches:"^OK\\b" \
  as=text
```

**Common Patterns**:
- All checks must pass (AND logic)
- Failure messages are concise and specific
- Exit code 3 indicates expectation failure

**See Also**: [Error Handling](ERRORS.md) for exit codes

## Behavior Clauses

### follow=

**Purpose**: Control redirect following behavior.

**Format**: `follow=smart`

**Repeatable**: No

**Values**: `smart` (only value currently supported)

**Default Behavior**:
- `read` and `save`: Follow up to 5 redirects
- `send`, `upload`: Do not follow redirects
- `authenticate`: Follows redirects (to capture Set-Cookie)

**Smart Follow**:
- For write verbs: Only follows 301, 302, 303, 307, 308 redirects
- For read verbs: Follows all redirects (up to 5)
- Prints advisory for 301/302/303 on write verbs if not following

**Examples**:
```bash
# Smart follow for write verb (only 307/308)
req send https://api.example.com/create \
  using=POST \
  with='{"data":"value"}' \
  follow=smart

# Read follows by default
req read https://example.com/redirect
```

**See Also**: [Advanced Usage](ADVANCED.md) for redirect details

### retry=

**Purpose**: Retry failed requests.

**Format**: `retry=<count>`

**Repeatable**: No

**Values**: Positive integer (number of retry attempts)

**Behavior**:
- Retries on transient errors
- Does not retry on expectation failures (exit code 3)
- Does not retry on grammar errors (exit code 5)

**Examples**:
```bash
# Retry 3 times
req read https://api.example.com/users retry=3 as=json

# With timeout
req read https://api.example.com/users retry=3 under=10s as=json
```

**Note**: Retry implementation details may vary. See [Advanced Usage](ADVANCED.md).

### under=

**Purpose**: Set timeout or size limit.

**Format**: `under=<duration>` or `under=<size>`

**Repeatable**: No

**Duration Format**: `<number><unit>` where unit is `s`, `m`, or `h`
- Examples: `30s`, `5m`, `1h`

**Size Format**: `<number><unit>` where unit is `B`, `KB`, `MB`, or `GB`
- Examples: `10MB`, `1GB`, `500KB`

**Examples**:
```bash
# Timeout
req read https://api.example.com/users under=30s as=json

# Size limit
req save https://example.com/large-file.zip under=100MB to=file.zip

# Both timeout and retry
req read https://api.example.com/users retry=3 under=10s as=json
```

**Errors**:
- Timeout exceeded → Exit code 4
- Size limit exceeded → Exit code 4

## Clause Precedence and Ordering

Clauses can appear in any order. The following are equivalent:

```bash
req read https://api.example.com/users as=json include='header: Accept: application/json'

req read https://api.example.com/users include='header: Accept: application/json' as=json
```

## Clause Conflicts

### Duplicate Singletons

Singleton clauses (like `as=`, `with=`, `expect=`) cannot appear twice:

```bash
# Error: duplicate singleton clause
req read https://api.example.com/users as=json as=text
```

### include= Override Behavior

Explicit `include=` for Authorization or Cookie headers overrides session auto-application:

```bash
# Session exists for api.example.com
# This will use explicit header, not session
req read https://api.example.com/users \
  include='header: Authorization: Bearer explicit-token' \
  as=json
```

## See Also

- [Grammar Reference](GRAMMAR.md) - Complete grammar specification
- [Verbs Reference](VERBS.md) - Verb-specific clause compatibility
- [Examples Cookbook](EXAMPLES.md) - Real-world clause usage
- [Error Handling](ERRORS.md) - Clause-related errors

