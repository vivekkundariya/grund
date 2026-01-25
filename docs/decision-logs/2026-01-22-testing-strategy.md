# Testing Guide for Grund

This document outlines the testing strategy and how to run tests for the Grund project.

## Testing Strategy

Grund follows a **layered testing approach** aligned with its DDD architecture:

1. **Unit Tests** - Domain layer (pure business logic)
2. **Integration Tests** - Application layer (use cases with mocked dependencies)
3. **End-to-End Tests** - Full CLI workflow with test fixtures

## Test Structure

```
grund/
├── internal/
│   ├── domain/
│   │   ├── service/
│   │   │   └── service_test.go              # Unit tests
│   │   ├── dependency/
│   │   │   └── graph_test.go                # Unit tests
│   │   └── infrastructure/
│   │       └── infrastructure_test.go       # Unit tests
│   ├── application/
│   │   ├── commands/
│   │   │   ├── up_command_test.go           # Integration tests
│   │   │   ├── down_command_test.go
│   │   │   └── restart_command_test.go
│   │   └── queries/
│   │       └── status_query_test.go         # Integration tests
│   └── infrastructure/
│       ├── config/
│       │   ├── service_repository_test.go   # Integration tests
│       │   └── registry_repository_test.go
│       └── docker/
│           └── orchestrator_test.go
├── test/
│   ├── fixtures/                            # Test fixtures
│   │   ├── services/
│   │   │   ├── service-a/
│   │   │   │   ├── grund.yaml
│   │   │   │   └── Dockerfile
│   │   │   ├── service-b/
│   │   │   │   ├── grund.yaml
│   │   │   │   └── Dockerfile
│   │   │   └── service-c/
│   │   │       └── grund.yaml               # For cycle tests
│   │   └── services.yaml
│   ├── integration/                         # E2E tests
│   │   └── cli_test.go
│   ├── helpers/                             # Test helpers
│   │   └── helpers.go
│   └── mocks/                               # Mock implementations
│       └── mocks.go
```

## Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test package
go test ./internal/domain/service/...

# Run tests in verbose mode
go test -v ./...

# Run tests with race detector
go test -race ./...
```

## Test Categories

### 1. Unit Tests (Domain Layer)

Test pure business logic without external dependencies.

**Example: Testing Service Validation**
```go
func TestService_Validate(t *testing.T) {
    tests := []struct {
        name    string
        service *service.Service
        wantErr bool
    }{
        {
            name: "valid service",
            service: &service.Service{
                Name: "test-service",
                Port: port(8080),
                Health: service.HealthConfig{Endpoint: "/health"},
            },
            wantErr: false,
        },
        // ... more test cases
    }
}
```

### 2. Integration Tests (Application Layer)

Test use cases with mocked dependencies.

**Example: Testing UpCommandHandler**
```go
func TestUpCommandHandler_Handle(t *testing.T) {
    // Setup mocks
    mockRepo := &MockServiceRepository{}
    mockOrchestrator := &MockContainerOrchestrator{}
    
    handler := commands.NewUpCommandHandler(
        mockRepo,
        mockOrchestrator,
        // ...
    )
    
    // Test
    err := handler.Handle(context.Background(), commands.UpCommand{
        ServiceNames: []string{"service-a"},
    })
    
    assert.NoError(t, err)
    // Verify mocks were called correctly
}
```

### 3. End-to-End Tests

Test full CLI workflows with test fixtures.

**Example: Testing CLI Commands**
```go
func TestCLI_UpCommand(t *testing.T) {
    // Setup test environment
    testDir := setupTestEnvironment(t)
    defer cleanupTestEnvironment(t, testDir)
    
    // Run CLI command
    cmd := exec.Command("grund", "up", "service-a")
    cmd.Dir = testDir
    output, err := cmd.CombinedOutput()
    
    assert.NoError(t, err)
    assert.Contains(t, string(output), "Starting services")
}
```

## Test Fixtures

Test fixtures are located in `test/fixtures/` and include:

- Sample `grund.yaml` files for different scenarios
- Sample `services.yaml` for service registry
- Mock service directories

## Mocking

Use interfaces (ports) for dependency injection, making it easy to create mocks.

**Example Mock:**
```go
type MockServiceRepository struct {
    FindByNameFunc func(name service.ServiceName) (*service.Service, error)
}

func (m *MockServiceRepository) FindByName(name service.ServiceName) (*service.Service, error) {
    if m.FindByNameFunc != nil {
        return m.FindByNameFunc(name)
    }
    return nil, fmt.Errorf("not implemented")
}
```

## Test Coverage Goals

- **Domain Layer**: 90%+ coverage (pure logic, easy to test)
- **Application Layer**: 80%+ coverage (with mocks)
- **Infrastructure Layer**: 70%+ coverage (integration tests)
- **CLI Layer**: 60%+ coverage (E2E tests)

## Continuous Integration

Tests should run on:
- Every pull request
- Before merging to main
- On every commit (optional, can be nightly)

## Writing Good Tests

1. **Follow AAA Pattern**: Arrange, Act, Assert
2. **Test One Thing**: Each test should verify one behavior
3. **Use Descriptive Names**: Test names should describe what they test
4. **Keep Tests Fast**: Unit tests should run in milliseconds
5. **Isolate Tests**: Tests should not depend on each other
6. **Clean Up**: Always clean up test resources

## Example Test Workflow

```bash
# 1. Run unit tests (fast)
go test ./internal/domain/... -v

# 2. Run integration tests (medium speed)
go test ./internal/application/... -v

# 3. Run E2E tests (slower, requires Docker)
go test ./test/integration/... -v

# 4. Check coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Troubleshooting

### Tests Fail Due to Missing Docker
- Ensure Docker is running
- E2E tests require Docker Desktop

### Tests Fail Due to Port Conflicts
- Use random ports in tests
- Clean up containers after tests

### Flaky Tests
- Add retries for external dependencies
- Use timeouts appropriately
- Mock external services when possible
