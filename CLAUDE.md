# CLAUDE.md - Grund Project Guidelines

## Build & Test Commands

**Always use make commands for building and testing:**

```bash
# Build
make build          # Build binary to bin/grund

# Install
make install        # Install to $GOPATH/bin

# Testing
make test           # Run all tests
make test-unit      # Run domain layer tests only
make test-integration # Run application layer tests
make test-e2e       # Run end-to-end CLI tests
make test-coverage  # Generate coverage report (coverage.html)
make test-race      # Run tests with race detection

# Cleanup
make clean          # Remove bin/ and coverage files

# Development
make run            # Run without building (go run)
```

## Project Structure

This is a DDD (Domain-Driven Design) Go project:

```
cmd/grund/          # Entry point
internal/
  cli/              # Cobra commands (presentation layer)
  application/      # Use cases, commands, queries, ports
  domain/           # Pure business logic (service, infrastructure, dependency)
  infrastructure/   # Adapters (docker, aws, config, generator)
  config/           # Configuration management
test/               # Mocks, helpers, integration tests
```

## Code Standards

- Error wrapping: Always use `fmt.Errorf("context: %w", err)`
- Interfaces: Define where used (consumer-side), not where implemented
- Use `any` instead of `interface{}`
- Follow existing patterns in the codebase
