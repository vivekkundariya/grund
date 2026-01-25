# LocalDev - System Design Document

## 1. Executive Summary

**LocalDev** is a CLI tool and orchestration system that enables developers to selectively spin up services and their dependencies with a single command. Developers declare dependencies in their service repos, and LocalDev resolves the full dependency tree, provisions infrastructure resources, and starts everything in the correct order.

**Core Experience:**

```bash
$ localdev up service-a

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

---

## 2. Technical Context

### 2.1 Environment

- **Developer OS:** macOS (majority)
- **RAM Available:** 16GB-24GB (target: use < 50%)
- **Docker Desktop:** Approved for use
- **Team Size:** 10-15 developers (future: 30)

### 2.2 Existing Stack

- **Services:** 5 currently, max 10 future (polyrepo structure)
- **Languages:** Go (majority), Python (AI services)
- **Databases:** PostgreSQL, MongoDB, Redis
- **Message Brokers:** AWS SQS, SNS
- **Cloud Services:** AWS S3, SQS, SNS (use LocalStack for local)
- **Architecture:** Microservices + 1 legacy monolith being decomposed

### 2.3 Design Decisions

- **CLI Language:** Go
- **Orchestration:** Hybrid (CLI generates Docker Compose)
- **Config Format:** `localdev.yaml` in each service repo
- **Service Discovery:** Environment variables, localhost with different ports
- **AWS Emulation:** LocalStack

---

## 3. Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Developer Machine                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌─────────────┐      ┌──────────────────────────────────────────────┐    │
│   │             │      │           Central Orchestration Repo          │    │
│   │  Developer  │      │                                              │    │
│   │  Terminal   │      │  ┌─────────────┐  ┌──────────────────────┐  │    │
│   │             │      │  │ localdev    │  │ Infrastructure       │  │    │
│   └──────┬──────┘      │  │ CLI (Go)    │  │ Catalog              │  │    │
│          │             │  └─────────────┘  │                      │  │    │
│          │             │                   │  ├── postgres/       │  │    │
│          ▼             │                   │  ├── mongodb/        │  │    │
│   ┌─────────────┐      │                   │  ├── redis/          │  │    │
│   │ localdev    │◄─────┤                   │  ├── localstack/     │  │    │
│   │ CLI         │      │                   │  └── ...             │  │    │
│   └──────┬──────┘      │                   └──────────────────────┘  │    │
│          │             │                                              │    │
│          │             │  ┌──────────────────────────────────────┐   │    │
│          │             │  │ Service Registry (services.yaml)     │   │    │
│          │             │  │                                      │   │    │
│          │             │  │ service-a:                           │   │    │
│          │             │  │   repo: git@github.com:org/svc-a     │   │    │
│          │             │  │   path: ~/projects/service-a         │   │    │
│          │             │  │ service-b:                           │   │    │
│          │             │  │   repo: git@github.com:org/svc-b     │   │    │
│          │             │  │   path: ~/projects/service-b         │   │    │
│          │             │  └──────────────────────────────────────┘   │    │
│          │             │                                              │    │
│          │             └──────────────────────────────────────────────┘    │
│          │                                                                  │
│          │ reads localdev.yaml from each service                           │
│          ▼                                                                  │
│   ┌─────────────────────────────────────────────────────────────────┐      │
│   │                     Service Repositories                         │      │
│   │                                                                  │      │
│   │  ~/projects/service-a/          ~/projects/service-b/           │      │
│   │  ├── localdev.yaml              ├── localdev.yaml               │      │
│   │  ├── Dockerfile                 ├── Dockerfile                  │      │
│   │  └── src/                       └── src/                        │      │
│   │                                                                  │      │
│   └─────────────────────────────────────────────────────────────────┘      │
│                                                                             │
│          │                                                                  │
│          │ generates                                                        │
│          ▼                                                                  │
│   ┌─────────────────────────────────────────────────────────────────┐      │
│   │              docker-compose.generated.yaml                       │      │
│   │                                                                  │      │
│   │  - Aggregated infrastructure                                    │      │
│   │  - Aggregated services                                          │      │
│   │  - Resolved environment variables                               │      │
│   │  - Correct startup order                                        │      │
│   └──────────────────────────┬──────────────────────────────────────┘      │
│                              │                                              │
│                              │ docker-compose up                            │
│                              ▼                                              │
│   ┌─────────────────────────────────────────────────────────────────┐      │
│   │                        Docker Engine                             │      │
│   │                                                                  │      │
│   │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────┐   │      │
│   │  │ postgres │ │ mongodb  │ │  redis   │ │   localstack     │   │      │
│   │  │  :5432   │ │  :27017  │ │  :6379   │ │     :4566        │   │      │
│   │  └──────────┘ └──────────┘ └──────────┘ │ (S3/SQS/SNS)     │   │      │
│   │                                          └──────────────────┘   │      │
│   │  ┌──────────────────┐ ┌──────────────────┐                      │      │
│   │  │    service-a     │ │    service-b     │                      │      │
│   │  │      :8080       │ │      :8081       │                      │      │
│   │  └──────────────────┘ └──────────────────┘                      │      │
│   │                                                                  │      │
│   └─────────────────────────────────────────────────────────────────┘      │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Component Details

### 4.1 Central Orchestration Repo Structure

```
localdev-orchestration/
├── cmd/
│   └── localdev/
│       └── main.go                 # CLI entry point
├── internal/
│   ├── cli/
│   │   ├── root.go                 # Root command, flags
│   │   ├── up.go                   # localdev up
│   │   ├── down.go                 # localdev down
│   │   ├── status.go               # localdev status
│   │   ├── logs.go                 # localdev logs
│   │   ├── restart.go              # localdev restart
│   │   ├── reset.go                # localdev reset
│   │   ├── config.go               # localdev config
│   │   └── init.go                 # localdev init
│   ├── config/
│   │   ├── schema.go               # Structs for localdev.yaml
│   │   ├── parser.go               # YAML parsing
│   │   ├── validator.go            # Config validation
│   │   └── registry.go             # Service registry handling
│   ├── resolver/
│   │   ├── graph.go                # Dependency graph
│   │   ├── topological.go          # Topological sort
│   │   └── circular.go             # Circular dependency detection
│   ├── generator/
│   │   ├── compose.go              # Generate docker-compose.yaml
│   │   ├── env.go                  # Environment variable resolution
│   │   └── templates.go            # Go templates for compose
│   ├── provisioner/
│   │   ├── localstack.go           # AWS resource provisioning
│   │   ├── postgres.go             # Database/migration handling
│   │   └── mongodb.go              # MongoDB initialization
│   ├── executor/
│   │   ├── docker.go               # Docker compose wrapper
│   │   ├── health.go               # Health check polling
│   │   └── logs.go                 # Log streaming
│   └── ui/
│       ├── spinner.go              # Progress spinners
│       ├── table.go                # Status tables
│       └── colors.go               # Terminal colors
├── infrastructure/
│   ├── postgres/
│   │   └── compose.yaml
│   ├── mongodb/
│   │   └── compose.yaml
│   ├── redis/
│   │   └── compose.yaml
│   └── localstack/
│       └── compose.yaml
├── templates/
│   └── localdev.yaml.tmpl          # Template for localdev init
├── services.yaml
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

