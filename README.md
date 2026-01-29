# Grund - Local Development Orchestration Tool

**Grund** is a CLI tool that enables developers to selectively spin up microservices and their dependencies with a single command. Declare dependencies in your service repos, and Grund resolves the full dependency tree, provisions infrastructure (databases, queues, caches), and starts everything in the correct order.

## Table of Contents

- [Why Grund?](#why-grund)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Getting Started](#getting-started)
- [Integration Guide](#integration-guide)
- [Commands Reference](#commands-reference)
- [Configuration Reference](#configuration-reference)
- [Shell Completion](#shell-completion)
- [Troubleshooting](#troubleshooting)

## Why Grund?

In a microservices architecture, running a single service locally often requires:
- Multiple dependent services
- Databases (PostgreSQL, MongoDB)
- Message queues (SQS, SNS)
- Object storage (S3)
- Cache (Redis)

Grund solves this by:
1. **Declarative dependencies** - Each service declares what it needs in a `grund.yaml`
2. **Automatic resolution** - Grund builds the full dependency tree
3. **Infrastructure provisioning** - Databases, queues, and buckets are created automatically
4. **Correct startup order** - Dependencies start before dependents

```bash
$ grund up payment-service

Starting infrastructure...
  ✓ postgres (localhost:5432)
  ✓ redis (localhost:6379)
  ✓ localstack (localhost:4566)
    → sqs: payment-queue ✓

Starting services...
  ✓ user-service (localhost:8081)
  ✓ payment-service (localhost:8080)

Ready!
```

## Prerequisites

- **Docker** and **Docker Compose** v2+
- **Go 1.21+** (for building from source)

## Installation

### Using Go Install

```bash
go install github.com/vivekkundariya/grund@latest
```

### From Source

```bash
git clone https://github.com/vivekkundariya/grund.git
cd grund
make install
```

### Verify Installation

```bash
grund --version
```

## Getting Started

### Step 1: Create an Orchestration Repository

Create a central repository to manage all your services:

```bash
mkdir my-company-local-dev
cd my-company-local-dev
```

### Step 2: Initialize Global Configuration

```bash
grund setup
```

This creates `~/.grund/config.yaml` with default settings.

### Step 3: Create Services Registry

Create a `services.yaml` file in your orchestration repo:

```yaml
version: "1"

services:
  user-service:
    repo: git@github.com:mycompany/user-service.git
    path: ~/projects/user-service

  payment-service:
    repo: git@github.com:mycompany/payment-service.git
    path: ~/projects/payment-service

  notification-service:
    repo: git@github.com:mycompany/notification-service.git
    path: ~/projects/notification-service
```

### Step 4: Set Environment Variable

Add to your `~/.zshrc` or `~/.bashrc`:

```bash
export GRUND_CONFIG=~/my-company-local-dev/services.yaml
```

Reload your shell:
```bash
source ~/.zshrc
```

### Step 5: Initialize Services

In each service repository, create a `grund.yaml`:

```bash
cd ~/projects/user-service
grund init
```

The interactive wizard will guide you through:
- Service name, type, and port
- Infrastructure requirements (PostgreSQL, Redis, etc.)
- Service dependencies

### Step 6: Start Services

```bash
grund up user-service
```

## Integration Guide

### Adding Grund to an Existing Service

1. **Navigate to your service directory:**
   ```bash
   cd ~/projects/my-service
   ```

2. **Initialize Grund:**
   ```bash
   grund init
   ```

3. **Or create `grund.yaml` manually:**
   ```yaml
   version: "1"

   service:
     name: my-service
     type: go  # go, python, or node
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
     services: []
     infrastructure: {}

   env:
     APP_ENV: development
     LOG_LEVEL: debug

   env_refs: {}
   ```

4. **Add infrastructure as needed:**
   ```bash
   grund add postgres --database my_service_db
   grund add redis
   grund add sqs --queue events-queue
   ```

5. **Register in `services.yaml`:**
   ```yaml
   services:
     my-service:
       repo: git@github.com:mycompany/my-service.git
       path: ~/projects/my-service
   ```

### Adding Dependencies Between Services

If `payment-service` depends on `user-service`:

```bash
cd ~/projects/payment-service
grund add service user-service
```

This adds to `grund.yaml`:
```yaml
requires:
  services:
    - user-service

env_refs:
  USER_SERVICE_URL: "http://${user-service.host}:${user-service.port}"
```

### Infrastructure Options

#### PostgreSQL
```bash
grund add postgres --database myapp_db --migrations ./migrations
```

#### MongoDB
```bash
grund add mongodb --database myapp_db
```

#### Redis
```bash
grund add redis
```

#### SQS (via LocalStack)
```bash
grund add sqs --queue order-events --dlq
```

#### SNS (via LocalStack)
```bash
grund add sns --topic notifications
```

#### S3 (via LocalStack)
```bash
grund add s3 --bucket uploads
```

## Commands Reference

### `grund up`
Start services and their dependencies.

```bash
grund up <service...>           # Start specific services
grund up user-service           # Start one service
grund up user-service payment   # Start multiple services
grund up user-service --no-deps # Start without dependencies
grund up user-service --build   # Force rebuild containers
grund up user-service --infra-only # Only start infrastructure
```

### `grund down`
Stop all running services.

```bash
grund down
```

### `grund status`
Show running services and their status.

```bash
grund status
```

Output:
```
╭──────────────────┬─────────────╮
│ SERVICE          │ STATUS      │
├──────────────────┼─────────────┤
│ postgres         │ ● running   │
│ redis            │ ● running   │
│ user-service     │ ● running   │
│ payment-service  │ ○ not running│
╰──────────────────┴─────────────╯
```

### `grund logs`
View service logs.

```bash
grund logs                           # All services
grund logs user-service              # Specific service
grund logs user-service order-service  # Multiple services
grund logs -f                        # Follow mode (like tail -f)
grund logs --tail 50                 # Last 50 lines
grund logs user-service -f           # Follow specific service
```

### `grund restart`
Restart specific services.

```bash
grund restart user-service
grund restart user-service payment-service
grund restart user-service --build  # Rebuild before restart
```

### `grund reset`
Stop services and clean up resources.

```bash
grund reset              # Stop all services
grund reset -v           # Stop and remove volumes (database data)
grund reset -v --images  # Full cleanup (volumes + images)
```

### `grund config`
Show resolved configuration for a service.

```bash
grund config user-service
```

Output:
```
  Service: user-service
  Type:    go
  Port:    8080

  Dependencies: payment-service
  Infrastructure: postgres, redis

╭─────────────────────┬──────────────────────────────────────╮
│ Environment Variable│ Value                                │
├─────────────────────┼──────────────────────────────────────┤
│ APP_ENV             │ development                          │
│ DATABASE_URL        │ postgres://postgres:postgres@...     │
│ REDIS_URL           │ redis://localhost:6379               │
╰─────────────────────┴──────────────────────────────────────╯
```

### `grund init`
Initialize `grund.yaml` in current directory.

```bash
cd ~/projects/my-service
grund init
```

### `grund add`
Add resources to existing `grund.yaml`.

```bash
grund add postgres --database mydb
grund add mongodb --database mydb
grund add redis
grund add sqs --queue my-queue --dlq
grund add sns --topic my-topic
grund add s3 --bucket my-bucket
grund add service other-service
```

### `grund setup`
Initialize global configuration at `~/.grund/config.yaml`.

```bash
grund setup
```

## Configuration Reference

### Service Configuration (`grund.yaml`)

```yaml
version: "1"

service:
  name: my-service              # Service name (must match services.yaml)
  type: go                      # go, python, or node
  port: 8080                    # Port the service listens on
  build:
    dockerfile: Dockerfile      # Path to Dockerfile
    context: .                  # Build context
  health:
    endpoint: /health           # Health check endpoint
    interval: 5s                # Check interval
    timeout: 3s                 # Request timeout
    retries: 10                 # Retries before unhealthy

requires:
  services:                     # Service dependencies
    - user-service
    - auth-service
  infrastructure:
    postgres:
      database: my_db           # Database name
      migrations: ./migrations  # Optional: migrations path
    mongodb:
      database: my_mongo_db
    redis: true
    sqs:
      queues:
        - name: order-queue
          dlq: true             # Create dead-letter queue
        - name: notification-queue
    sns:
      topics:
        - name: order-events
    s3:
      buckets:
        - name: uploads
        - name: exports

env:                            # Static environment variables
  APP_ENV: development
  LOG_LEVEL: debug

env_refs:                       # Dynamic environment variables
  DATABASE_URL: "postgres://postgres:postgres@${postgres.host}:${postgres.port}/${self.postgres.database}"
  REDIS_URL: "redis://${redis.host}:${redis.port}"
  USER_SERVICE_URL: "http://${user-service.host}:${user-service.port}"
  AWS_ENDPOINT: "http://${localstack.host}:${localstack.port}"
```

### Services Registry (`services.yaml`)

```yaml
version: "1"

services:
  user-service:
    repo: git@github.com:mycompany/user-service.git
    path: ~/projects/user-service

  payment-service:
    repo: git@github.com:mycompany/payment-service.git
    path: ~/projects/payment-service
```

### Global Configuration (`~/.grund/config.yaml`)

```yaml
default_services_path: ~/my-company-local-dev/services.yaml
```

### Environment Variable Interpolation

Use these placeholders in `env_refs`:

| Placeholder | Description |
|-------------|-------------|
| `${postgres.host}` | PostgreSQL hostname |
| `${postgres.port}` | PostgreSQL port (5432) |
| `${mongodb.host}` | MongoDB hostname |
| `${mongodb.port}` | MongoDB port (27017) |
| `${redis.host}` | Redis hostname |
| `${redis.port}` | Redis port (6379) |
| `${localstack.endpoint}` | LocalStack endpoint URL |
| `${localstack.host}` | LocalStack hostname |
| `${localstack.port}` | LocalStack port (4566) |
| `${localstack.region}` | AWS region (us-east-1) |
| `${localstack.account_id}` | AWS account ID (000000000000) |
| `${localstack.access_key_id}` | AWS access key ID |
| `${localstack.secret_access_key}` | AWS secret access key |
| `${sqs.<queue>.url}` | SQS queue URL |
| `${sqs.<queue>.arn}` | SQS queue ARN |
| `${sqs.<queue>.dlq}` | SQS dead-letter queue URL |
| `${sns.<topic>.arn}` | SNS topic ARN |
| `${s3.<bucket>.url}` | S3 bucket URL |
| `${<service>.host}` | Dependent service hostname |
| `${<service>.port}` | Dependent service port |
| `${self.postgres.database}` | This service's database name |

## Shell Completion

Enable tab completion for commands and service names.

### Zsh
```bash
# Add to ~/.zshrc
eval "$(grund completion zsh)"
```

### Bash
```bash
# Add to ~/.bashrc
eval "$(grund completion bash)"
```

### Fish
```bash
grund completion fish | source
```

After setup:
```bash
grund u<TAB>        → grund up
grund up pay<TAB>   → grund up payment-service
grund --<TAB>       → --config  --verbose  --help
```

## Troubleshooting

### "services.yaml not found"

Set the `GRUND_CONFIG` environment variable:
```bash
export GRUND_CONFIG=/path/to/services.yaml
```

Or use the `--config` flag:
```bash
grund --config=/path/to/services.yaml up my-service
```

### "Container is unhealthy"

Check container logs:
```bash
grund logs <service>
docker logs grund-<service>
```

### "Port already in use"

Stop existing services:
```bash
grund down
grund reset
```

Or check what's using the port:
```bash
lsof -i :8080
```

### "LocalStack resources not created"

Ensure LocalStack is healthy before provisioning:
```bash
grund status
```

If issues persist:
```bash
grund reset -v
grund up <service>
```

### View verbose output

Use `-v` flag for debug information:
```bash
grund -v up my-service
```

## License

MIT
