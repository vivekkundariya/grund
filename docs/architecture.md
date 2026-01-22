# Grund Architecture

This document describes the architecture of Grund, which follows **SOLID principles** and **Domain-Driven Design (DDD)**.

## Architecture Overview

Grund uses a **layered architecture** with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────┐
│              Presentation Layer (CLI)                   │
│  - Cobra commands                                       │
│  - User interaction                                     │
└─────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│            Application Layer (Use Cases)                │
│  - Commands (CQRS)                                      │
│  - Queries (CQRS)                                       │
│  - Ports (Interfaces)                                   │
└─────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│              Domain Layer (Business Logic)               │
│  - Entities (Service, Infrastructure)                  │
│  - Value Objects (Port, ServiceName)                   │
│  - Domain Services (Dependency Graph)                   │
└─────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│         Infrastructure Layer (Adapters)                │
│  - Docker (Orchestration)                              │
│  - AWS (LocalStack)                                     │
│  - Config (File System)                                 │
│  - Generator (Docker Compose)                            │
└─────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### 1. Domain Layer (`internal/domain/`)

**Purpose**: Contains core business logic and domain models.

**Components**:
- **Entities**: `service.Service`, infrastructure types
- **Value Objects**: `Port`, `ServiceName`, `ServiceType`
- **Domain Services**: `dependency.Graph` (dependency resolution)

**Principles**:
- No dependencies on other layers
- Pure business logic
- Rich domain models with behavior

**Example**:
```go
// Domain entity with behavior
type Service struct {
    Name         string
    Port         Port  // Value object
    Dependencies ServiceDependencies
}

func (s *Service) Validate() error { ... }
func (s *Service) RequiresInfrastructure(infraType string) bool { ... }
```

### 2. Application Layer (`internal/application/`)

**Purpose**: Orchestrates domain objects to fulfill use cases.

**Components**:
- **Commands**: `commands.UpCommandHandler`, `commands.DownCommandHandler`
- **Queries**: `queries.StatusQueryHandler`, `queries.ConfigQueryHandler`
- **Ports**: Interfaces defining contracts (`ports.ServiceRepository`, `ports.ContainerOrchestrator`)

**Principles**:
- Depends only on domain layer and ports (interfaces)
- Implements CQRS (Command Query Responsibility Segregation)
- Single Responsibility Principle (SRP)

**Example**:
```go
// Command handler - orchestrates domain logic
type UpCommandHandler struct {
    serviceRepo  ports.ServiceRepository  // Interface, not implementation
    orchestrator ports.ContainerOrchestrator
}

func (h *UpCommandHandler) Handle(ctx context.Context, cmd UpCommand) error {
    // 1. Load services (via repository interface)
    // 2. Build dependency graph (domain service)
    // 3. Provision infrastructure (via provisioner interface)
    // 4. Start services (via orchestrator interface)
}
```

### 3. Infrastructure Layer (`internal/infrastructure/`)

**Purpose**: Implements adapters for external systems.

**Components**:
- **Docker**: `docker.DockerOrchestrator`, `docker.HTTPHealthChecker`
- **AWS**: `aws.LocalStackProvisioner`
- **Config**: `config.ServiceRepositoryImpl`, `config.ServiceRegistryRepositoryImpl`
- **Generator**: `generator.ComposeGeneratorImpl`, `generator.EnvironmentResolverImpl`

**Principles**:
- Implements interfaces defined in application layer (Dependency Inversion)
- Handles all external concerns (Docker, file system, AWS)
- Can be swapped without changing application/domain layers

**Example**:
```go
// Infrastructure adapter implements application port
type DockerOrchestrator struct {
    composeFile string
}

func (d *DockerOrchestrator) StartServices(ctx context.Context, services []service.ServiceName) error {
    // Docker-specific implementation
    cmd := exec.Command("docker", "compose", "up", "-d")
    // ...
}
```

### 4. Presentation Layer (`internal/presentation/cli/`)