### 4.2 Service Registry (services.yaml)

This file maps service names to their repository locations:

```yaml
# services.yaml
version: "1"

services:
  core-monolith:
    repo: git@github.com:yourorg/core-monolith.git
    path: ~/projects/core-monolith    # Local path where repo is cloned
    
  service-a:
    repo: git@github.com:yourorg/service-a.git
    path: ~/projects/service-a
    
  service-b:
    repo: git@github.com:yourorg/service-b.git
    path: ~/projects/service-b
    
  ai-service-1:
    repo: git@github.com:yourorg/ai-service-1.git
    path: ~/projects/ai-service-1
    
  ai-service-2:
    repo: git@github.com:yourorg/ai-service-2.git
    path: ~/projects/ai-service-2

# Path expansion rules
path_defaults:
  base: ~/projects    # Default: ~/projects/{service-name}
```

---

### 4.3 Service localdev.yaml Schema

Each service defines its own dependencies. This is the complete schema specification:

```yaml
# ~/projects/service-a/localdev.yaml
version: "1"

service:
  name: service-a
  type: go                          # go | python | node
  port: 8080
  
  # How to build/run
  build:
    dockerfile: Dockerfile          # Path relative to service root
    context: .
    
  # Or run directly (for hot reload during development)
  run:
    command: go run ./cmd/main.go
    hot_reload: true
    
  # Health check
  health:
    endpoint: /health
    interval: 5s
    timeout: 3s
    retries: 10

# Service dependencies
requires:
  services:
    - service-b
    
  infrastructure:
    postgres:
      database: service_a_db
      migrations: ./migrations      # Optional: run migrations
      seed: ./seeds/dev.sql         # Optional: seed data
      
    redis: true                     # Simple flag, no special config
    
    sqs:
      queues:
        - name: order-created-queue
          dlq: true                 # Create dead-letter queue
        - name: payment-queue
        
    sns:
      topics:
        - name: order-events
          subscriptions:
            - queue: order-created-queue    # Auto-subscribe SQS to SNS
            
    s3:
      buckets:
        - name: user-uploads
          seed: ./fixtures/s3/      # Optional: pre-populate files

# Environment variables
# These will be injected into the container
env:
  APP_ENV: development
  LOG_LEVEL: debug
  
# Environment variables with references (resolved by localdev)
env_refs:
  DATABASE_URL: "postgres://postgres:postgres@${postgres.host}:${postgres.port}/${self.postgres.database}"
  REDIS_URL: "redis://${redis.host}:${redis.port}"
  SERVICE_B_URL: "http://${service-b.host}:${service-b.port}"
  AWS_ENDPOINT: "http://${localstack.host}:${localstack.port}"
  ORDER_QUEUE_URL: "${sqs.order-created-queue.url}"
  S3_BUCKET: "${s3.user-uploads.name}"
```

