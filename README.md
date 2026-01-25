# Grund - Local Development Orchestration Tool

**Grund** is a CLI tool and orchestration system that enables developers to selectively spin up services and their dependencies with a single command. Developers declare dependencies in their service repos, and Grund resolves the full dependency tree, provisions infrastructure resources, and starts everything in the correct order.

## Quick Start

```bash
$ grund up service-a

Resolving dependencies...
  service-a → service-b → [postgres, mongodb, localstack]

Starting infrastructure...
  postgres .............. ✓ localhost:5432
  mongodb ............... ✓ localhost:27017
  localstack ............ ✓ localhost:4566
    → sqs: order-queue ✓
    → s3: uploads ✓

Starting services...
  service-b ............. ✓ localhost:8081
  service-a ............. ✓ localhost:8080

✅ Ready in 47s
```

## Installation

```bash
# Build from source
make build

# Or install directly
make install
```

## Commands

| Command   | Description                              |
|-----------|------------------------------------------|
| `up`      | Start services and dependencies          |
| `down`    | Stop all running services                |
| `status`  | Show running services and health         |
| `logs`    | View aggregated or per-service logs      |
| `restart` | Restart specific service(s)              |
| `reset`   | Stop services and clean up resources     |
| `config`  | Show resolved configuration for a service|
| `init`    | Initialize grund.yaml in current service |
| `add`     | Add a resource to existing grund.yaml    |
| `setup`   | Initialize global ~/.grund/config.yaml   |

## Project Structure

```
grund/
├── cmd/
│   └── grund/              # CLI entry point
├── internal/
│   ├── cli/                 # CLI commands
│   ├── config/              # Configuration parsing
│   ├── resolver/            # Dependency resolution
│   ├── generator/           # Docker Compose generation
│   ├── provisioner/         # Infrastructure provisioning
│   ├── executor/            # Docker execution
│   └── ui/                  # Terminal UI components
├── infrastructure/          # Infrastructure definitions
│   ├── postgres/
│   ├── mongodb/
│   ├── redis/
│   └── localstack/
├── templates/               # Configuration templates
├── services.yaml            # Service registry
└── go.mod
```

## Configuration

### Service Configuration (grund.yaml)

Each service defines its dependencies in a `grund.yaml` file:

```yaml
version: "1"

service:
  name: service-a
  type: go
  port: 8080
  build:
    dockerfile: Dockerfile
    context: .
  health:
    endpoint: /health
    interval: 5s
    timeout: 3s
    retries: 10

requires:
  services:
    - service-b
  infrastructure:
    postgres:
      database: service_a_db
    redis: true
    sqs:
      queues:
        - name: order-created-queue
          dlq: true

env:
  APP_ENV: development

env_refs:
  DATABASE_URL: "postgres://postgres:postgres@${postgres.host}:${postgres.port}/${self.postgres.database}"
```

### Service Registry (services.yaml)

The orchestration repo contains a `services.yaml` file mapping service names to repositories:

```yaml
version: "1"

services:
  service-a:
    repo: git@github.com:yourorg/service-a.git
    path: ~/projects/service-a
```

## Development

```bash
# Run tests
make test

# Build binary
make build

# Run locally
make run
```

## License

MIT