**Purpose**: User interface (CLI commands).

**Components**:
- Cobra command definitions
- Command/Query translation
- Output formatting

**Principles**:
- Thin layer that delegates to application layer
- No business logic
- Handles user input/output

**Example**:
```go
var upCmd = &cobra.Command{
    RunE: func(cmd *cobra.Command, args []string) error {
        // Translate CLI args to command
        upCmd := commands.UpCommand{ServiceNames: args}
        // Delegate to application layer
        return container.UpCommandHandler.Handle(cmd.Context(), upCmd)
    },
}
```

## SOLID Principles Applied

### 1. Single Responsibility Principle (SRP)
- Each class/package has one reason to change
- `UpCommandHandler` only handles starting services
- `DockerOrchestrator` only handles Docker operations
- `Service` only represents service domain logic

### 2. Open/Closed Principle (OCP)
- Open for extension, closed for modification
- New infrastructure types can be added without changing existing code
- New provisioners can be added via `CompositeInfrastructureProvisioner`

### 3. Liskov Substitution Principle (LSP)
- Implementations can be substituted via interfaces
- Any `ports.ServiceRepository` implementation can be used
- Any `ports.ContainerOrchestrator` implementation can be used

### 4. Interface Segregation Principle (ISP)
- Interfaces are small and focused
- `ports.ServiceRepository` only has service-related methods
- `ports.InfrastructureProvisioner` has separate methods per infrastructure type

### 5. Dependency Inversion Principle (DIP)
- High-level modules depend on abstractions (interfaces)
- Application layer depends on `ports.*` interfaces, not implementations
- Infrastructure layer implements those interfaces

## Dependency Injection

Dependencies are wired in `internal/application/wiring/wiring.go`:

```go
type Container struct {
    ServiceRepo      ports.ServiceRepository
    Orchestrator     ports.ContainerOrchestrator
    UpCommandHandler *commands.UpCommandHandler
    // ...
}

func NewContainer(root string) (*Container, error) {
    // Create infrastructure adapters
    registryRepo := config.NewServiceRegistryRepository(...)
    serviceRepo := config.NewServiceRepository(registryRepo)
    orchestrator := docker.NewDockerOrchestrator(...)
    
    // Wire application services
    upHandler := commands.NewUpCommandHandler(
        serviceRepo,
        orchestrator,
        // ...
    )
    
    return &Container{...}, nil
}
```

## Benefits of This Architecture

1. **Testability**: Easy to mock interfaces for testing
2. **Maintainability**: Clear separation of concerns
3. **Flexibility**: Can swap implementations (e.g., Docker → Kubernetes)
4. **Scalability**: Easy to add new features without breaking existing code
5. **Domain Focus**: Business logic is isolated and testable

## File Structure

```
internal/
├── domain/              # Core business logic
│   ├── service/         # Service domain
│   ├── infrastructure/   # Infrastructure domain
│   └── dependency/      # Dependency resolution domain
├── application/         # Use cases
│   ├── commands/        # Command handlers (CQRS)
│   ├── queries/         # Query handlers (CQRS)
│   ├── ports/           # Interfaces (contracts)
│   └── wiring/         # Dependency injection
├── infrastructure/      # Adapters
│   ├── docker/          # Docker adapter
│   ├── aws/            # AWS/LocalStack adapter
│   ├── config/          # Config file adapter
│   └── generator/       # Compose generator adapter
└── presentation/        # CLI
    └── cli/             # Cobra commands
```

## Design Patterns Used

1. **Repository Pattern**: `ServiceRepository` abstracts data access
2. **Adapter Pattern**: Infrastructure adapters adapt external systems
3. **Command Pattern**: Commands encapsulate operations
4. **CQRS**: Separate commands and queries
5. **Dependency Injection**: Container wires dependencies
6. **Composite Pattern**: `CompositeInfrastructureProvisioner` combines provisioners
7. **Value Object Pattern**: `Port`, `ServiceName` are immutable value objects