---

### 4.4 Infrastructure Catalog

Each infrastructure component has a standardized definition:

**infrastructure/postgres/compose.yaml:**

```yaml
services:
  postgres:
    image: postgres:15-alpine
    ports:
      - "${POSTGRES_PORT:-5432}:5432"
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

**infrastructure/mongodb/compose.yaml:**

```yaml
services:
  mongodb:
    image: mongo:6
    ports:
      - "${MONGODB_PORT:-27017}:27017"
    volumes:
      - mongodb_data:/data/db
    healthcheck:
      test: echo 'db.runCommand("ping").ok' | mongosh localhost:27017/test --quiet
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  mongodb_data:
```

**infrastructure/redis/compose.yaml:**

```yaml
services:
  redis:
    image: redis:7-alpine
    ports:
      - "${REDIS_PORT:-6379}:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  redis_data:
```

**infrastructure/localstack/compose.yaml:**

```yaml
services:
  localstack:
    image: localstack/localstack:latest
    ports:
      - "4566:4566"
    environment:
      SERVICES: s3,sqs,sns
      DEBUG: 0
      DATA_DIR: /var/lib/localstack/data
    volumes:
      - localstack_data:/var/lib/localstack
      - /var/run/docker.sock:/var/run/docker.sock
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:4566/_localstack/health"]
      interval: 5s
      timeout: 5s
      retries: 10

volumes:
  localstack_data:
```

---

## 5. CLI Commands & UX

### 5.1 Command Overview

```bash
localdev <command> [options] [services...]
```

| Command   | Description                              |
|-----------|------------------------------------------|
| `up`      | Start services and dependencies          |
| `down`    | Stop all running services                |
| `status`  | Show running services and health         |
| `logs`    | View aggregated or per-service logs      |
| `restart` | Restart specific service(s)              |
| `reset`   | Stop, clean volumes, restart fresh       |
| `config`  | Validate and show resolved config        |
| `init`    | Initialize localdev in current service   |
| `clone`   | Clone all registered service repos       |

---

### 5.2 Command Details

**Start services:**

```bash
# Start single service with all dependencies
$ localdev up service-a

# Start multiple services
$ localdev up service-a service-b

# Start with specific options
$ localdev up service-a --no-deps          # Only service-a, no dependencies
$ localdev up service-a --infra-only       # Only infrastructure, no services
$ localdev up service-a --build            # Force rebuild containers
$ localdev up service-a --local            # Run service locally (not in container)
```

**View status:**

```bash
$ localdev status

┌─────────────────────────────────────────────────────────────────┐
│ LocalDev Status                                                 │
├──────────────┬──────────┬───────────────────┬──────────────────┤
│ Name         │ Status   │ Endpoint          │ Health           │
├──────────────┼──────────┼───────────────────┼──────────────────┤
│ postgres     │ running  │ localhost:5432    │ healthy          │
│ mongodb      │ running  │ localhost:27017   │ healthy          │
│ localstack   │ running  │ localhost:4566    │ healthy          │
│ service-b    │ running  │ localhost:8081    │ healthy          │
│ service-a    │ running  │ localhost:8080    │ healthy          │
└──────────────┴──────────┴───────────────────┴──────────────────┘

