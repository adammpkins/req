# Verbs Reference

This document provides detailed documentation for each verb in `req`.

## Overview

Verbs are the action words that define what `req` should do. Each verb has:
- A default HTTP method
- Compatible clauses
- Specific behaviors
- Use cases

## Verb List

- [read](#read) - GET request, print to stdout
- [save](#save) - GET request, save to file
- [send](#send) - GET by default, POST if body provided
- [upload](#upload) - POST with multipart form data
- [watch](#watch) - GET with SSE or polling
- [inspect](#inspect) - HEAD request
- [authenticate](#authenticate) - Login and store session
- [session](#session) - Session management

## read

**Purpose**: Fetch a resource and print the response to stdout.

**Default Method**: `GET`

**Compatible Methods**: `GET`, `HEAD`, `OPTIONS`

**Use Cases**:
- Fetching API endpoints
- Reading JSON data
- Checking resource availability
- Retrieving content

### Default Behavior

- Follows redirects (up to 5)
- Prints response body to stdout
- Auto-detects output format if `as=` not specified

### Examples

```bash
# Simple GET request
req read https://api.example.com/users

# With JSON formatting
req read https://api.example.com/users as=json

# With query parameters
req read https://api.example.com/search \
  include='param: q=search query' \
  as=json

# With headers
req read https://api.example.com/users \
  include='header: Authorization: Bearer $TOKEN' \
  as=json

# Using HEAD instead of GET
req read https://api.example.com/users using=HEAD
```

### Common Patterns

**API Endpoint with Auth:**
```bash
req read https://api.example.com/users \
  include='header: Authorization: Bearer $TOKEN' \
  as=json
```

**Search with Multiple Params:**
```bash
req read https://api.example.com/search \
  include='param: q=test; param: page=1; param: limit=10' \
  as=json
```

## save

**Purpose**: Download a resource and save it to a file.

**Default Method**: `GET`

**Compatible Methods**: `GET`, `POST`

**Use Cases**:
- Downloading files
- Saving API responses to disk
- Archiving resources

### Default Behavior

- Follows redirects (up to 5)
- Writes response to file (stdout empty)
- Auto-extracts filename from URL if `to=` not specified
- Default output format: `raw`

### Examples

```bash
# Save with explicit filename
req save https://example.com/file.zip to=file.zip

# Auto-detect filename from URL
req save https://example.com/document.pdf

# Save to directory (filename extracted from URL)
req save https://example.com/file.zip to=/tmp/

# Save with POST request
req save https://api.example.com/export \
  using=POST \
  with='{"format":"csv"}' \
  to=export.csv
```

### Filename Extraction

If `to=` is not specified or points to a directory, the filename is extracted from the URL path. The filename is URL-decoded and cleaned.

Examples:
- `https://example.com/file.zip` → `file.zip`
- `https://example.com/docs/document.pdf` → `document.pdf`
- `https://example.com/` → `download`

## send

**Purpose**: Send data to a server. Flexible verb that adapts based on context.

**Default Method**: `GET` (changes to `POST` if `with=` is present)

**Compatible Methods**: `POST`, `PUT`, `PATCH`

**Use Cases**:
- Creating resources (POST)
- Updating resources (PUT/PATCH)
- Sending JSON data
- Form submissions

### Default Behavior

- Defaults to GET if no body
- Changes to POST if `with=` clause present
- Does not follow redirects by default (write verb)
- Can use `follow=smart` for safe redirects

### Examples

```bash
# GET request (no body)
req send https://api.example.com/users as=json

# POST with JSON body (method inferred)
req send https://api.example.com/users \
  with='{"name":"Alice","email":"alice@example.com"}' \
  as=json

# Explicit POST
req send https://api.example.com/users \
  using=POST \
  with='{"name":"Alice"}' \
  as=json

# PUT request
req send https://api.example.com/users/1 \
  using=PUT \
  with='{"name":"Bob"}' \
  as=json

# PATCH request
req send https://api.example.com/users/1 \
  using=PATCH \
  with='{"email":"new@example.com"}' \
  as=json

# With assertions
req send https://api.example.com/users \
  using=POST \
  with='{"name":"Alice"}' \
  expect=status:201, header:Content-Type=application/json \
  as=json
```

### Common Patterns

**Create Resource:**
```bash
req send https://api.example.com/users \
  using=POST \
  include='header: Authorization: Bearer $TOKEN' \
  with='{"name":"Alice"}' \
  expect=status:201 \
  as=json
```

**Update Resource:**
```bash
req send https://api.example.com/users/1 \
  using=PUT \
  include='header: Authorization: Bearer $TOKEN' \
  with='{"name":"Updated"}' \
  expect=status:200 \
  as=json
```

## upload

**Purpose**: Upload files using multipart/form-data.

**Default Method**: `POST`

**Compatible Methods**: `POST`, `PUT`

**Use Cases**:
- File uploads
- Form submissions with files
- Multipart form data

### Default Behavior

- Requires `attach=` or `with=` clause (error otherwise)
- Automatically sets `Content-Type: multipart/form-data`
- Overrides manual Content-Type header with a note

### Examples

```bash
# Single file upload
req upload https://api.example.com/upload \
  attach='part: name=file, file=@./avatar.png' \
  as=json

# File with metadata
req upload https://api.example.com/upload \
  attach='part: name=avatar, file=@./me.png, filename=avatar.png, type=image/png' \
  as=json

# Multiple parts (file + text)
req upload https://api.example.com/upload \
  attach='part: name=file, file=@./document.pdf; part: name=description, value=My document' \
  as=json

# With Content-Type override note
req upload https://api.example.com/upload \
  include='header: Content-Type: application/json' \
  attach='part: name=file, file=@test.png' \
  as=json
# stderr: Note: Content-Type overridden for multipart
```

### Common Patterns

**Image Upload with Metadata:**
```bash
req upload https://api.example.com/images \
  include='header: Authorization: Bearer $TOKEN' \
  attach='part: name=image, file=@./photo.jpg, type=image/jpeg; part: name=title, value=My Photo' \
  expect=status:201 \
  as=json
```

## watch

**Purpose**: Monitor a resource for changes using polling or Server-Sent Events (SSE).

**Default Method**: `GET`

**Compatible Methods**: `GET` only

**Use Cases**:
- Monitoring API endpoints
- Watching for updates
- Real-time data streams
- Polling for status changes

### Default Behavior

- Uses GET requests
- Output format depends on TTY detection:
  - TTY: Timestamped lines
  - Non-TTY: Raw lines

### Examples

```bash
# Watch endpoint (polling)
req watch https://api.example.com/status

# Watch with SSE
req watch https://api.example.com/stream

# Watch with formatting
req watch https://api.example.com/events as=json
```

**Note**: Watch mode implementation details may vary. See [Advanced Usage](ADVANCED.md) for more information.

## inspect

**Purpose**: Check resource headers without fetching the body.

**Default Method**: `HEAD`

**Compatible Methods**: `HEAD`, `GET`, `OPTIONS`

**Use Cases**:
- Checking if resource exists
- Inspecting headers
- Checking content type
- Verifying ETags

### Default Behavior

- Uses HEAD method
- Default output format: `json`
- Does not fetch response body

### Examples

```bash
# Check headers
req inspect https://api.example.com/users

# Check with GET (if HEAD not supported)
req inspect https://api.example.com/users using=GET

# Check specific headers
req inspect https://api.example.com/users \
  expect=header:Content-Type=application/json
```

### Common Patterns

**Check Resource Exists:**
```bash
req inspect https://api.example.com/users/1 \
  expect=status:200
```

**Verify Content-Type:**
```bash
req inspect https://api.example.com/document.pdf \
  expect=header:Content-Type=application/pdf
```

## authenticate

**Purpose**: Authenticate with a server and store session credentials.

**Default Method**: `POST` (if `with=` present), otherwise requires `using=`

**Compatible Methods**: Any (typically POST)

**Use Cases**:
- Login flows
- Obtaining API tokens
- Session establishment

### Default Behavior

- Follows redirects (to capture Set-Cookie headers)
- Captures Set-Cookie headers from redirect responses
- Extracts `access_token` from JSON response body
- Stores session per host in `~/.config/req/session_<host>.json`
- Session files have strict permissions (0600)

### Examples

```bash
# POST login
req authenticate https://api.example.com/login \
  using=POST \
  with='{"username":"user","password":"pass"}'

# GET login (with query params)
req authenticate https://api.example.com/login \
  using=GET \
  include='param: username=user; param: password=pass'

# Session automatically used for subsequent requests
req read https://api.example.com/me as=json
```

### Session Storage

Sessions are stored in:
- **Location**: `~/.config/req/session_<host>.json`
- **Permissions**: `0600` (owner read/write only)
- **Format**: JSON with cookies and authorization token

See [Session Management](SESSIONS.md) for detailed information.

## session

**Purpose**: Manage stored sessions.

**Subcommands**: `show`, `clear`, `use`

**Use Cases**:
- Viewing stored sessions
- Clearing sessions
- Exporting session for scripts

### Subcommands

#### session show

Display stored session information (redacted by default).

```bash
# Show session (redacted)
req session show api.example.com

# Show session as JSON
req session show api.example.com as=json
```

#### session clear

Delete a stored session.

```bash
req session clear api.example.com
```

#### session use

Print environment variable stub for shell scoping.

```bash
req session use api.example.com
# Output: export REQ_SESSION_HOST="api.example.com"
```

### Examples

```bash
# List all sessions (via show)
req session show api.example.com

# Clear session
req session clear api.example.com

# Use session in script
eval $(req session use api.example.com)
```

## Verb Method Compatibility

| Verb | Default Method | Compatible Methods |
|------|---------------|-------------------|
| `read` | GET | GET, HEAD, OPTIONS |
| `save` | GET | GET, POST |
| `send` | GET (POST if `with=`) | POST, PUT, PATCH |
| `upload` | POST | POST, PUT |
| `watch` | GET | GET |
| `inspect` | HEAD | HEAD, GET, OPTIONS |
| `authenticate` | POST (if `with=`) | Any |
| `session` | N/A | N/A |

**Note**: Using incompatible methods results in a validation error.

## See Also

- [Grammar Reference](GRAMMAR.md) - Complete grammar specification
- [Clauses Reference](CLAUSES.md) - Available clauses for each verb
- [Examples Cookbook](EXAMPLES.md) - Real-world examples
- [Session Management](SESSIONS.md) - Detailed session documentation

