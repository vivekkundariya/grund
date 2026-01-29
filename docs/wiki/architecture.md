# Grund Architecture

This document describes the architecture of Grund, which follows **SOLID principles** and **Domain-Driven Design (DDD)**.

## Documentation Index

| Document | Description |
|----------|-------------|
| [Architecture](./architecture.md) | This document - system architecture and design |
| [CLI Commands](./cli-commands.md) | Complete CLI reference with examples |
| [Algorithms](./algorithms.md) | Dependency resolution, topological sort, etc. |
| [Configuration](./configuration.md) | All config files and schemas |
| [Adding Infrastructure](./adding-new-infrastructure.md) | Guide to extend Grund |

**Decision Logs:**
| Document | Description |
|----------|-------------|
| [System Design](../plans/2026-01-22-localdev-system-design.md) | Original system design decisions |
| [Testing Strategy](../plans/2026-01-22-testing-strategy.md) | Testing approach and patterns |

## Architecture Overview

Grund uses a **layered architecture** with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                  CLI Layer (internal/cli)                   │
│  - Cobra commands (up, down, status, logs, etc.)           │
│  - Flag parsing and user interaction                        │
│  - Delegates to application layer                           │
└─────────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│            Application Layer (internal/application)         │
│  - Commands: UpCommandHandler, DownCommandHandler, etc.    │
│  - Queries: StatusQueryHandler, ConfigQueryHandler         │
│  - Ports: Interfaces defining contracts                     │
│  - Wiring: Dependency injection container                   │
└─────────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│              Domain Layer (internal/domain)                 │
│  - Entities: Service, InfrastructureRequirements           │
│  - Value Objects: Port, ServiceName, ServiceType           │
│  - Domain Services: Dependency graph resolution             │
└─────────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│         Infrastructure Layer (internal/infrastructure)      │
│  - Docker: Orchestrator, Provisioners, HealthChecker       │
│  - AWS: LocalStack provisioner                              │
│  - Config: Service repositories                             │
│  - Generator: Compose file, Environment resolver            │
└─────────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│              Cross-Cutting Concerns                         │
│  - UI (internal/ui): Logging, colors, spinners, tables     │
│  - Config (internal/config): Global/local config parsing   │
└─────────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### 1. Domain Layer (`internal/domain/`)

**Purpose**: Contains core business logic and domain models with no external dependencies.

**Packages**:
- `service/` - Service entity and value objects
- `infrastructure/` - Infrastructure requirements and configs
- `dependency/` - Dependency graph resolution

**Key Types**:

```go
// Service entity with behavior
type Service struct {
    Name         string
    Type         ServiceType          // Value object
    Port         Port                 // Value object
    Build        *BuildConfig
    Health       HealthConfig
    Dependencies ServiceDependencies
    Environment  Environment
}

func (s *Service) Validate() error { ... }
func (s *Service) RequiresInfrastructure(infraType string) bool { ... }

// Value objects
type Port struct { value int }
type ServiceName string
type ServiceType string  // "go", "python", "node"

// Infrastructure requirements aggregate
type InfrastructureRequirements struct {
    Postgres *PostgresConfig
    MongoDB  *MongoDBConfig
    Redis    *RedisConfig
    SQS      *SQSConfig
    SNS      *SNSConfig
    S3       *S3Config
}

func (r *InfrastructureRequirements) Has(infraType string) bool { ... }
func Aggregate(requirements ...InfrastructureRequirements) InfrastructureRequirements { ... }
```

**Principles**:
- Zero dependencies on other layers
- Pure business logic
- Rich domain models with behavior
- Infrastructure aggregation handles multi-service scenarios

### 2. Application Layer (`internal/application/`)

**Purpose**: Orchestrates domain objects to fulfill use cases via CQRS pattern.

**Packages**:
- `commands/` - Write operations (Up, Down, Restart)
- `queries/` - Read operations (Status, Config)
- `ports/` - Interface definitions (contracts)
- `wiring/` - Dependency injection container