AWS Resources (LocalStack):
  SQS Queues:    order-created-queue, payment-queue
  SNS Topics:    order-events
  S3 Buckets:    user-uploads
```

**View logs:**

```bash
# All services
$ localdev logs

# Specific service
$ localdev logs service-a

# Follow logs
$ localdev logs -f service-a

# Last N lines
$ localdev logs --tail 100 service-a
```

**Restart service:**

```bash
# Restart single service (keeps infra running)
$ localdev restart service-a

# Restart with rebuild
$ localdev restart service-a --build
```

**Reset environment:**

```bash
# Full reset (removes volumes, fresh start)
$ localdev reset

# Reset specific service data
$ localdev reset --service service-a
```

**Show resolved config:**

```bash
$ localdev config service-a

Resolved configuration for service-a:
────────────────────────────────────

Dependencies:
  Services: service-b
  Infrastructure: postgres, redis, localstack (sqs, sns, s3)

Environment Variables:
  APP_ENV=development
  DATABASE_URL=postgres://postgres:postgres@localhost:5432/service_a_db
  REDIS_URL=redis://localhost:6379
  SERVICE_B_URL=http://localhost:8081
  AWS_ENDPOINT=http://localhost:4566
  ORDER_QUEUE_URL=http://localhost:4566/000000000000/order-created-queue

AWS Resources:
  SQS:
    - order-created-queue
    - payment-queue
  SNS:
    - order-events → order-created-queue
  S3:
    - user-uploads
```

**Initialize new service:**

```bash
$ cd ~/projects/new-service
$ localdev init

Creating localdev.yaml...

? Service name: new-service
? Service type: (go/python/node) go
? Port: 8082
? Depends on services: (comma-separated) service-a
? Needs Postgres? (y/n) y
? Needs MongoDB? (y/n) n
? Needs Redis? (y/n) y
? Needs SQS? (y/n) y
? Needs S3? (y/n) n

Created localdev.yaml ✓
```

**Clone all repos:**

```bash
$ localdev clone

Cloning registered services...
  core-monolith ... ✓ ~/projects/core-monolith
  service-a ....... ✓ ~/projects/service-a
  service-b ....... ✓ ~/projects/service-b
  ai-service-1 .... ✓ ~/projects/ai-service-1
  ai-service-2 .... ✓ ~/projects/ai-service-2

All services cloned successfully.
```

---

## 6. Dependency Resolution Algorithm

```
┌─────────────────────────────────────────────────────────────────┐
│                   Dependency Resolution Flow                     │
└─────────────────────────────────────────────────────────────────┘

Input: localdev up service-a
                │
                ▼
┌─────────────────────────────────────────────────────────────────┐
│ 1. Load service-a/localdev.yaml                                 │
│    → requires: service-b, postgres, redis, sqs, s3              │
└─────────────────────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. Load service-b/localdev.yaml (dependency)                    │
│    → requires: mongodb, sns                                     │
└─────────────────────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. Build Dependency Graph                                       │
│                                                                 │
│    service-a ──depends──► service-b                             │
│        │                      │                                 │
│        ▼                      ▼                                 │
│    postgres              mongodb                                │
│    redis                 sns                                    │
│    sqs                                                          │
│    s3                                                           │
└─────────────────────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. Detect Circular Dependencies                                 │
│    → If found, error with clear message                         │
└─────────────────────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. Topological Sort (startup order)                             │
│                                                                 │
│    Level 0 (infra):  postgres, mongodb, redis, localstack       │
│    Level 1 (services): service-b                                │
│    Level 2 (services): service-a                                │
└─────────────────────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────────────┐
│ 6. Aggregate AWS Resources                                      │
│                                                                 │
│    SQS: order-created-queue, payment-queue (from service-a)     │
│    SNS: order-events (from service-a)                           │
│    S3:  user-uploads (from service-a)                           │
└─────────────────────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────────────┐
│ 7. Generate docker-compose.generated.yaml                       │
└─────────────────────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────────────┐
│ 8. Execute Startup Sequence                                     │
│                                                                 │
│    a. docker-compose up -d postgres mongodb redis localstack    │
│    b. Wait for health checks                                    │
│    c. Provision AWS resources (SQS, SNS, S3)                    │
│    d. docker-compose up -d service-b                            │
│    e. Wait for health check                                     │
│    f. docker-compose up -d service-a                            │
│    g. Wait for health check                                     │
│    h. Print summary                                             │
└─────────────────────────────────────────────────────────────────┘
```

---

## 7. Generated Docker Compose

Example `docker-compose.generated.yaml`:

```yaml
# AUTO-GENERATED by localdev - DO NOT EDIT
# Generated at: 2024-01-15T10:30:00Z
# Services: service-a, service-b

