# Grund Configuration Reference

This document describes all configuration files used by Grund.

## Configuration Files Overview

| File | Location | Purpose |
|------|----------|---------|
| `services.yaml` | Orchestration repo root | Service registry - lists all services |
| `grund.yaml` | Each service directory | Service configuration |
| `config.yaml` | `~/.grund/config.yaml` | Global user settings |
| `docker-compose.generated.yaml` | Orchestration repo root | Auto-generated, do not edit |

---

## services.yaml (Service Registry)

The service registry maps service names to their repository locations.

### Schema

```yaml
version: "1"

services:
  <service-name>:
    repo: <git-url>       # Git repository URL
    path: <local-path>    # Local path to service directory
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | string | Yes | Schema version (currently "1") |
| `services` | map | Yes | Map of service names to locations |
| `services.<name>.repo` | string | No | Git repository URL for cloning |
| `services.<name>.path` | string | Yes | Local filesystem path to service |

### Example

```yaml
version: "1"

services:
  user-service:
    repo: git@github.com:company/user-service.git
    path: /Users/dev/projects/user-service

  order-service:
    repo: git@github.com:company/order-service.git
    path: /Users/dev/projects/order-service

  notification-service:
    path: /Users/dev/projects/notification-service

  # Services without repo can still be used if path exists
  legacy-api:
    path: ../legacy-api
```

### Path Resolution

- **Absolute paths**: Used as-is
- **Relative paths**: Resolved relative to services.yaml location

---

## grund.yaml (Service Configuration)

Each service has a `grund.yaml` file in its root directory.

### Complete Schema

```yaml
version: "1"

service:
  name: <service-name>
  type: <go|python|node>
  port: <port-number>

  build:
    dockerfile: <path-to-dockerfile>
    context: <build-context>

  # OR for interpreted languages
  run:
    command: <run-command>
    hot_reload: <boolean>

  health:
    endpoint: <health-endpoint>
    interval: <duration>
    timeout: <duration>
    retries: <number>

requires:
  services:
    - <service-name>
    - <service-name>

  infrastructure:
    postgres:
      database: <db-name>
      migrations: <migrations-path>
      seed: <seed-file-path>

    mongodb:
      database: <db-name>
      seed: <seed-file-path>

    redis: true

    sqs:
      queues:
        - name: <queue-name>
          dlq: <boolean>

    sns:
      topics:
        - name: <topic-name>
          subscriptions:
            - queue: <queue-name>

    s3:
      buckets:
        - name: <bucket-name>
          seed: <seed-directory>

# Static environment variables
env:
  KEY: value
  ANOTHER_KEY: value

# Dynamic environment variable references
env_refs:
  DATABASE_URL: "postgres://postgres:postgres@${postgres.host}:${postgres.port}/${self.postgres.database}"
  REDIS_URL: "redis://${redis.host}:${redis.port}"
```

### Service Section

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Service name (must match services.yaml) |
| `type` | string | Yes | `go`, `python`, or `node` |
| `port` | integer | Yes | Port the service listens on (1-65535) |

### Build Section

For compiled/containerized services.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `dockerfile` | string | Yes | Path to Dockerfile |
| `context` | string | Yes | Build context directory (usually `.`) |

### Run Section

For interpreted languages (alternative to build).

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `command` | string | Yes | Command to run the service |
| `hot_reload` | boolean | No | Enable hot reloading |

### Health Section

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `endpoint` | string | Yes | | Health check URL path |
| `interval` | duration | No | `5s` | Time between health checks |
| `timeout` | duration | No | `3s` | Timeout for each check |
| `retries` | integer | No | `10` | Retries before unhealthy |

### Requires Section

#### Services

List of other services this service depends on.

```yaml
requires:
  services:
    - user-service
    - auth-service
```

#### Infrastructure

##### PostgreSQL

```yaml
infrastructure:
  postgres:
    database: myservice_db     # Required: database name
    migrations: ./migrations   # Optional: path to migration files
    seed: ./seed.sql          # Optional: path to seed data
