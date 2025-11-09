# Getting Started with req

This guide will help you install `req` and run your first commands.

## Installation

### Using Go Install

The easiest way to install `req` is using `go install`:

```bash
go install github.com/adammpkins/req/cmd/req@latest
```

This will install `req` to your `$GOPATH/bin` directory (or `$HOME/go/bin` by default). Make sure this directory is in your `PATH`.

### Download Pre-built Binaries

Pre-built binaries for Linux, macOS, and Windows are available on the [Releases](https://github.com/adammpkins/req/releases) page.

Download the appropriate binary for your platform and architecture, then:

**Linux/macOS:**
```bash
chmod +x req
sudo mv req /usr/local/bin/
```

**Windows:**
```powershell
# Add to PATH or use directly
.\req.exe
```

### Build from Source

If you want to build from source:

```bash
git clone https://github.com/adammpkins/req.git
cd req
make build
```

The binary will be in `./bin/req`.

## System Requirements

- **Go version:** 1.24 or later (if building from source)
- **Operating System:** Linux, macOS, or Windows
- **Network:** Internet connectivity for making HTTP requests

## Verify Installation

After installation, verify that `req` is working:

```bash
req --version
```

You should see version information. You can also check the help:

```bash
req help
```

## Your First Command

Let's start with a simple GET request:

```bash
req read https://httpbin.org/json
```

This will fetch JSON data from httpbin.org and print it to stdout.

### Pretty Print JSON

To format the JSON output nicely:

```bash
req read https://httpbin.org/json as=json
```

### Add Query Parameters

Add query parameters using the `include` clause:

```bash
req read https://httpbin.org/get \
  include='param: foo=bar; param: baz=qux' \
  as=json
```

### Send a POST Request

Send JSON data with a POST request:

```bash
req send https://httpbin.org/post \
  with='{"name":"Alice","email":"alice@example.com"}' \
  as=json
```

## Understanding the Grammar

`req` uses a simple, natural grammar:

```
req <verb> <url> [clauses...]
```

- **verb**: The action you want to perform (read, send, save, etc.)
- **url**: The target URL
- **clauses**: Optional modifiers (headers, body, output format, etc.)

### Common Verbs

- `read` - GET request, print response to stdout
- `send` - GET by default, POST if body is provided
- `save` - GET request, save response to file
- `upload` - POST with multipart form data

### Common Clauses

- `as=json` - Format output as JSON
- `include='header: Name: Value'` - Add headers, params, or cookies
- `with='{"data":"value"}'` - Request body
- `expect=status:200` - Assert response status

## Next Steps

1. **Learn the Grammar**: Read the [Grammar Reference](GRAMMAR.md) for complete syntax details
2. **Explore Verbs**: See the [Verbs Reference](VERBS.md) for all available verbs
3. **Understand Clauses**: Check the [Clauses Reference](CLAUSES.md) for all modifiers
4. **See Examples**: Browse the [Examples Cookbook](EXAMPLES.md) for common patterns
5. **Migrate from curl**: If you're coming from curl, see the [Migration Guide](CURL_MIGRATION.md)

## Common First Steps

### 1. Make an Authenticated Request

```bash
TOKEN="your-token-here"
req read https://api.example.com/users \
  include="header: Authorization: Bearer $TOKEN" \
  as=json
```

### 2. Save a File

```bash
req save https://example.com/file.zip to=file.zip
```

### 3. Upload a File

```bash
req upload https://api.example.com/upload \
  attach='part: name=file, file=@./document.pdf' \
  as=json
```

### 4. Test an API Endpoint

```bash
req send https://api.example.com/users \
  using=POST \
  with='{"name":"Bob"}' \
  expect=status:201, header:Content-Type=application/json \
  as=json
```

### 5. Authenticate and Use Session

```bash
# Authenticate and store session
req authenticate https://api.example.com/login \
  using=POST \
  with='{"username":"user","password":"pass"}'

# Subsequent requests automatically use the session
req read https://api.example.com/me as=json
```

## Getting Help

- `req help` - Show help message
- `req explain "<command>"` - See how a command is parsed without executing it
- Check the [Error Handling Guide](ERRORS.md) if you encounter issues

## Troubleshooting

### Command Not Found

If you get "command not found", make sure:
1. `req` is installed and in your PATH
2. The binary has execute permissions (`chmod +x req`)
3. You've restarted your terminal after installation

### Permission Denied

On Linux/macOS, ensure the binary has execute permissions:
```bash
chmod +x req
```

### Network Errors

If you get network errors:
- Check your internet connection
- Verify the URL is correct
- Check if a proxy is needed (use `via=` clause)
- See [Error Handling](ERRORS.md) for more details

## See Also

- [Grammar Reference](GRAMMAR.md) - Complete syntax reference
- [Examples Cookbook](EXAMPLES.md) - More examples
- [Error Handling](ERRORS.md) - Troubleshooting guide