version: "3.8"

networks:
  localdev:
    driver: bridge

volumes:
  postgres_data:
  mongodb_data:
  redis_data:
  localstack_data:

services:
  # ============ Infrastructure ============
  
  postgres:
    image: postgres:15-alpine
    container_name: localdev-postgres
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./infrastructure/postgres/init:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - localdev

  mongodb:
    image: mongo:6
    container_name: localdev-mongodb
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db
    healthcheck:
      test: echo 'db.runCommand("ping").ok' | mongosh localhost:27017/test --quiet
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - localdev

  redis:
    image: redis:7-alpine
    container_name: localdev-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - localdev

  localstack:
    image: localstack/localstack:latest
    container_name: localdev-localstack
    ports:
      - "4566:4566"
    environment:
      SERVICES: s3,sqs,sns
      DEBUG: 0
    volumes:
      - localstack_data:/var/lib/localstack
      - /var/run/docker.sock:/var/run/docker.sock
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:4566/_localstack/health"]
      interval: 5s
      timeout: 5s
      retries: 10
    networks:
      - localdev

  # ============ Services ============
  
  service-b:
    build:
      context: /Users/dev/projects/service-b
      dockerfile: Dockerfile
    container_name: localdev-service-b
    ports:
      - "8081:8081"
    environment:
      APP_ENV: development
      MONGO_URL: mongodb://mongodb:27017/service_b_db
      AWS_ENDPOINT: http://localstack:4566
      AWS_ACCESS_KEY_ID: test
      AWS_SECRET_ACCESS_KEY: test
      AWS_REGION: us-east-1
    depends_on:
      mongodb:
        condition: service_healthy
      localstack:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8081/health"]
      interval: 5s
      timeout: 3s
      retries: 10
    networks:
      - localdev

  service-a:
    build:
      context: /Users/dev/projects/service-a
      dockerfile: Dockerfile
    container_name: localdev-service-a
    ports:
      - "8080:8080"
    environment:
      APP_ENV: development
      DATABASE_URL: postgres://postgres:postgres@postgres:5432/service_a_db
      REDIS_URL: redis://redis:6379
      SERVICE_B_URL: http://service-b:8081
      AWS_ENDPOINT: http://localstack:4566
      AWS_ACCESS_KEY_ID: test
      AWS_SECRET_ACCESS_KEY: test
      AWS_REGION: us-east-1
      ORDER_QUEUE_URL: http://localstack:4566/000000000000/order-created-queue
      S3_BUCKET: user-uploads
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      localstack:
        condition: service_healthy
      service-b:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 5s
      timeout: 3s
      retries: 10
    networks:
      - localdev
```

---

## 8. LocalStack Resource Provisioning

After LocalStack is healthy, the CLI provisions AWS resources.

**internal/provisioner/localstack.go (conceptual):**

```go
package provisioner

