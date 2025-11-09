# req Documentation

Welcome to the comprehensive documentation for `req`, a semantic HTTP client written in Go that replaces traditional curl syntax with a natural, intent-based grammar.

## Documentation Overview

This documentation is organized into several sections covering all aspects of `req`, from getting started to advanced usage patterns.

## Quick Navigation

### Getting Started
- **[Getting Started Guide](GETTING_STARTED.md)** - Installation, quick start, and first steps
- **[Grammar Reference](GRAMMAR.md)** - Complete grammar specification and parsing rules

### Core References
- **[Verbs Reference](VERBS.md)** - Detailed documentation for each verb (read, save, send, upload, watch, inspect, authenticate, session)
- **[Clauses Reference](CLAUSES.md)** - Complete reference for all clauses with examples and edge cases

### Examples and Guides
- **[Examples Cookbook](EXAMPLES.md)** - Comprehensive examples organized by use case
- **[Advanced Usage](ADVANCED.md)** - Advanced patterns, tips, and tricks
- **[curl Migration Guide](CURL_MIGRATION.md)** - Side-by-side comparisons and migration from curl

### Specialized Topics
- **[Authentication](AUTHENTICATION.md)** - All authentication methods (Basic Auth, Bearer tokens, sessions)
- **[Session Management](SESSIONS.md)** - Deep dive into session storage and auto-application
- **[Error Handling](ERRORS.md)** - Exit codes, error messages, and troubleshooting
- **[Security Best Practices](SECURITY.md)** - Security considerations and best practices

### Technical Documentation
- **[Architecture](ARCHITECTURE.md)** - System architecture with mermaid diagrams
- **[Cross-Shell Quoting](QUOTING.md)** - Quoting guide for different shells

### Contributing
- **[Contributing Guidelines](CONTRIBUTING.md)** - How to contribute to req

## Documentation Structure

```
docs/
├── README.md              # This file - documentation index
├── GETTING_STARTED.md     # Installation and quick start
├── GRAMMAR.md             # Complete grammar specification
├── VERBS.md               # Verb reference
├── CLAUSES.md             # Clause reference
├── EXAMPLES.md            # Examples cookbook
├── ADVANCED.md            # Advanced usage patterns
├── ARCHITECTURE.md        # System architecture
├── SESSIONS.md            # Session management
├── AUTHENTICATION.md      # Authentication methods
├── ERRORS.md              # Error handling
├── SECURITY.md            # Security best practices
├── CURL_MIGRATION.md      # curl migration guide
├── QUOTING.md             # Cross-shell quoting guide
└── CONTRIBUTING.md        # Contribution guidelines
```

## Quick Links by Topic

### I want to...

**Learn the basics:**
- Start with [Getting Started](GETTING_STARTED.md)
- Understand the [Grammar](GRAMMAR.md)
- See [Examples](EXAMPLES.md)

**Understand a specific feature:**
- Verbs: [Verbs Reference](VERBS.md)
- Clauses: [Clauses Reference](CLAUSES.md)
- Authentication: [Authentication Guide](AUTHENTICATION.md)
- Sessions: [Session Management](SESSIONS.md)

**Solve a problem:**
- [Error Handling](ERRORS.md) - Exit codes and troubleshooting
- [Security Best Practices](SECURITY.md) - Security considerations
- [Cross-Shell Quoting](QUOTING.md) - Quoting issues

**Migrate from curl:**
- [curl Migration Guide](CURL_MIGRATION.md)

**Contribute:**
- [Contributing Guidelines](CONTRIBUTING.md)
- [Architecture](ARCHITECTURE.md) - Understand the codebase

## Documentation Status

**Version:** v0.1  
**Last Updated:** 2025-01-XX  
**Status:** Complete

All documentation is current with req v0.1. If you find any discrepancies or have suggestions for improvement, please open an issue or submit a pull request.

## See Also

- [Main README](../README.md) - Project overview and quick reference
- [Grammar Specification](../.cursor/rules/grammar.mdc) - Source of truth for grammar
- [GitHub Repository](https://github.com/adammpkins/req) - Source code and issues

