# Secrets Management Design

**Date:** 2026-01-28
**Status:** Approved

## Problem

Developers store secrets (API keys, LLM keys) in `.env` files or shell environment, but grund doesn't consume these when starting services. Services fail with missing credentials.

## Solution

Explicit secrets declaration in `grund.yaml` with centralized storage in `~/.grund/secrets.env`.

## Configuration Schema

**grund.yaml:**

```yaml
secrets:
  OPENAI_API_KEY:
    description: "OpenAI API key for embeddings"
    required: true
  STRIPE_SECRET_KEY:
    description: "Stripe secret key for payments"
    required: true
  ANALYTICS_KEY:
    description: "Mixpanel key for tracking"
    required: false
```

**~/.grund/secrets.env:**

```bash
OPENAI_API_KEY=sk-xxx
STRIPE_SECRET_KEY=sk_test_xxx
```

## Resolution Order

Priority (highest first):
1. `~/.grund/secrets.env` file
2. Shell environment

## CLI Commands

### `grund secrets list <service>`

Shows all secrets required by target services and their status:

```
$ grund secrets list user-service

Secrets for user-service (and dependencies):

  OPENAI_API_KEY      ✓ found    "OpenAI API key for embeddings"
  STRIPE_SECRET_KEY   ✗ missing  "Stripe secret key for payments"
  ANALYTICS_KEY       ○ optional "Mixpanel key for tracking" (not set)

Source: ~/.grund/secrets.env (3 keys loaded)

Missing required secrets: 1
Run 'grund secrets init' to generate a template.
```

### `grund secrets init <service>`

Generates `~/.grund/secrets.env` with placeholders:

```
$ grund secrets init user-service

Created ~/.grund/secrets.env with 2 placeholders:

  OPENAI_API_KEY=     # OpenAI API key for embeddings
  STRIPE_SECRET_KEY=  # Stripe secret key for payments

Edit the file and add your values.
```

If file exists, appends missing keys without overwriting existing values.

## Fail-Fast Behavior

**During `grund up`:**

```
$ grund up user-service

Checking secrets...
  ✗ STRIPE_SECRET_KEY is required but not found

Error: Missing required secrets

Run 'grund secrets list user-service' to see all required secrets.
Run 'grund secrets init user-service' to generate a template.
```

**When all secrets present:**

```
$ grund up user-service

Checking secrets...
  ✓ 3 secrets loaded

Starting infrastructure...
```

## Container Injection

Resolved secrets added to `docker-compose.generated.yaml`:

```yaml
services:
  user-service:
    environment:
      # From env:
      APP_ENV: development
      # From env_refs:
      DATABASE_URL: postgres://...
      # From secrets:
      OPENAI_API_KEY: sk-xxx
      STRIPE_SECRET_KEY: sk_test_xxx
```

## Implementation Files

| File | Change |
|:-----|:-------|
| `internal/config/schema.go` | Add `Secrets map[string]SecretConfig` to `ServiceConfig` |
| `internal/domain/infrastructure/` | Add `SecretRequirement` domain type |
| `internal/infrastructure/generator/compose_generator.go` | Inject resolved secrets into environment |
| `internal/infrastructure/generator/secrets_loader.go` | New: load and resolve secrets |
| `internal/cli/secrets.go` | New: `secrets list` and `secrets init` commands |
| `internal/cli/up.go` | Add secrets validation before starting |
| `internal/cli/root.go` | Register secrets command |

## Edge Cases

- `~/.grund/secrets.env` doesn't exist → Fall back to shell only
- Secret defined in both places → `secrets.env` wins
- No secrets declared → Skip validation entirely
- `grund secrets init` with existing file → Append missing, preserve existing
