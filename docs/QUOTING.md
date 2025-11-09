# Cross Shell Quoting Guide

This guide provides examples of how to properly quote `req` commands in different shells.

## General Rules

- Values containing semicolons (`;`) must be quoted
- Values containing spaces should be quoted
- Values containing special characters should be quoted
- Environment variables are expanded by the shell before being passed to `req`

## Bash / Zsh

### Single Quotes (Recommended)
Single quotes preserve everything literally:

```bash
req read https://api.example.com/search include='header: Authorization: Bearer token; param: q=search query' as=json
```

### Double Quotes
Double quotes allow variable expansion:

```bash
TOKEN="abc123"
req send https://api.example.com/users include="header: Authorization: Bearer $TOKEN" with='{"name":"Adam"}'
```

### Escaping
Use backslash to escape special characters:

```bash
req read https://api.example.com/search include='param: q=test\;value' as=json
```

## Fish Shell

Fish uses different quoting rules:

```fish
req read https://api.example.com/search include='header: Authorization: Bearer token; param: q=search query' as=json
```

For variables in fish:

```fish
set TOKEN "abc123"
req send https://api.example.com/users include="header: Authorization: Bearer $TOKEN" with='{"name":"Adam"}'
```

## PowerShell

PowerShell uses backticks for escaping:

```powershell
req read https://api.example.com/search include='header: Authorization: Bearer token; param: q=search query' as=json
```

For variables in PowerShell:

```powershell
$TOKEN = "abc123"
req send https://api.example.com/users include="header: Authorization: Bearer $TOKEN" with='{"name":"Adam"}'
```

## Common Patterns

### Include Clause with Multiple Items

```bash
# Bash/Zsh
req read https://api.example.com/search \
  include='header: Accept: application/json; param: q=test; cookie: session=abc123' \
  as=json

# Fish
req read https://api.example.com/search \
  include='header: Accept: application/json; param: q=test; cookie: session=abc123' \
  as=json

# PowerShell
req read https://api.example.com/search `
  include='header: Accept: application/json; param: q=test; cookie: session=abc123' `
  as=json
```

### Expect Clause with Multiple Checks

```bash
req read https://api.example.com/users \
  expect='status:200, header:Content-Type=application/json, contains:"items"' \
  as=json
```

### Attach Clause with File Paths

```bash
req upload https://api.example.com/upload \
  attach='part: name=file, file=@./avatar.png, type=image/png; part: name=meta, value={"name":"adam"}' \
  as=json
```

## Curl vs req Mapping

| curl command | req equivalent |
|-------------|----------------|
| `curl https://api.example.com/users` | `req read https://api.example.com/users` |
| `curl -H "Authorization: Bearer $TOKEN" https://api.example.com/users` | `req read https://api.example.com/users include="header: Authorization: Bearer $TOKEN"` |
| `curl -X POST -d '{"name":"Adam"}' https://api.example.com/users` | `req send https://api.example.com/users with='{"name":"Adam"}'` |
| `curl -X POST -F "file=@avatar.png" https://api.example.com/upload` | `req upload https://api.example.com/upload attach='part: name=file, file=@avatar.png'` |
| `curl -b "session=abc123" https://api.example.com/users` | `req read https://api.example.com/users include='cookie: session=abc123'` |
| `curl -L https://example.com` | `req read https://example.com` (follows redirects by default) |
| `curl -k https://self-signed.example.com` | `req read https://self-signed.example.com insecure=true` |
| `curl --proxy http://proxy:8080 https://api.example.com` | `req read https://api.example.com via=http://proxy:8080` |
| `curl -X POST --data-binary @file.json https://api.example.com` | `req send https://api.example.com with=@file.json` |
| `curl -X POST --data @- https://api.example.com` | `echo '{"data":"value"}' \| req send https://api.example.com with=@-` |

## Tips

1. **Always quote include= values** - They often contain semicolons and spaces
2. **Use single quotes for JSON** - Prevents shell from interpreting special characters
3. **Use double quotes when you need variable expansion** - But be careful with nested quotes
4. **Test with `req explain`** - See how your command is parsed before executing:
   ```bash
   req explain "read https://api.example.com/users include='header: Authorization: Bearer token'"
   ```