import (
    "context"
    "fmt"
    
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/service/sns"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

type AWSResources struct {
    SQS struct {
        Queues []QueueConfig
    }
    SNS struct {
        Topics []TopicConfig
    }
    S3 struct {
        Buckets []BucketConfig
    }
}

type QueueConfig struct {
    Name string
    DLQ  bool
}

type TopicConfig struct {
    Name          string
    Subscriptions []SubscriptionConfig
}

type SubscriptionConfig struct {
    Queue string
}

type BucketConfig struct {
    Name     string
    SeedPath string
}

func ProvisionAWSResources(ctx context.Context, resources AWSResources, endpoint string) error {
    cfg := createLocalStackConfig(endpoint)
    
    sqsClient := sqs.NewFromConfig(cfg)
    snsClient := sns.NewFromConfig(cfg)
    s3Client := s3.NewFromConfig(cfg)
    
    queueArns := make(map[string]string)
    
    // Create SQS Queues
    for _, queue := range resources.SQS.Queues {
        // Create DLQ first if requested
        if queue.DLQ {
            dlqName := queue.Name + "-dlq"
            _, err := sqsClient.CreateQueue(ctx, &sqs.CreateQueueInput{
                QueueName: aws.String(dlqName),
            })
            if err != nil {
                return fmt.Errorf("failed to create DLQ %s: %w", dlqName, err)
            }
        }
        
        result, err := sqsClient.CreateQueue(ctx, &sqs.CreateQueueInput{
            QueueName: aws.String(queue.Name),
        })
        if err != nil {
            return fmt.Errorf("failed to create queue %s: %w", queue.Name, err)
        }
        
        // Get queue ARN for SNS subscription
        attrs, _ := sqsClient.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
            QueueUrl:       result.QueueUrl,
            AttributeNames: []string{"QueueArn"},
        })
        queueArns[queue.Name] = attrs.Attributes["QueueArn"]
    }
    
    // Create SNS Topics
    for _, topic := range resources.SNS.Topics {
        result, err := snsClient.CreateTopic(ctx, &sns.CreateTopicInput{
            Name: aws.String(topic.Name),
        })
        if err != nil {
            return fmt.Errorf("failed to create topic %s: %w", topic.Name, err)
        }
        
        // Subscribe SQS queues to topic
        for _, sub := range topic.Subscriptions {
            queueArn, ok := queueArns[sub.Queue]
            if !ok {
                return fmt.Errorf("queue %s not found for subscription", sub.Queue)
            }
            
            _, err := snsClient.Subscribe(ctx, &sns.SubscribeInput{
                TopicArn: result.TopicArn,
                Protocol: aws.String("sqs"),
                Endpoint: aws.String(queueArn),
            })
            if err != nil {
                return fmt.Errorf("failed to subscribe %s to %s: %w", sub.Queue, topic.Name, err)
            }
        }
    }
    
    // Create S3 Buckets
    for _, bucket := range resources.S3.Buckets {
        _, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
            Bucket: aws.String(bucket.Name),
        })
        if err != nil {
            return fmt.Errorf("failed to create bucket %s: %w", bucket.Name, err)
        }
        
        // Upload seed files if specified
        if bucket.SeedPath != "" {
            if err := uploadSeedFiles(ctx, s3Client, bucket.Name, bucket.SeedPath); err != nil {
                return fmt.Errorf("failed to seed bucket %s: %w", bucket.Name, err)
            }
        }
    }
    
    return nil
}
```

---

## 9. Go Structs for Config Schema

**internal/config/schema.go:**

```go
package config

import "time"

// ServiceConfig represents the localdev.yaml in each service
type ServiceConfig struct {
    Version  string        `yaml:"version"`
    Service  ServiceInfo   `yaml:"service"`
    Requires Requirements  `yaml:"requires"`
    Env      map[string]string `yaml:"env"`
    EnvRefs  map[string]string `yaml:"env_refs"`
}

type ServiceInfo struct {
    Name   string       `yaml:"name"`
    Type   string       `yaml:"type"`   // go, python, node
    Port   int          `yaml:"port"`
    Build  *BuildConfig `yaml:"build,omitempty"`
    Run    *RunConfig   `yaml:"run,omitempty"`
    Health HealthConfig `yaml:"health"`
}

type BuildConfig struct {
    Dockerfile string `yaml:"dockerfile"`
    Context    string `yaml:"context"`
}

type RunConfig struct {
    Command   string `yaml:"command"`
    HotReload bool   `yaml:"hot_reload"`
}

type HealthConfig struct {
    Endpoint string        `yaml:"endpoint"`
    Interval time.Duration `yaml:"interval"`
    Timeout  time.Duration `yaml:"timeout"`
    Retries  int           `yaml:"retries"`
}

type Requirements struct {
    Services       []string              `yaml:"services"`
    Infrastructure InfrastructureConfig  `yaml:"infrastructure"`
}

type InfrastructureConfig struct {
    Postgres   *PostgresConfig   `yaml:"postgres,omitempty"`
    MongoDB    *MongoDBConfig    `yaml:"mongodb,omitempty"`
    Redis      interface{}       `yaml:"redis,omitempty"`      // bool or RedisConfig
    SQS        *SQSConfig        `yaml:"sqs,omitempty"`
    SNS        *SNSConfig        `yaml:"sns,omitempty"`
    S3         *S3Config         `yaml:"s3,omitempty"`
}

