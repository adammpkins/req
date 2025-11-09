# Error Handling

This document describes error handling, exit codes, and troubleshooting in `req`.

## Exit Codes

`req` uses specific exit codes to indicate different types of outcomes:

| Exit Code | Meaning | When It Occurs |
|-----------|---------|----------------|
| 0 | Success | Request completed successfully, all expectations passed |
| 3 | Expectation Failed | Request succeeded but an expectation check failed |
| 4 | Network Error | Network failure, timeout, TLS error, or HTTP error |
| 5 | Grammar/Parse Error | Command parsing error or validation failure |

## Exit Code 0: Success

The request completed successfully and all expectations (if any) passed.

```bash
req read https://api.example.com/users \
  expect=status:200 \
  as=json
# Exit code: 0
```

## Exit Code 3: Expectation Failed

The request completed successfully, but one or more expectation checks failed.

### Common Causes

- Status code mismatch
- Header value mismatch
- Content doesn't contain expected text
- JSONPath doesn't match
- Regex doesn't match

### Examples

```bash
# Status code mismatch
req read https://api.example.com/users \
  expect=status:200 \
  as=json
# HTTP 404
# expected status 200, got 404
# Error: expectation failed
# Exit code: 3

# Header mismatch
req read https://api.example.com/users \
  expect=header:Content-Type=application/json \
  as=json
# Content-Type: text/html
# expected header Content-Type=application/json, got text/html
# Error: expectation failed
# Exit code: 3

# Content check failure
req read https://api.example.com/status \
  expect=contains:"success" \
  as=text
# Response: "error"
# expected body to contain "success"
# Error: expectation failed
# Exit code: 3
```

### Handling in Scripts

```bash
#!/bin/bash
if req read https://api.example.com/users \
  expect=status:200 \
  as=json; then
  echo "Success"
else
  exit_code=$?
  if [ $exit_code -eq 3 ]; then
    echo "Expectation failed"
  fi
  exit $exit_code
fi
```

## Exit Code 4: Network Error

A network-related error occurred during request execution.

### Common Causes

- Connection timeout
- DNS resolution failure
- TLS/SSL certificate error
- HTTP error status (when no expectations)
- Size limit exceeded
- Proxy connection failure

### Examples

```bash
# Connection timeout
req read https://slow-api.example.com/users under=1s
# Error: request failed: context deadline exceeded
# Exit code: 4

# TLS certificate error
req read https://self-signed.example.com/users
# Error: request failed: x509: certificate signed by unknown authority
# Exit code: 4

# HTTP error (no expectations)
req read https://api.example.com/notfound
# HTTP 404
# Error: HTTP 404 Not Found
# Exit code: 4

# Size limit exceeded
req save https://example.com/large-file.zip under=1MB
# Error: size limit exceeded
# Exit code: 4
```

### Handling TLS Errors

For self-signed certificates, use `insecure=true`:

```bash
req read https://self-signed.example.com/users insecure=true as=json
# Warning: TLS verification disabled
```

**Security Warning**: Never use `insecure=true` in production.

### Handling Timeouts

Increase timeout or add retries:

```bash
# Increase timeout
req read https://slow-api.example.com/users under=30s as=json

# Add retries
req read https://unreliable-api.example.com/users retry=3 under=10s as=json
```

## Exit Code 5: Grammar/Parse Error

A command parsing or validation error occurred.

### Common Causes

- Unknown clause key
- Duplicate singleton clause
- Malformed include item
- Invalid URL
- File not found
- Method-verb incompatibility

### Examples

```bash
# Unknown clause
req read https://api.example.com/users invalid=clause
# Error: parse error at position 2 (token: "invalid"): unknown clause
# Exit code: 5

# Duplicate singleton
req read https://api.example.com/users as=json as=text
# Error: parse error at position 3 (token: "as"): duplicate singleton clause 'as'
# Exit code: 5

# Malformed include
req read https://api.example.com/users include='header: Invalid'
# Error: parse error at position 2 (token: "include"): header item missing Name: Value format
# Exit code: 5

# Invalid URL
req read not-a-url
# Error: parse error at position 1 (token: "not-a-url"): expected URL or host
# Exit code: 5

# File not found
req send https://api.example.com/users with=@nonexistent.json
# Error: file not found: nonexistent.json
# Exit code: 5

# Method-verb incompatibility
req read https://api.example.com/users using=POST
# Error: verb 'read' is incompatible with method 'POST'
# Exit code: 5
```

### Error Messages with Suggestions

`req` provides suggestions for common typos:

```bash
req read https://api.example.com/users incldue='header: Accept: application/json'
# Error: parse error at position 2 (token: "incldue"): unknown clause (did you mean "include"?)
```

## Error Message Format

Error messages follow this format:

```
Error: <error type>: <details>
```