```

##### MongoDB

```yaml
infrastructure:
  mongodb:
    database: myservice_db     # Required: database name
    seed: ./seed.json         # Optional: path to seed data
```

##### Redis

```yaml
infrastructure:
  redis: true                  # Just needs to be present
```

##### SQS (Simple Queue Service)

```yaml
infrastructure:
  sqs:
    queues:
      - name: orders           # Queue name
        dlq: true              # Create dead-letter queue
      - name: notifications
        dlq: false
```

##### SNS (Simple Notification Service)

```yaml
infrastructure:
  sns:
    topics:
      - name: order-events
        subscriptions:
          - queue: orders      # Subscribe SQS queue to topic
          - queue: analytics
      - name: user-events
```

##### S3 (Simple Storage Service)

```yaml
infrastructure:
  s3:
    buckets:
      - name: uploads
        seed: ./seed-data      # Optional: directory to upload
      - name: documents
```

### Environment Variables

#### Static Variables (`env`)

```yaml
env:
  APP_ENV: development
  LOG_LEVEL: debug
  FEATURE_FLAG: "true"
```

#### Dynamic References (`env_refs`)

Use `${placeholder}` syntax for values resolved at startup.

```yaml
env_refs:
  # Database connections
  DATABASE_URL: "postgres://postgres:postgres@${postgres.host}:${postgres.port}/${self.postgres.database}"
  MONGO_URL: "mongodb://${mongodb.host}:${mongodb.port}/${self.mongodb.database}"
  REDIS_URL: "redis://${redis.host}:${redis.port}"

  # AWS/LocalStack
  AWS_ENDPOINT: "${localstack.endpoint}"
  AWS_REGION: "${localstack.region}"

  # Queue URLs
  ORDERS_QUEUE_URL: "${sqs.orders.url}"
  ORDERS_DLQ_URL: "${sqs.orders.dlq}"

  # Topic ARNs
  EVENTS_TOPIC_ARN: "${sns.order-events.arn}"

  # Service dependencies
  USER_SERVICE_URL: "http://${user-service.host}:${user-service.port}"
  AUTH_SERVICE_URL: "http://${auth-service.host}:${auth-service.port}"
```

### Complete Example

```yaml
version: "1"

service:
  name: order-service
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
    - user-service
    - notification-service
  infrastructure:
    postgres:
      database: orders_db
      migrations: ./migrations
    redis: true
    sqs:
      queues:
        - name: orders
          dlq: true
        - name: order-confirmations
          dlq: false
    sns:
      topics:
        - name: order-events
          subscriptions:
            - queue: order-confirmations

env:
  APP_ENV: development
  LOG_LEVEL: debug

env_refs:
  DATABASE_URL: "postgres://postgres:postgres@${postgres.host}:${postgres.port}/${self.postgres.database}"
  REDIS_URL: "redis://${redis.host}:${redis.port}"
  ORDERS_QUEUE_URL: "${sqs.orders.url}"
  ORDER_EVENTS_TOPIC_ARN: "${sns.order-events.arn}"
  USER_SERVICE_URL: "http://${user-service.host}:${user-service.port}"
  NOTIFICATION_SERVICE_URL: "http://${notification-service.host}:${notification-service.port}"
```

---

## ~/.grund/config.yaml (Global Configuration)

User-level settings for Grund.

### Schema

```yaml
# Default filename to search for
default_services_file: services.yaml

# Default orchestration repository path
default_orchestration_repo: ~/projects/orchestration

# Where to clone service repositories
services_base_path: ~/projects

# Docker settings
docker:
  compose_command: docker compose

# LocalStack settings
localstack:
  endpoint: http://localhost:4566
  region: us-east-1
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `default_services_file` | string | `services.yaml` | Filename to search for |
| `default_orchestration_repo` | string | | Path to orchestration repo |
| `services_base_path` | string | `~/projects` | Where to clone services |
| `docker.compose_command` | string | `docker compose` | Docker compose command |
| `localstack.endpoint` | string | `http://localhost:4566` | LocalStack endpoint |
| `localstack.region` | string | `us-east-1` | AWS region for LocalStack |