**Command Handlers**:

```go
// UpCommandHandler - starts services with dependencies
type UpCommandHandler struct {
    serviceRepo      ports.ServiceRepository
    registryRepo     ports.ServiceRegistryRepository
    orchestrator     ports.ContainerOrchestrator
    provisioner      ports.InfrastructureProvisioner
    composeGenerator ports.ComposeGenerator
    healthChecker    ports.HealthChecker
}

func (h *UpCommandHandler) Handle(ctx context.Context, cmd UpCommand) error {
    // 1. Load services and resolve dependencies
    // 2. Aggregate infrastructure requirements
    // 3. Generate docker-compose.yaml
    // 4. Start infrastructure and wait for health
    // 5. Provision resources (databases, queues, etc.)
    // 6. Start application services
}

// DownCommandHandler - stops all services
type DownCommandHandler struct {
    orchestrator ports.ContainerOrchestrator
}

// RestartCommandHandler - restarts specific service
type RestartCommandHandler struct {
    orchestrator ports.ContainerOrchestrator
}
```

**Query Handlers**:

```go
// StatusQueryHandler - gets service statuses
type StatusQueryHandler struct {
    orchestrator ports.ContainerOrchestrator
}

// ConfigQueryHandler - resolves service configuration
type ConfigQueryHandler struct {
    serviceRepo  ports.ServiceRepository
    registryRepo ports.ServiceRegistryRepository
    envResolver  ports.EnvironmentResolver
}
```

**Ports (Interfaces)**:

```go
// Repository interfaces
type ServiceRepository interface {
    FindByName(name service.ServiceName) (*service.Service, error)
    FindAll() ([]*service.Service, error)
    Save(svc *service.Service) error
}

type ServiceRegistryRepository interface {
    GetServicePath(name service.ServiceName) (string, error)
    GetAllServices() (map[service.ServiceName]ServiceEntry, error)
}

// Orchestrator interface
type ContainerOrchestrator interface {
    StartInfrastructure(ctx context.Context) error
    StartServices(ctx context.Context, services []service.ServiceName) error
    StopServices(ctx context.Context) error
    RestartService(ctx context.Context, name service.ServiceName) error
    GetServiceStatus(ctx context.Context, name service.ServiceName) (ServiceStatus, error)
    GetAllServiceStatuses(ctx context.Context) ([]ServiceStatus, error)
}

// Provisioner interface
type InfrastructureProvisioner interface {
    ProvisionPostgres(ctx context.Context, config *infrastructure.PostgresConfig) error
    ProvisionMongoDB(ctx context.Context, config *infrastructure.MongoDBConfig) error
    ProvisionRedis(ctx context.Context, config *infrastructure.RedisConfig) error
    ProvisionLocalStack(ctx context.Context, req infrastructure.InfrastructureRequirements) error
}

// Generator interfaces
type ComposeGenerator interface {
    Generate(services []*service.Service, infra infrastructure.InfrastructureRequirements) (*ComposeFileSet, error)
}

// ComposeFileSet tracks generated per-service compose files
type ComposeFileSet struct {
    InfrastructurePath string            // ~/.grund/tmp/infrastructure/docker-compose.yaml
    ServicePaths       map[string]string // service name -> path
}

type EnvironmentResolver interface {
    Resolve(envRefs map[string]string, context EnvironmentContext) (map[string]string, error)
}

// Health checker interface
type HealthChecker interface {
    CheckHealth(ctx context.Context, endpoint string, timeout int) error
    WaitForHealthy(ctx context.Context, endpoint string, interval, timeout int, retries int) error
}
```

### 3. Infrastructure Layer (`internal/infrastructure/`)

**Purpose**: Implements adapters for external systems (Docker, AWS, file system).

**Packages**:
- `docker/` - Docker orchestrator, provisioners, health checker
- `aws/` - LocalStack provisioner for AWS services
- `config/` - File-based repositories
- `generator/` - Compose file and environment variable generators