type PostgresConfig struct {
    Database   string `yaml:"database"`
    Migrations string `yaml:"migrations,omitempty"`
    Seed       string `yaml:"seed,omitempty"`
}

type MongoDBConfig struct {
    Database string `yaml:"database"`
    Seed     string `yaml:"seed,omitempty"`
}

type RedisConfig struct {
    // Future: specific redis config if needed
}

type SQSConfig struct {
    Queues []QueueConfig `yaml:"queues"`
}

type QueueConfig struct {
    Name string `yaml:"name"`
    DLQ  bool   `yaml:"dlq,omitempty"`
}

type SNSConfig struct {
    Topics []TopicConfig `yaml:"topics"`
}

type TopicConfig struct {
    Name          string               `yaml:"name"`
    Subscriptions []SubscriptionConfig `yaml:"subscriptions,omitempty"`
}

type SubscriptionConfig struct {
    Queue string `yaml:"queue"`
}

type S3Config struct {
    Buckets []BucketConfig `yaml:"buckets"`
}

type BucketConfig struct {
    Name string `yaml:"name"`
    Seed string `yaml:"seed,omitempty"`
}

// ServiceRegistry represents the services.yaml in the orchestration repo
type ServiceRegistry struct {
    Version      string                     `yaml:"version"`
    Services     map[string]ServiceEntry    `yaml:"services"`
    PathDefaults PathDefaults               `yaml:"path_defaults"`
}

type ServiceEntry struct {
    Repo string `yaml:"repo"`
    Path string `yaml:"path"`
}

type PathDefaults struct {
    Base string `yaml:"base"`
}
```

---

## 10. Implementation Phases

### Phase 1: Foundation (Week 1-2)

- [ ] Project setup with Go modules
- [ ] CLI skeleton using cobra
- [ ] Config parsing (localdev.yaml schema)
- [ ] Service registry loading (services.yaml)
- [ ] Basic `localdev up` with single service (no dependencies)
- [ ] Docker Compose generation (basic)
- [ ] Infrastructure catalog (postgres, mongodb, redis)

### Phase 2: Dependency Resolution (Week 3)

- [ ] Dependency graph building
- [ ] Topological sort algorithm
- [ ] Circular dependency detection
- [ ] Multi-service `localdev up`
- [ ] `localdev down` command

### Phase 3: AWS/LocalStack Integration (Week 4)

- [ ] LocalStack infrastructure definition
- [ ] SQS queue provisioning
- [ ] SNS topic provisioning
- [ ] S3 bucket provisioning
- [ ] SNS → SQS subscriptions
- [ ] Environment variable resolution (env_refs)

### Phase 4: Developer Experience (Week 5)

- [ ] `localdev status` with health checks
- [ ] `localdev logs` (aggregated and per-service)
- [ ] `localdev restart` command
- [ ] `localdev reset` command
- [ ] `localdev config` (show resolved config)

### Phase 5: Polish & Onboarding (Week 6)

- [ ] `localdev init` (interactive setup)
- [ ] `localdev clone` (clone all repos)
- [ ] Error messages and troubleshooting hints
- [ ] Documentation
- [ ] Team onboarding guide

### Phase 6: Advanced Features (Future)

- [ ] Hot reload mode (`--local` flag)
- [ ] Profiles for common flows
- [ ] Resource usage monitoring
- [ ] Parallel startup optimization
- [ ] Plugin system for custom infrastructure

---

## 11. Key Libraries to Use

| Purpose | Library | Notes |
|---------|---------|-------|
| CLI Framework | `github.com/spf13/cobra` | Industry standard for Go CLIs |
| YAML Parsing | `gopkg.in/yaml.v3` | Native YAML support |
| Docker Compose | `github.com/compose-spec/compose-go` | Parse/generate compose files |
| AWS SDK | `github.com/aws/aws-sdk-go-v2` | For LocalStack provisioning |
| Terminal UI | `github.com/charmbracelet/bubbles` | Spinners, tables, colors |
| Progress | `github.com/schollz/progressbar/v3` | Progress indicators |
| Logging | `github.com/rs/zerolog` | Structured logging |

---

## 12. Success Metrics

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Full environment startup | < 5 minutes | Time from `localdev up` to all healthy |
| RAM usage | < 50% of available (8-12GB) | `docker stats` |
| New developer onboarding | < 30 minutes | First successful `localdev up` |
| Config errors caught | 100% before startup | Validation in `localdev config` |
| Service restart time | < 30 seconds | Single service restart |

---

## 13. Example Workflow

### Developer Day 1: Onboarding

```bash
# 1. Install localdev (one-time)
$ brew install yourorg/tap/localdev