### Path Expansion

- `~` is expanded to home directory
- Relative paths resolved from config file location

---

## docker-compose.generated.yaml (Auto-Generated)

Generated by `grund up`. **Do not edit manually.**

### Example Generated Output

```yaml
# AUTO-GENERATED by grund - DO NOT EDIT
# Regenerate with: grund up <services>

version: "3.8"

services:
  postgres:
    image: postgres:15-alpine
    container_name: grund-postgres
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: orders_db
    volumes:
      - postgres-data:/var/lib/postgresql/data
    networks:
      - grund-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: grund-redis
    ports:
      - "6379:6379"
    networks:
      - grund-network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

  localstack:
    image: localstack/localstack:latest
    container_name: grund-localstack
    ports:
      - "4566:4566"
    environment:
      SERVICES: sqs,sns
      DEBUG: "0"
      AWS_DEFAULT_REGION: us-east-1
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - localstack-data:/var/lib/localstack
    networks:
      - grund-network
    healthcheck:
      test: ["CMD-SHELL", "curl -s http://localhost:4566/_localstack/health"]
      interval: 10s
      timeout: 5s
      retries: 10
      start_period: 20s

  order-service:
    container_name: grund-order-service
    build:
      context: /Users/dev/projects/order-service
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      APP_ENV: development
      DATABASE_URL: postgres://postgres:postgres@postgres:5432/orders_db
      REDIS_URL: redis://redis:6379
      ORDERS_QUEUE_URL: http://localstack:4566/000000000000/orders
    networks:
      - grund-network
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      localstack:
        condition: service_healthy
    healthcheck:
      test: ["CMD-SHELL", "curl -sf http://localhost:8080/health || exit 1"]
      interval: 5s
      timeout: 3s
      retries: 10

networks:
  grund-network:
    driver: bridge

volumes:
  postgres-data: {}
  localstack-data: {}
```

---

## Environment Variable Reference

### Available Placeholders

| Category | Placeholder | Description |
|----------|-------------|-------------|
| **PostgreSQL** | `${postgres.host}` | Container hostname (`postgres`) |
| | `${postgres.port}` | Port (`5432`) |
| | `${postgres.database}` | Database name |
| | `${postgres.username}` | Username (`postgres`) |
| | `${postgres.password}` | Password (`postgres`) |
| **MongoDB** | `${mongodb.host}` | Container hostname (`mongodb`) |
| | `${mongodb.port}` | Port (`27017`) |
| | `${mongodb.database}` | Database name |
| **Redis** | `${redis.host}` | Container hostname (`redis`) |
| | `${redis.port}` | Port (`6379`) |
| **LocalStack** | `${localstack.endpoint}` | Full endpoint URL |
| | `${localstack.host}` | Hostname (`localstack`) |
| | `${localstack.port}` | Port (`4566`) |
| | `${localstack.region}` | AWS region |
| | `${localstack.account_id}` | Account ID (`000000000000`) |
| **SQS** | `${sqs.<name>.url}` | Queue URL |
| | `${sqs.<name>.arn}` | Queue ARN |
| | `${sqs.<name>.dlq}` | Dead-letter queue URL |
| | `${sqs.<name>.name}` | Queue name |
| **SNS** | `${sns.<name>.arn}` | Topic ARN |
| | `${sns.<name>.name}` | Topic name |
| **S3** | `${s3.<name>.url}` | Bucket URL |
| | `${s3.<name>.name}` | Bucket name |
| **Services** | `${<service>.host}` | Service container name |
| | `${<service>.port}` | Service port |
| **Self** | `${self.host}` | Current service name |
| | `${self.port}` | Current service port |
| | `${self.postgres.database}` | This service's database |
| | `${self.mongodb.database}` | This service's MongoDB |

---

## See Also

- [CLI Commands](./cli-commands.md)
- [Algorithms](./algorithms.md)
- [Architecture Overview](./architecture.md)
- [Adding New Infrastructure](./adding-new-infrastructure.md)
