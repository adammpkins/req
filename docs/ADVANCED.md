# Advanced Usage

This document covers advanced usage patterns, tips, and tricks for `req`.

## Advanced Patterns

### Chaining Requests

Chain multiple requests where the output of one feeds into the next:

```bash
#!/bin/bash
# Create user and get ID
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

### Conditional Execution

Execute requests conditionally based on previous results:

```bash
#!/bin/bash
# Check if resource exists
if req read https://api.example.com/users/1 \
  expect=status:200 > /dev/null 2>&1; then
  # Update existing
  req send https://api.example.com/users/1 \
    using=PUT \
    include="header: Authorization: Bearer $TOKEN" \
    with='{"name":"Updated"}' \
    as=json
else
  # Create new
  req send https://api.example.com/users \
    using=POST \
    include="header: Authorization: Bearer $TOKEN" \
    with='{"name":"New User"}' \
    as=json
fi
```

### Polling for Status

Poll an endpoint until a condition is met:

```bash
#!/bin/bash
max_attempts=30
attempt=0
interval=2

while [ $attempt -lt $max_attempts ]; do
  if req read https://api.example.com/jobs/123 \
    include="header: Authorization: Bearer $TOKEN" \
    expect=contains:"completed" \
    as=json > /dev/null 2>&1; then
    echo "Job completed"
    break
  fi
  sleep $interval
  attempt=$((attempt + 1))
  echo "Attempt $attempt/$max_attempts..."
done

if [ $attempt -eq $max_attempts ]; then
  echo "Job did not complete in time"
  exit 1
fi
```

### Rate Limiting

Implement rate limiting in scripts:

```bash
#!/bin/bash
rate_limit=1  # seconds between requests

for item in {1..100}; do
  req read "https://api.example.com/items/$item" \
    include="header: Authorization: Bearer $TOKEN" \
    as=json
  sleep $rate_limit
done
```

### Batch Operations

Process multiple items:

```bash
#!/bin/bash
items=("item1" "item2" "item3")

for item in "${items[@]}"; do
  req send https://api.example.com/items \
    using=POST \
    include="header: Authorization: Bearer $TOKEN" \
    with="{\"name\":\"$item\"}" \
    expect=status:201 \
    as=json
done
```

## Scripting Patterns

### Error Handling

Comprehensive error handling:

```bash
#!/bin/bash
set -euo pipefail

response=$(req read https://api.example.com/users \
  include="header: Authorization: Bearer $TOKEN" \
  expect=status:200 \
  as=json 2>&1)
exit_code=$?

case $exit_code in
  0)
    echo "Success: $response"
    ;;
  3)
    echo "Expectation failed: $response" >&2
    exit 1
    ;;
  4)
    echo "Network error: $response" >&2
    exit 1
    ;;
  5)
    echo "Parse error: $response" >&2
    exit 1
    ;;
  *)
    echo "Unknown error: $response" >&2
    exit 1
    ;;
esac
```

### Function Wrappers

Create reusable functions:

```bash
#!/bin/bash

api_get() {
  local endpoint=$1
  req read "https://api.example.com/$endpoint" \
    include="header: Authorization: Bearer $TOKEN" \
    expect=status:200 \
    as=json
}

api_post() {
  local endpoint=$1
  local data=$2
  req send "https://api.example.com/$endpoint" \
    using=POST \
    include="header: Authorization: Bearer $TOKEN" \
    with="$data" \
    expect=status:201 \
    as=json
}

# Usage
users=$(api_get "users")
new_user=$(api_post "users" '{"name":"Alice"}')
```

### Configuration Files

Store configuration in files:

```bash
#!/bin/bash
# config.sh
export API_BASE_URL="https://api.example.com"
export API_TOKEN="your-token"
export API_TIMEOUT="30s"

# script.sh
source config.sh

req read "$API_BASE_URL/users" \
  include="header: Authorization: Bearer $API_TOKEN" \
  under="$API_TIMEOUT" \
  as=json
```

## Integration with Other Tools

### jq Integration

Process JSON responses with `jq`:

```bash
# Get user IDs
req read https://api.example.com/users \
  include="header: Authorization: Bearer $TOKEN" \
  as=json | jq -r '.[] | .id'

# Filter and transform
req read https://api.example.com/users \
  include="header: Authorization: Bearer $TOKEN" \
  as=json | jq '[.[] | select(.active == true) | {id, name}]'
```

### grep Integration

Search in responses:

```bash
# Find lines containing "error"
req read https://api.example.com/logs \
  include="header: Authorization: Bearer $TOKEN" \
  as=text | grep -i error
```

### awk Integration

Process text responses:

```bash
# Extract specific fields
req read https://api.example.com/data \
  include="header: Authorization: Bearer $TOKEN" \
  as=csv | awk -F',' '{print $1, $3}'