# 2. Clone orchestration repo
$ git clone git@github.com:yourorg/localdev-orchestration.git
$ cd localdev-orchestration

# 3. Clone all service repos
$ localdev clone

# 4. Start working on service-a
$ localdev up service-a

# 5. Verify everything is running
$ localdev status
```

### Developer Day N: Normal Development

```bash
# Start my service and dependencies
$ localdev up service-a

# Make code changes...

# Restart just my service
$ localdev restart service-a

# View logs
$ localdev logs -f service-a

# End of day
$ localdev down
```

---

## 14. Error Handling Examples

**Circular Dependency:**

```
$ localdev up service-a

Error: Circular dependency detected

  service-a → service-b → service-c → service-a
              ↑___________________________|

Please review the dependency chain and remove the cycle.
```

**Missing Service:**

```
$ localdev up service-a

Error: Service 'service-b' not found

service-a depends on 'service-b' but it is not registered in services.yaml

Registered services:
  - core-monolith
  - service-a
  - ai-service-1
  - ai-service-2

Did you mean to add 'service-b' to services.yaml?
```

**Missing localdev.yaml:**

```
$ localdev up service-a

Error: No localdev.yaml found for 'service-a'

Expected location: /Users/dev/projects/service-a/localdev.yaml

Run 'localdev init' in the service directory to create one.
```

**Port Conflict:**

```
$ localdev up service-a

Error: Port conflict detected

  Port 8080 is required by both:
    - service-a
    - core-monolith

Please update one of the services to use a different port.
```

---

## 15. Future Considerations

### Profiles (Future Enhancement)

```yaml
# services.yaml
profiles:
  checkout-flow:
    services:
      - service-a
      - service-b
      - core-monolith
      
  ai-pipeline:
    services:
      - ai-service-1
      - ai-service-2
```

```bash
$ localdev up --profile checkout-flow
```

### Remote Development (Future Enhancement)

```bash
# Connect to shared development environment
$ localdev connect staging

# Run local service against remote dependencies
$ localdev up service-a --remote-deps
```

### Resource Monitoring (Future Enhancement)

```bash
$ localdev stats

┌────────────────────────────────────────────────────┐
│ Resource Usage                                     │
├──────────────┬─────────┬─────────┬────────────────┤
│ Container    │ CPU     │ Memory  │ Network I/O    │
├──────────────┼─────────┼─────────┼────────────────┤
│ postgres     │ 0.5%    │ 256 MB  │ 1.2 MB/s       │
│ mongodb      │ 0.3%    │ 512 MB  │ 0.8 MB/s       │
│ localstack   │ 1.2%    │ 1.1 GB  │ 2.1 MB/s       │
│ service-a    │ 2.1%    │ 128 MB  │ 0.5 MB/s       │
│ service-b    │ 1.8%    │ 96 MB   │ 0.3 MB/s       │
├──────────────┼─────────┼─────────┼────────────────┤
│ Total        │ 5.9%    │ 2.1 GB  │ 4.9 MB/s       │
└──────────────┴─────────┴─────────┴────────────────┘
```

---

## 16. Appendix: Quick Reference

### localdev.yaml Minimal Example

```yaml
version: "1"

service:
  name: my-service
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
  infrastructure:
    postgres:
      database: my_service_db

env:
  APP_ENV: development

env_refs:
  DATABASE_URL: "postgres://postgres:postgres@${postgres.host}:${postgres.port}/${self.postgres.database}"
```

### services.yaml Minimal Example

```yaml
version: "1"

services:
  my-service:
    repo: git@github.com:yourorg/my-service.git
    path: ~/projects/my-service

path_defaults:
  base: ~/projects
```

### Common Commands

```bash
localdev up <service>           # Start service with dependencies
localdev up <svc1> <svc2>       # Start multiple services
localdev down                   # Stop everything
localdev status                 # Show running services
localdev logs <service>         # View service logs
localdev logs -f <service>      # Follow logs
localdev restart <service>      # Restart single service
localdev reset                  # Full reset with clean volumes
localdev config <service>       # Show resolved configuration
localdev init                   # Initialize localdev.yaml
localdev clone                  # Clone all registered repos
```