**Docker Orchestrator**:

```go
// Implements ports.ContainerOrchestrator
type DockerOrchestrator struct {
    composeFile string
    workingDir  string
}

func (d *DockerOrchestrator) StartServices(ctx context.Context, services []service.ServiceName) error {
    // Execute: docker compose -f <file> up -d --build <services>
}

func (d *DockerOrchestrator) StopServices(ctx context.Context) error {
    // Execute: docker compose -f <file> down
}
```

**Provisioner Pattern (Composite)**:

```go
// Individual provisioners
type PostgresProvisioner struct{}
type MongoDBProvisioner struct{}
type RedisProvisioner struct{}
type LocalStackProvisioner struct{ endpoint string }

// Composite provisioner combines all
type CompositeInfrastructureProvisioner struct {
    postgres   *PostgresProvisioner
    mongodb    *MongoDBProvisioner
    redis      *RedisProvisioner
    localstack *LocalStackProvisioner
}

func (c *CompositeInfrastructureProvisioner) ProvisionPostgres(ctx context.Context, config *infrastructure.PostgresConfig) error {
    return c.postgres.Provision(ctx, config)
}
```

**Environment Resolver**:

```go
// Resolves ${placeholder} syntax to actual values
type EnvironmentResolverImpl struct{}

func (r *EnvironmentResolverImpl) Resolve(envRefs map[string]string, ctx ports.EnvironmentContext) (map[string]string, error) {
    // Resolves: ${infra.postgres.host} → "postgres"
    //           ${sqs.my-queue.url} → "http://localstack:4566/000000000000/my-queue"
    //           ${service.api.host} → "api"
}
```

### 4. CLI Layer (`internal/cli/`)

**Purpose**: User interface via Cobra commands.

**Commands**:
- `up` - Start services with dependencies
- `down` - Stop all services
- `status` - Show service status
- `logs` - Stream service logs
- `restart` - Restart specific service
- `reset` - Stop and remove all containers/volumes
- `config` - Show configuration
- `init` - Initialize a new service
- `setup` - Initialize global config

**Structure**:

```go
// root.go - Main command and initialization
var rootCmd = &cobra.Command{
    Use:   "grund",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        // Initialize config resolver
        // Create dependency injection container
        container, err = wiring.NewContainerWithConfig(orchestrationRoot, servicesPath, configResolver)
    },
}

// up.go - Delegates to application layer
var upCmd = &cobra.Command{
    RunE: func(cmd *cobra.Command, args []string) error {
        upCmd := commands.UpCommand{ServiceNames: args}
        return container.UpCommandHandler.Handle(cmd.Context(), upCmd)
    },
}
```

### 5. Cross-Cutting Concerns

**UI Package (`internal/ui/`)**:

```go
// Logging with levels and colors
func Debug(format string, args ...any)    // Only with --verbose
func Infof(format string, args ...any)
func Successf(format string, args ...any)
func Warnf(format string, args ...any)
func Errorf(format string, args ...any)

// Progress indicators
func Step(format string, args ...any)     // → Starting...
func SubStep(format string, args ...any)  //   • Substep...
func Header(format string, args ...any)
func ServiceStatus(name, status string, healthy bool)
```

**Config Package (`internal/config/`)**:

```go
// ConfigResolver handles config file discovery
type ConfigResolver struct {
    GlobalConfig *GlobalConfig
}

func (r *ConfigResolver) ResolveServicesFile() (servicesPath, orchestrationRoot string, err error)

// GlobalConfig at ~/.grund/config.yaml
type GlobalConfig struct {
    DefaultServicesFile      string
    DefaultOrchestrationRepo string
    ServicesBasePath         string
    Docker                   DockerConfig
    LocalStack               LocalStackConfig
}
```

## Dependency Injection

All dependencies are wired in `internal/application/wiring/wiring.go`:

