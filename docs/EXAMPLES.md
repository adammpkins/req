# Examples Cookbook

This document provides comprehensive examples organized by use case.

## Table of Contents

- [Basic Requests](#basic-requests)
- [Authentication](#authentication)
- [File Operations](#file-operations)
- [Multipart Uploads](#multipart-uploads)
- [API Testing](#api-testing)
- [Scripting Patterns](#scripting-patterns)
- [CI/CD Integration](#cicd-integration)
- [Error Handling](#error-handling)

## Basic Requests

### Simple GET Request

```bash
req read https://api.example.com/users
```

### GET with JSON Formatting

```bash
req read https://api.example.com/users as=json
```

### GET with Query Parameters

```bash
req read https://api.example.com/search \
  include='param: q=search query; param: page=1' \
  as=json
```

### GET with Headers

```bash
req read https://api.example.com/users \
  include='header: Accept: application/json; header: X-Trace: 1' \
  as=json
```

### POST Request

```bash
req send https://api.example.com/users \
  with='{"name":"Alice","email":"alice@example.com"}' \
  as=json
```

### PUT Request

```bash
req send https://api.example.com/users/1 \
  using=PUT \
  with='{"name":"Bob","email":"bob@example.com"}' \
  as=json
```

### PATCH Request

```bash
req send https://api.example.com/users/1 \
  using=PATCH \
  with='{"email":"newemail@example.com"}' \
  as=json
```

### DELETE Request

```bash
req send https://api.example.com/users/1 using=DELETE
```

## Authentication

### Basic Auth

```bash
req read https://httpbin.org/basic-auth/user/passwd \
  include='basic: user:passwd' \
  expect=status:200 \
  as=json
```

### Bearer Token

```bash
TOKEN="your-token-here"
req read https://api.example.com/users \
  include="header: Authorization: Bearer $TOKEN" \
  as=json
```

### Session-Based Auth

```bash
# Authenticate and store session
req authenticate https://api.example.com/login \
  using=POST \
  with='{"username":"user","password":"pass"}'

# Session automatically used
req read https://api.example.com/me as=json

# Explicit override
req read https://api.example.com/me \
  include='header: Authorization: Bearer different-token' \
  as=json
```

### Multiple Auth Methods

```bash
# Basic Auth + Bearer Token (Bearer wins)
req read https://api.example.com/users \
  include='basic: user:pass; header: Authorization: Bearer token' \
  as=json
```

## File Operations

### Download File

```bash
req save https://example.com/file.zip to=file.zip
```

### Download with Auto-Detected Filename

```bash
req save https://example.com/document.pdf
```

### Download to Directory

```bash
req save https://example.com/file.zip to=/tmp/
```

### Upload File

```bash
req upload https://api.example.com/upload \
  attach='part: name=file, file=@./document.pdf' \
  as=json
```

### Upload with Metadata

```bash
req upload https://api.example.com/upload \
  attach='part: name=file, file=@./photo.jpg, filename=photo.jpg, type=image/jpeg; part: name=title, value=My Photo' \
  as=json
```

### Read from File

```bash
req send https://api.example.com/users \
  with=@user.json \
  as=json
```

### Read from Stdin

```bash
echo '{"name":"Alice"}' | req send https://api.example.com/users with=@- as=json
```

## Multipart Uploads

### Single File Upload

```bash
req upload https://api.example.com/upload \
  attach='part: name=avatar, file=@./avatar.png' \
  as=json
```

### Multiple Files

```bash
req upload https://api.example.com/upload \
  attach='part: name=file1, file=@./file1.pdf; part: name=file2, file=@./file2.pdf' \
  as=json
```

### File and Text Parts

```bash
req upload https://api.example.com/upload \
  attach='part: name=image, file=@./photo.jpg; part: name=description, value=My photo description' \
  as=json
```

### File with JSON Metadata

```bash
req upload https://api.example.com/upload \
  attach='part: name=file, file=@./document.pdf; part: name=meta, value={"title":"Document","tags":["important"]}' \
  as=json
```

## API Testing

### Test Status Code

```bash
req read https://api.example.com/users \
  expect=status:200 \
  as=json
```

### Test Multiple Conditions

```bash
req send https://api.example.com/users \
  using=POST \
  with='{"name":"Alice"}' \
  expect=status:201, header:Content-Type=application/json, contains:"id" \
  as=json
```

### Test JSON Structure

```bash
req read https://api.example.com/users \
  expect=jsonpath:"$.items[0].id" \
  as=json
```

### Test Response Content

```bash
req read https://api.example.com/status \
  expect=contains:"success" \
  as=text
```

### Test with Regex

```bash
req read https://api.example.com/status \
  expect=matches:"^OK\\b" \
  as=text
```

## Scripting Patterns

### Check API Health

```bash
#!/bin/bash
if req read https://api.example.com/health expect=status:200 > /dev/null 2>&1; then
  echo "API is healthy"
else
  echo "API is down"
  exit 1
fi
```

### Create Resource and Get ID

```bash
#!/bin/bash
response=$(req send https://api.example.com/users \
  using=POST \
  include="header: Authorization: Bearer $TOKEN" \
  with='{"name":"Alice"}' \
  expect=status:201 \
  as=json)

user_id=$(echo "$response" | jq -r '.id')
echo "Created user with ID: $user_id"
```

### Batch Operations

```bash
#!/bin/bash
for name in Alice Bob Charlie; do
  req send https://api.example.com/users \
    using=POST \
    include="header: Authorization: Bearer $TOKEN" \
    with="{\"name\":\"$name\"}" \
    expect=status:201 \
    as=json
done
```

### Conditional Execution

```bash
#!/bin/bash
if req read https://api.example.com/users/1 expect=status:200 > /dev/null 2>&1; then
  echo "User exists, updating..."
  req send https://api.example.com/users/1 \
    using=PUT \
    include="header: Authorization: Bearer $TOKEN" \
    with='{"name":"Updated"}' \
    as=json
else
  echo "User not found, creating..."
  req send https://api.example.com/users \
    using=POST \
    include="header: Authorization: Bearer $TOKEN" \
    with='{"name":"New User"}' \
    as=json
fi
```

### Error Handling in Scripts

```bash
#!/bin/bash
response=$(req read https://api.example.com/users as=json)
exit_code=$?

if [ $exit_code -eq 0 ]; then
  echo "Success: $response"
elif [ $exit_code -eq 3 ]; then
  echo "Expectation failed"
  exit 1
elif [ $exit_code -eq 4 ]; then
  echo "Network error"
  exit 1
elif [ $exit_code -eq 5 ]; then
  echo "Parse error"
  exit 1
fi
```

## CI/CD Integration

### GitHub Actions

```yaml
name: API Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Install req
        run: go install github.com/adammpkins/req/cmd/req@latest
      
      - name: Test API endpoint
        env:
          API_TOKEN: ${{ secrets.API_TOKEN }}
        run: |
          req read https://api.example.com/users \
            include="header: Authorization: Bearer $API_TOKEN" \
            expect=status:200, header:Content-Type=application/json \
            as=json
```

### GitLab CI

```yaml
test_api:
  script:
    - go install github.com/adammpkins/req/cmd/req@latest
    - req read https://api.example.com/users
          include="header: Authorization: Bearer $API_TOKEN"
          expect=status:200
          as=json
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any
    stages {
        stage('Test API') {
            steps {
                sh '''
                    go install github.com/adammpkins/req/cmd/req@latest
                    req read https://api.example.com/users \\
                      include="header: Authorization: Bearer ${API_TOKEN}" \\
                      expect=status:200 \\
                      as=json
                '''
            }
        }
    }
}
```

## Error Handling

### Retry on Failure

```bash
req read https://api.example.com/users \
  retry=3 \
  under=10s \
  as=json
```

### Handle Timeout

```bash
if ! req read https://api.example.com/users under=5s as=json; then
  echo "Request timed out"
  exit 1
fi
```

### Validate Response

```bash
req send https://api.example.com/users \
  using=POST \
  with='{"name":"Alice"}' \
  expect=status:201, header:Content-Type=application/json, contains:"id" \
  as=json || {
    echo "Validation failed"
    exit 1
  }
```

### Check Exit Codes

```bash
req read https://api.example.com/users as=json
case $? in
  0) echo "Success" ;;
  3) echo "Expectation failed" ;;
  4) echo "Network error" ;;
  5) echo "Parse error" ;;
esac
```

## Advanced Patterns

### Chaining Requests

```bash
#!/bin/bash
# Create user
create_response=$(req send https://api.example.com/users \
  using=POST \
  include="header: Authorization: Bearer $TOKEN" \
  with='{"name":"Alice"}' \
  expect=status:201 \
  as=json)

user_id=$(echo "$create_response" | jq -r '.id')

# Get created user
req read "https://api.example.com/users/$user_id" \
  include="header: Authorization: Bearer $TOKEN" \
  expect=status:200 \
  as=json
```

### Polling for Status

```bash
#!/bin/bash
max_attempts=10
attempt=0

while [ $attempt -lt $max_attempts ]; do
  if req read https://api.example.com/jobs/123 \
    include="header: Authorization: Bearer $TOKEN" \
    expect=contains:"completed" \
    as=json > /dev/null 2>&1; then
    echo "Job completed"
    break
  fi
  sleep 2
  attempt=$((attempt + 1))
done
```

### Rate Limiting

```bash
#!/bin/bash
for item in {1..100}; do
  req read "https://api.example.com/items/$item" \
    include="header: Authorization: Bearer $TOKEN" \
    as=json
  sleep 1  # Rate limit: 1 request per second
done
```

## See Also

- [Verbs Reference](VERBS.md) - Verb-specific examples
- [Clauses Reference](CLAUSES.md) - Clause usage examples
- [Advanced Usage](ADVANCED.md) - More advanced patterns
- [Error Handling](ERRORS.md) - Error handling patterns

