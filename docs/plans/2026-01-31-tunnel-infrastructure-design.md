# Tunnel Infrastructure for S3/LLM Integration

**Date:** 2026-01-31
**Status:** Draft
**Author:** Vivek Kundariya

## Summary

Add tunnel infrastructure support to Grund, enabling services to expose local endpoints (LocalStack, APIs, etc.) to the internet via cloudflared or ngrok. This enables use cases like passing S3 presigned URLs to cloud LLMs.

## Motivation

When using LocalStack for local S3 development, presigned URLs point to `localhost:4566` which cloud LLMs (OpenAI, Anthropic) cannot access. By tunneling LocalStack to a public URL, applications can generate presigned URLs that LLMs can fetch.

## Design

### Configuration Schema

Add tunnel as a new infrastructure type in `grund.yaml`:

```yaml
version: "1"

service:
  name: my-service
  type: go
  port: 8080

requires:
  infrastructure:
    s3:
      buckets:
        - name: uploads
    tunnel:
      provider: cloudflared  # or "ngrok"
      targets:
        - name: localstack
          host: ${localstack.host}
          port: ${localstack.port}
        - name: my-api
          host: localhost
          port: 8080

env_refs:
  AWS_ENDPOINT: "${localstack.endpoint}"
  AWS_PUBLIC_ENDPOINT: "${tunnel.localstack.url}"
  S3_PUBLIC_BASE: "${tunnel.localstack.url}/uploads"
```

### New Config Structs

```go
type TunnelConfig struct {
    Provider string         `yaml:"provider"` // "cloudflared" or "ngrok"
    Targets  []TunnelTarget `yaml:"targets"`
}

type TunnelTarget struct {
    Name string `yaml:"name"` // identifier for ${tunnel.<name>.url}
    Host string `yaml:"host"` // supports placeholders
    Port string `yaml:"port"` // supports placeholders (string for interpolation)
}
```

### New Environment Placeholders

| Placeholder | Description | Example Value |
|-------------|-------------|---------------|
| `${tunnel.<name>.url}` | Full public HTTPS URL | `https://random-abc.trycloudflare.com` |
| `${tunnel.<name>.host}` | Hostname portion only | `random-abc.trycloudflare.com` |

### Tunnel Lifecycle

**Startup sequence (during `grund up`):**

1. Start infrastructure (LocalStack, Postgres, etc.)
2. Resolve tunnel target placeholders (`${localstack.host}` → `localhost`)
3. Start tunnel process for each target
4. Wait for tunnel URLs to be available (30s timeout)
5. Store tunnel URLs in environment context
6. Resolve `env_refs` including `${tunnel.*}` placeholders
7. Start services with fully resolved environment

**Shutdown sequence (during `grund down`):**

1. Stop services
2. Stop tunnel processes
3. Stop infrastructure

### Provider Implementation

**Interface:**

```go
type TunnelProvider interface {
    Start(target TunnelTarget) (*Tunnel, error)
    Stop(tunnel *Tunnel) error
    GetURL(tunnel *Tunnel) string
}

type Tunnel struct {
    Name      string
    PublicURL string
    Process   *os.Process
}
```

**Cloudflared:**

```bash
cloudflared tunnel --url http://localhost:4566
```

- No account required for quick tunnels
- Outputs URL: `https://random-words.trycloudflare.com`
- Parse stdout for URL pattern

**ngrok:**

```bash
ngrok http 4566 --log stdout --log-format json
```

- Free tier: 1 tunnel, session limits
- Requires auth token for multiple tunnels
- Parse JSON log for public URL

### Error Handling

**Startup failures:**

| Scenario | Behavior |
|----------|----------|
| Provider binary missing | Error with install command |
| Tunnel fails to start | Retry 3 times, then fail `grund up` |
| URL not received in 30s | Timeout error, cleanup, fail |
| Port already in use | Clear error: "Port 4566 not accessible" |

**Runtime failures:**

| Scenario | Behavior |
|----------|----------|
| Tunnel process dies | Log warning, services continue |
| `grund down` with dead tunnel | Graceful cleanup, no error |

**Placeholder resolution order:**

1. Infrastructure placeholders (`${localstack.host}`) resolved first
2. Tunnel target configs use resolved infra values
3. Tunnel starts, URLs captured
4. `${tunnel.*}` placeholders now available
5. `env_refs` resolved with all placeholders

**Validation:**

- Provider must be `cloudflared` or `ngrok`
- Each target must have unique `name`
- `host` and `port` required for each target
- Warn if target port is unreachable before tunneling

## Implementation

### Files to Modify

| File | Changes |
|------|---------|
| `internal/config/schema.go` | Add `TunnelConfig`, `TunnelTarget` structs |
| `internal/infrastructure/generator/env_resolver.go` | Add `resolveTunnel()` for `${tunnel.*}` |
| `internal/application/ports/compose.go` | Add `TunnelContext` to `EnvironmentContext` |
| `internal/application/commands/up_command.go` | Start tunnels in lifecycle |
| `internal/cli/up.go` | Handle tunnel status output |

### New Files

| File | Purpose |
|------|---------|
| `internal/infrastructure/tunnel/provider.go` | `TunnelProvider` interface |
| `internal/infrastructure/tunnel/cloudflared.go` | Cloudflared implementation |
| `internal/infrastructure/tunnel/ngrok.go` | ngrok implementation |
| `internal/infrastructure/tunnel/manager.go` | Lifecycle management |

### CLI Changes

- `grund up` - starts tunnels after infrastructure, before services
- `grund down` - stops tunnels after services
- `grund status` - shows tunnel status and public URLs

### Testing Strategy

- Unit tests for URL parsing (mock process output)
- Integration tests with actual cloudflared (if installed)
- E2E test: start tunnel → verify URL accessible → stop

## Usage Example

**Application code (Go with AWS SDK):**

```go
// Use the public endpoint for presigned URLs
publicEndpoint := os.Getenv("AWS_PUBLIC_ENDPOINT")

cfg, _ := config.LoadDefaultConfig(ctx,
    config.WithEndpointResolver(aws.EndpointResolverFunc(
        func(service, region string) (aws.Endpoint, error) {
            return aws.Endpoint{URL: publicEndpoint}, nil
        },
    )),
)

s3Client := s3.NewFromConfig(cfg)
presigner := s3.NewPresignClient(s3Client)

// Generate presigned URL using public endpoint
presignedURL, _ := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
    Bucket: aws.String("uploads"),
    Key:    aws.String("document.pdf"),
}, s3.WithPresignExpires(15*time.Minute))

// This URL is now accessible by cloud LLMs
fmt.Println(presignedURL.URL)
// https://random-abc.trycloudflare.com/uploads/document.pdf?X-Amz-...
```

## Alternatives Considered

**Approach B: Tunnel as LocalStack Option**

Nest tunnel config under S3/SQS/SNS. Rejected because:
- Less flexible (can't tunnel non-AWS services)
- Tighter coupling
- Doesn't follow existing Grund patterns

## Open Questions

None - design validated through discussion.