```go
type Container struct {
    // Configuration
    ConfigResolver    *appconfig.ConfigResolver
    OrchestrationRoot string
    ServicesPath      string

    // Repositories
    ServiceRepo  interface{}  // ports.ServiceRepository
    RegistryRepo interface{}  // ports.ServiceRegistryRepository

    // Infrastructure
    Orchestrator     interface{}  // ports.ContainerOrchestrator
    Provisioner      interface{}  // ports.InfrastructureProvisioner
    ComposeGenerator interface{}  // ports.ComposeGenerator
    EnvResolver      interface{}  // ports.EnvironmentResolver
    HealthChecker    interface{}  // ports.HealthChecker

    // Command Handlers
    UpCommandHandler      *commands.UpCommandHandler
    DownCommandHandler    *commands.DownCommandHandler
    RestartCommandHandler *commands.RestartCommandHandler

    // Query Handlers
    StatusQueryHandler *queries.StatusQueryHandler
    ConfigQueryHandler *queries.ConfigQueryHandler
}

func NewContainerWithConfig(orchestrationRoot, servicesPath string, configResolver *appconfig.ConfigResolver) (*Container, error) {
    // 1. Create repositories
    registryRepo, _ := config.NewServiceRegistryRepository(servicesPath)
    serviceRepo := config.NewServiceRepository(registryRepo)

    // 2. Create infrastructure adapters
    orchestrator := docker.NewDockerOrchestrator(composeFile, orchestrationRoot)
    healthChecker := docker.NewHTTPHealthChecker()

    // 3. Create composite provisioner
    provisioner := docker.NewCompositeInfrastructureProvisioner(
        docker.NewPostgresProvisioner(),
        docker.NewMongoDBProvisioner(),
        docker.NewRedisProvisioner(),
        aws.NewLocalStackProvisioner(localstackEndpoint),
    )

    // 4. Create generators (tmpDir is ~/.grund/tmp)
    composeGenerator := generator.NewComposeGenerator(tmpDir)
    envResolver := generator.NewEnvironmentResolver()

    // 5. Wire command/query handlers
    upHandler := commands.NewUpCommandHandler(serviceRepo, registryRepo, orchestrator, provisioner, composeGenerator, healthChecker)
    // ...
}
```

## SOLID Principles Applied

### 1. Single Responsibility Principle (SRP)
- `UpCommandHandler` only orchestrates service startup
- `DockerOrchestrator` only handles Docker Compose operations
- `PostgresProvisioner` only handles Postgres-specific setup
- `EnvironmentResolver` only resolves environment variables

### 2. Open/Closed Principle (OCP)
- New infrastructure types can be added without modifying existing code
- Add new provisioner → register in `CompositeInfrastructureProvisioner`
- See [Adding New Infrastructure](./adding-new-infrastructure.md) for the extension pattern

### 3. Liskov Substitution Principle (LSP)
- All `ports.*` interfaces can be substituted with test doubles
- Any `ContainerOrchestrator` implementation works (Docker, Podman, etc.)

### 4. Interface Segregation Principle (ISP)
- `ServiceRepository` only has service-related methods
- `ContainerOrchestrator` is focused on container operations
- `InfrastructureProvisioner` has separate methods per infrastructure type

### 5. Dependency Inversion Principle (DIP)
- Application layer depends on `ports.*` interfaces, not implementations
- CLI layer depends on application layer handlers
- All concrete implementations are in infrastructure layer

## Design Patterns Used

1. **Repository Pattern**: `ServiceRepository`, `ServiceRegistryRepository`
2. **Adapter Pattern**: Infrastructure adapters implement application ports
3. **Command Pattern**: Commands encapsulate write operations
4. **CQRS**: Separate command and query handlers
5. **Dependency Injection**: Container wires all dependencies
6. **Composite Pattern**: `CompositeInfrastructureProvisioner` combines provisioners
7. **Value Object Pattern**: `Port`, `ServiceName`, `ServiceType`
8. **Aggregate Pattern**: `InfrastructureRequirements.Aggregate()` merges requirements

