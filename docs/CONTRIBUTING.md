# Contributing to req

Thank you for your interest in contributing to `req`! This guide will help you get started.

## Development Setup

### Prerequisites

- Go 1.24 or later
- Git
- Make (optional, for build automation)

### Getting the Source

```bash
git clone https://github.com/adammpkins/req.git
cd req
```

### Building

```bash
# Build binary
make build

# Binary will be in ./bin/req
./bin/req --version
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific test
go test ./tests -run TestParseBasicAuth -v

# Run with race detector
go test -race ./...
```

## Code Structure

```
req/
├── cmd/req/              # Main entry point
│   └── main.go          # CLI handling, argument parsing
├── internal/
│   ├── parser/          # Command parsing
│   │   └── parser.go    # Lexer and parser implementation
│   ├── planner/         # Execution planning
│   │   └── plan.go      # Plan generation and validation
│   ├── runtime/         # Request execution
│   │   └── executor.go  # HTTP client and execution
│   ├── session/         # Session management
│   │   └── session.go   # Session storage and retrieval
│   ├── types/           # Type definitions
│   │   └── command.go   # AST and plan types
│   └── grammar/         # Grammar definitions
│       └── grammar.go    # Structured grammar data
├── tests/               # Test suite
│   ├── parser_test.go   # Parser tests
│   ├── runtime_test.go  # Integration tests
│   └── ...
└── docs/                # Documentation
```

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Make Changes

- Write code following the existing style
- Add tests for new functionality
- Update documentation if needed

### 3. Test Your Changes

```bash
# Run tests
go test ./...

# Build and test manually
make build
./bin/req <your-test-command>
```

### 4. Commit Changes

```bash
git add .
git commit -m "feat: description of your change"
```

### 5. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub.

## Code Style

### Go Style

- Follow standard Go formatting (`go fmt`)
- Use `golangci-lint` for linting
- Follow Go naming conventions
- Write clear, self-documenting code

### Error Handling

- Return errors, don't panic
- Provide context with `fmt.Errorf("...: %w", err)`
- Use specific error types where appropriate

### Testing

- Write unit tests for new functionality
- Add integration tests for end-to-end scenarios
- Use table-driven tests where appropriate
- Test error cases

### Documentation

- Document exported functions and types
- Update relevant documentation files
- Add examples for new features

## Adding a New Feature

### 1. Grammar Changes

If adding a new clause or verb:

1. Update `.cursor/rules/grammar.mdc` (source of truth)
2. Update parser in `internal/parser/parser.go`
3. Update types in `internal/types/command.go`
4. Update planner in `internal/planner/plan.go`
5. Update executor if needed in `internal/runtime/executor.go`
6. Update grammar snapshot in `internal/grammar/grammar.go`
7. Add tests in `tests/`
8. Update documentation in `docs/`

### 2. Parser Changes

When modifying the parser:

1. Update tokenization logic if needed
2. Add parsing logic for new constructs
3. Add validation
4. Provide helpful error messages
5. Add tests

### 3. Executor Changes

When modifying execution:

1. Update request building if needed
2. Add execution logic
3. Handle errors appropriately
4. Set correct exit codes
5. Add integration tests

## Testing Guidelines

### Unit Tests

- Test individual functions and methods
- Use table-driven tests for multiple cases
- Test both success and error cases
- Mock external dependencies

Example:
```go
func TestParseBasicAuth(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
        check   func(*testing.T, *types.Command)
    }{
        // test cases
    }
    // ...
}
```

### Integration Tests

- Test end-to-end scenarios
- Use real HTTP endpoints (httpbin.org) when possible
- Test error handling
- Verify exit codes

Example:
```go
func TestRuntimeBasicAuth(t *testing.T) {
    cmdStr := `read https://httpbin.org/basic-auth/user/passwd include='basic: user:passwd' expect=status:200`
    // ... test implementation
}
```

### Golden Tests

For output that should be stable:

1. Create golden files in `tests/fixtures/`
2. Compare actual output with golden file
3. Update golden files when output intentionally changes

## Pull Request Process

### Before Submitting

- [ ] Code follows style guidelines
- [ ] All tests pass
- [ ] Documentation updated
- [ ] Grammar documentation updated (if grammar changed)
- [ ] Help output updated (if grammar changed)
- [ ] No linter errors

### PR Description

Include:
- Description of changes
- Why the change is needed
- How it was tested
- Any breaking changes

### Review Process

- Maintainers will review your PR
- Address feedback promptly
- Keep PRs focused and small when possible
- Update PR based on feedback

## Code Review Guidelines

### For Contributors

- Be open to feedback
- Explain your design decisions
- Respond to comments
- Update PR based on feedback

### For Reviewers

- Be constructive and respectful
- Explain reasoning for suggestions
- Approve when ready
- Request changes when needed

## Documentation Standards

### Code Comments

- Document exported functions and types
- Explain "why" not just "what"
- Keep comments up to date

### User Documentation

- Update relevant docs in `docs/`
- Add examples for new features
- Update grammar documentation if syntax changes
- Update help output

## Release Process

Releases are managed by maintainers:

1. Version bumping
2. Changelog updates
3. Tag creation
4. Binary builds
5. Release notes

## Getting Help

- Open an issue for questions
- Check existing issues and PRs
- Review documentation
- Ask in discussions

## See Also

- [Architecture](ARCHITECTURE.md) - System architecture
- [Grammar Reference](GRAMMAR.md) - Grammar specification
- [README](../README.md) - Project overview