### Parse Errors

```
parse error at position <N> (token: "<token>"): <message> [suggestion]
```

- **Position**: Token position in command
- **Token**: The problematic token
- **Message**: Description of the error
- **Suggestion**: Optional suggestion for correction

### Execution Errors

```
<error type>: <details>
```

Examples:
- `request failed: <network error>`
- `HTTP <code> <status>`
- `expectation failed`
- `file not found: <path>`

## Common Errors and Solutions

### "unknown clause"

**Problem**: Clause name is misspelled or doesn't exist.

**Solution**: Check spelling, use `req help` to see valid clauses.

```bash
# Wrong
req read https://api.example.com/users incldue='header: Accept: application/json'

# Correct
req read https://api.example.com/users include='header: Accept: application/json'
```

### "duplicate singleton clause"

**Problem**: A singleton clause appears twice.

**Solution**: Remove duplicate clause.

```bash
# Wrong
req read https://api.example.com/users as=json as=text

# Correct
req read https://api.example.com/users as=json
```

### "header item missing Name: Value format"

**Problem**: Header item doesn't have the required format.

**Solution**: Use `header: Name: Value` format.

```bash
# Wrong
req read https://api.example.com/users include='header: Invalid'

# Correct
req read https://api.example.com/users include='header: Accept: application/json'
```

### "basic item must be in format username:password"

**Problem**: Basic auth item missing colon separator.

**Solution**: Use `basic: username:password` format.

```bash
# Wrong
req read https://api.example.com/users include='basic: userpass'

# Correct
req read https://api.example.com/users include='basic: user:pass'
```

### "file not found"

**Problem**: File path in `with=` or `attach=` doesn't exist.

**Solution**: Verify file path is correct.

```bash
# Check file exists
ls -l ./file.json

# Use correct path
req send https://api.example.com/users with=@./file.json
```

### "verb 'X' is incompatible with method 'Y'"

**Problem**: HTTP method is not compatible with verb.

**Solution**: Use compatible method or different verb.

```bash
# Wrong
req read https://api.example.com/users using=POST

# Correct (use send for POST)
req send https://api.example.com/users using=POST with='{"data":"value"}'
```

### "expectation failed"

**Problem**: An expectation check failed.

**Solution**: Verify expected values match actual response.

```bash
# Check actual response
req read https://api.example.com/users as=json

# Adjust expectation
req read https://api.example.com/users \
  expect=status:200 \
  as=json
```

## Debugging Tips

### 1. Use `req explain`

See how your command is parsed without executing:

```bash
req explain "read https://api.example.com/users include='header: Accept: application/json'"
```

### 2. Check Exit Codes

```bash
req read https://api.example.com/users as=json
echo "Exit code: $?"
```

### 3. Verbose Output

Some errors provide additional context in stderr. Check stderr for details:

```bash
req read https://api.example.com/users as=json 2>&1
```

### 4. Test Components Separately

```bash
# Test URL parsing
req read https://api.example.com/users

# Test with headers
req read https://api.example.com/users include='header: Accept: application/json'

# Test with expectations
req read https://api.example.com/users expect=status:200
```

### 5. Verify Grammar

Check the [Grammar Reference](GRAMMAR.md) for correct syntax.

## Error Handling in Scripts

### Basic Error Handling

```bash
#!/bin/bash
set -e  # Exit on error

req read https://api.example.com/users \
  expect=status:200 \
  as=json
```

### Advanced Error Handling

```bash
#!/bin/bash
response=$(req read https://api.example.com/users as=json 2>&1)
exit_code=$?

if [ $exit_code -eq 0 ]; then
  echo "Success: $response"
elif [ $exit_code -eq 3 ]; then
  echo "Expectation failed: $response" >&2
  exit 1
elif [ $exit_code -eq 4 ]; then
  echo "Network error: $response" >&2
  exit 1
elif [ $exit_code -eq 5 ]; then
  echo "Parse error: $response" >&2
  exit 1
fi
```

### Retry Logic

```bash
#!/bin/bash
max_attempts=3
attempt=0

while [ $attempt -lt $max_attempts ]; do
  if req read https://api.example.com/users \
    retry=1 \
    under=10s \
    expect=status:200 \
    as=json > /dev/null 2>&1; then
    echo "Success"
    break
  fi
  attempt=$((attempt + 1))
  sleep 1
done

if [ $attempt -eq $max_attempts ]; then
  echo "Failed after $max_attempts attempts"
  exit 1
fi
```

## See Also

- [Grammar Reference](GRAMMAR.md) - Grammar and validation rules
- [Clauses Reference](CLAUSES.md) - Clause syntax and errors
- [Security Best Practices](SECURITY.md) - Security-related errors
- [Examples Cookbook](EXAMPLES.md) - Error handling examples