## File Structure

```
internal/
├── application/          # Use cases and ports
│   ├── commands/         # Command handlers (CQRS writes)
│   │   ├── up_command.go
│   │   ├── down_command.go
│   │   └── restart_command.go
│   ├── queries/          # Query handlers (CQRS reads)
│   │   ├── status_query.go
│   │   └── config_query.go
│   ├── ports/            # Interface contracts
│   │   ├── repository.go
│   │   ├── orchestrator.go
│   │   └── compose.go
│   └── wiring/           # Dependency injection
│       └── wiring.go
├── domain/               # Core business logic
│   ├── service/          # Service entity
│   │   └── service.go
│   ├── infrastructure/   # Infrastructure domain
│   │   └── infrastructure.go
│   └── dependency/       # Dependency resolution
│       └── graph.go
├── infrastructure/       # External system adapters
│   ├── docker/           # Docker orchestration
│   │   ├── orchestrator.go
│   │   ├── postgres_provisioner.go
│   │   ├── mongodb_provisioner.go
│   │   ├── redis_provisioner.go
│   │   ├── infrastructure_provisioner.go  # Composite
│   │   └── health_checker.go
│   ├── aws/              # AWS/LocalStack
│   │   └── provisioner.go
│   ├── config/           # File-based repositories
│   │   ├── service_repository.go
│   │   └── registry_repository.go
│   └── generator/        # Compose & env generation
│       ├── compose_generator.go
│       └── env_resolver.go
├── cli/                  # Cobra CLI commands
│   ├── root.go
│   ├── up.go
│   ├── down.go
│   ├── status.go
│   ├── logs.go
│   ├── restart.go
│   ├── reset.go
│   ├── config.go
│   ├── init.go
│   └── add.go
├── config/               # Configuration parsing
│   ├── resolver.go       # Config file discovery
│   ├── global.go         # Global config (~/.grund)
│   ├── schema.go         # YAML schema definitions
│   └── parser.go
└── ui/                   # User interface utilities
    ├── logger.go         # Structured logging
    ├── colors.go         # Terminal colors
    ├── spinner.go        # Progress spinners
    └── table.go          # Table formatting
```

## Data Flow Example: `grund up user-service`

```
┌─────────────────────────────────────────────────────────────────────────┐
│ 1. CLI: Parse args, initialize container                                │
│    up.go → wiring.Container                                             │
└────────────────────────────────────┬────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ 2. Application: UpCommandHandler.Handle()                               │
│    - Load service via ServiceRepository                                 │
│    - Resolve dependencies via dependency graph                          │
│    - Aggregate infrastructure requirements                              │
└────────────────────────────────────┬────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ 3. Infrastructure: Generate per-service compose files                   │
│    ComposeGenerator.Generate(services, infraRequirements)               │
└────────────────────────────────────┬────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ 4. Infrastructure: Start containers                                     │
│    - DockerOrchestrator.StartInfrastructure() → postgres, redis, etc.  │
│    - HealthChecker.WaitForHealthy() → wait for ready                   │
└────────────────────────────────────┬────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ 5. Infrastructure: Provision resources                                  │
│    - PostgresProvisioner.Provision() → create database                 │
│    - LocalStackProvisioner.Provision() → create queues/topics          │
└────────────────────────────────────┬────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ 6. Infrastructure: Start application services                           │
│    DockerOrchestrator.StartServices(["user-service"])                  │
└─────────────────────────────────────────────────────────────────────────┘
```

## Benefits of This Architecture

1. **Testability**: Easy to mock interfaces for unit testing
2. **Maintainability**: Clear separation of concerns
3. **Extensibility**: Add new infrastructure without modifying core logic
4. **Flexibility**: Can swap implementations (Docker → Podman, etc.)
5. **Domain Focus**: Business logic is isolated and testable
6. **Traceability**: Clear data flow through layers
