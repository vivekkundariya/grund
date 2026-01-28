# SNS Subscription Filter Policy Support

## Overview

Add support for AWS SNS subscription attributes (FilterPolicy, FilterPolicyScope, RawMessageDelivery, etc.) to grund, using AWS SDK structure directly.

## Configuration Schema

```yaml
requires:
  infrastructure:
    sqs:
      queues:
        - name: jupiter-general-tasks-consumer
        - name: jupiter-ai-response-consumer

    sns:
      topics:
        - name: mrf-event-bus
          subscriptions:
            - protocol: sqs
              endpoint: "${sqs.jupiter-general-tasks-consumer.arn}"
              attributes:
                FilterPolicy: '{"source":["MARS"],"task_type":["TASK_GENERATION"]}'
                FilterPolicyScope: MessageBody

            - protocol: sqs
              endpoint: "${sqs.jupiter-ai-response-consumer.arn}"
              attributes:
                FilterPolicy: '{"source":["JUPITER"]}'
                FilterPolicyScope: MessageAttributes
                RawMessageDelivery: "true"
```

## Struct Changes

### `internal/config/schema.go`

```go
type SubscriptionConfig struct {
    Protocol   string            `yaml:"protocol"`
    Endpoint   string            `yaml:"endpoint"`
    Attributes map[string]string `yaml:"attributes,omitempty"`
}
```

### `internal/domain/infrastructure/infrastructure.go`

```go
type SubscriptionConfig struct {
    Protocol   string
    Endpoint   string            // Template: "${sqs.queue-name.arn}"
    Attributes map[string]string // AWS subscription attributes
}
```

## Provisioning Flow

1. Create SQS queues → collect ARNs
2. Build `EnvironmentContext` with queue ARNs
3. Create SNS topics
4. For each subscription:
   - Resolve `${sqs.X.arn}` using existing `EnvironmentResolver`
   - Call `sns.Subscribe()` with resolved endpoint
   - Call `sns.SetSubscriptionAttributes()` for each attribute

## Files to Modify

| File | Change |
|:-----|:-------|
| `internal/config/schema.go` | Update `SubscriptionConfig` struct |
| `internal/domain/infrastructure/infrastructure.go` | Update domain `SubscriptionConfig` |
| `internal/infrastructure/aws/provisioner.go` | Add template resolution + `SetSubscriptionAttributes` |
| `docs/wiki/configuration.md` | Update SNS documentation |

## Breaking Change

Old syntax no longer supported:

```yaml
# Old (removed)
subscriptions:
  - queue: my-queue

# New (required)
subscriptions:
  - protocol: sqs
    endpoint: "${sqs.my-queue.arn}"
```

## Template Resolution

Reuses existing `EnvironmentResolverImpl` from `internal/infrastructure/generator/env_resolver.go`.

Supported templates:
- `${sqs.<queue-name>.arn}` → Queue ARN
- `${sqs.<queue-name>.url}` → Queue URL