```

### xargs Integration

Process multiple URLs:

```bash
# Process multiple endpoints
cat urls.txt | xargs -I {} req read {} \
  include="header: Authorization: Bearer $TOKEN" \
  as=json
```

## Performance Optimization

### Parallel Requests

Use `xargs -P` for parallel execution:

```bash
# Process 10 items in parallel
seq 1 100 | xargs -P 10 -I {} req read "https://api.example.com/items/{}" \
  include="header: Authorization: Bearer $TOKEN" \
  as=json
```

### Connection Reuse

Sessions help with connection reuse:

```bash
# Authenticate once
req authenticate https://api.example.com/login \
  using=POST \
  with='{"user":"user","pass":"pass"}'

# Subsequent requests reuse connection
for i in {1..100}; do
  req read "https://api.example.com/items/$i" as=json
done
```

### Timeout Tuning

Adjust timeouts for your use case:

```bash
# Fast API - short timeout
req read https://fast-api.example.com/users under=5s as=json

# Slow API - longer timeout
req read https://slow-api.example.com/users under=60s as=json
```

## Custom Output Formatting

### JSON Processing

Use `jq` for custom JSON formatting:

```bash
# Pretty print with custom format
req read https://api.example.com/users \
  include="header: Authorization: Bearer $TOKEN" \
  as=json | jq -r '.[] | "\(.id): \(.name)"'
```

### CSV Processing

Convert JSON to CSV:

```bash
# Convert to CSV
req read https://api.example.com/users \
  include="header: Authorization: Bearer $TOKEN" \
  as=json | jq -r '.[] | [.id, .name, .email] | @csv'
```

### Text Processing

Extract specific text:

```bash
# Extract status
req read https://api.example.com/status \
  include="header: Authorization: Bearer $TOKEN" \
  as=text | grep -oP 'status:\s*\K\w+'
```

## Environment Variable Patterns

### Per-Environment Configuration

```bash
#!/bin/bash
# Set environment
ENV=${1:-dev}

case $ENV in
  dev)
    export API_URL="https://dev-api.example.com"
    export API_TOKEN="$DEV_TOKEN"
    ;;
  staging)
    export API_URL="https://staging-api.example.com"
    export API_TOKEN="$STAGING_TOKEN"
    ;;
  prod)
    export API_URL="https://api.example.com"
    export API_TOKEN="$PROD_TOKEN"
    ;;
esac

req read "$API_URL/users" \
  include="header: Authorization: Bearer $API_TOKEN" \
  as=json
```

### Secret Management Integration

```bash
#!/bin/bash
# Get token from vault
export API_TOKEN=$(vault kv get -field=token secret/api)

# Use token
req read https://api.example.com/users \
  include="header: Authorization: Bearer $API_TOKEN" \
  as=json
```

## Redirect Handling

### Smart Redirects for Writes

Use `follow=smart` for write verbs to safely follow redirects:

```bash
# Only follows 307/308, not 301/302/303
req send https://api.example.com/create \
  using=POST \
  with='{"data":"value"}' \
  follow=smart \
  expect=status:200
```

### Redirect Advisory

For write verbs without `follow=smart`, advisories are printed:

```bash
req send https://api.example.com/create \
  using=POST \
  with='{"data":"value"}'
# If 303 redirect: Advisory: 303 redirect for write verb, not following
```

## Compression Handling

### Automatic Decompression

`req` automatically handles gzip and Brotli compression:

```bash
# Compression handled automatically
req read https://api.example.com/users \
  include="header: Authorization: Bearer $TOKEN" \
  as=json
# stderr: Decompressed response (if compressed)
```

### Custom Accept-Encoding

Override default compression:

```bash
# Only accept gzip
req read https://api.example.com/users \
  include='header: Accept-Encoding: gzip' \
  as=json

# No compression
req read https://api.example.com/users \
  include='header: Accept-Encoding: identity' \
  as=json
```

## Watch Mode Patterns

### Polling Pattern

```bash
#!/bin/bash
# Poll endpoint every 5 seconds
while true; do
  req watch https://api.example.com/events \
    expect=contains:"update" \
    as=json
  sleep 5
done
```

### SSE Pattern

```bash
# Server-Sent Events (if supported)
req watch https://api.example.com/stream \
  as=json
```

## Debugging Techniques

### Dry Run

Use `req explain` to see parsed command:

```bash
req explain "read https://api.example.com/users include='header: Authorization: Bearer token'"
```

### Verbose Output

Enable verbose mode:

```bash
req read https://api.example.com/users verbose as=json
```

### Stderr Inspection

Check stderr for details:

```bash
req read https://api.example.com/users as=json 2> debug.log
cat debug.log
```

## See Also

- [Examples Cookbook](EXAMPLES.md) - More examples
- [Grammar Reference](GRAMMAR.md) - Syntax details
- [Error Handling](ERRORS.md) - Error patterns
- [Session Management](SESSIONS.md) - Session patterns

