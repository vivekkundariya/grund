# CLI Command Restructure Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Restructure Grund CLI commands into a consistent, extensible hierarchy.

**Architecture:** Introduce resource-based subcommands (`service`, `config`) while keeping daily-use commands at root level. Clean break - no backward compatibility needed.

**Tech Stack:** Go, Cobra CLI framework

---

## New Command Structure

```
grund
├── up <services>              # Daily operations (root level)
├── down
├── status
├── logs [services]
├── restart <service>
├── reset
│
├── service
│   ├── init                   # Create grund.yaml
│   ├── validate               # Validate config
│   ├── list                   # List registered services
│   ├── graph                  # Show dependency graph
│   └── add
│       ├── postgres <db>      # Add PostgreSQL
│       ├── mongodb <db>       # Add MongoDB
│       ├── redis              # Add Redis
│       ├── queue <name>       # Add SQS queue
│       ├── topic <name>       # Add SNS topic
│       ├── bucket <name>      # Add S3 bucket
│       ├── tunnel <name>      # Add tunnel
│       └── dependency <svc>   # Add service dependency
│
├── config
│   ├── init                   # Global setup (~/.grund/)
│   ├── show [service]         # Show configuration
│   └── set <key> <value>      # Set config value (future)
│
├── secrets
│   ├── list <service>
│   └── init <service>
│
└── doctor                     # Future: check prerequisites
```

---

## Task Breakdown

### Task 1: Create Service Parent Command

**Files:**
- Create: `internal/cli/service/service.go`

**Implementation:**

```go
// internal/cli/service/service.go
package service

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
    Use:   "service",
    Short: "Manage service configuration",
    Long: `Commands for managing service configuration files (grund.yaml).

Examples:
  grund service init              Initialize a new service
  grund service add postgres mydb Add PostgreSQL database
  grund service add queue orders  Add SQS queue
  grund service validate          Validate configuration`,
}
```

**Test:** `go build ./...`

**Commit:** `feat(cli): Add service parent command`

---

### Task 2: Move init → service init

**Files:**
- Move: `internal/cli/init.go` → `internal/cli/service/init.go`
- Modify: `internal/cli/root.go` (remove old init, wire service cmd)

**Changes:**
1. Move `initCmd` and `runInit` to service package
2. Update package name and imports
3. Register under `service.Cmd`
4. Remove from root.go

**Test:** `go run main.go service init --help`

**Commit:** `feat(cli): Move init to service init`

---

### Task 3: Create Service Add Parent Command

**Files:**
- Create: `internal/cli/service/add/add.go`

**Implementation:**

```go
// internal/cli/service/add/add.go
package add

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
    Use:   "add",
    Short: "Add infrastructure or dependencies",
    Long: `Add infrastructure requirements or service dependencies to grund.yaml.

Infrastructure:
  grund service add postgres <database>
  grund service add queue <name> [--dlq]
  grund service add bucket <name>
  grund service add tunnel <name> --port <port>

Dependencies:
  grund service add dependency <service>`,
}

func init() {
    Cmd.AddCommand(postgresCmd)
    Cmd.AddCommand(mongodbCmd)
    Cmd.AddCommand(redisCmd)
    Cmd.AddCommand(queueCmd)
    Cmd.AddCommand(topicCmd)
    Cmd.AddCommand(bucketCmd)
    Cmd.AddCommand(tunnelCmd)
    Cmd.AddCommand(dependencyCmd)
}
```

**Test:** `go run main.go service add --help`

**Commit:** `feat(cli): Add service add parent command`

---

### Task 4: Create Individual Add Subcommands

**Files:**
- Create: `internal/cli/service/add/postgres.go`
- Create: `internal/cli/service/add/mongodb.go`
- Create: `internal/cli/service/add/redis.go`
- Create: `internal/cli/service/add/queue.go`
- Create: `internal/cli/service/add/topic.go`
- Create: `internal/cli/service/add/bucket.go`
- Create: `internal/cli/service/add/tunnel.go`
- Create: `internal/cli/service/add/dependency.go`
- Create: `internal/cli/service/add/helpers.go` (shared logic from old add.go)

**Example - postgres.go:**

```go
package add

import "github.com/spf13/cobra"

var (
    postgresMigrations string
    postgresSeed       string
)

var postgresCmd = &cobra.Command{
    Use:   "postgres <database>",
    Short: "Add PostgreSQL database",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        return addPostgres(args[0], postgresMigrations, postgresSeed)
    },
}

func init() {
    postgresCmd.Flags().StringVar(&postgresMigrations, "migrations", "", "Migrations path")
    postgresCmd.Flags().StringVar(&postgresSeed, "seed", "", "Seed SQL path")
}
```

**Example - queue.go:**

```go
package add

import "github.com/spf13/cobra"

var queueDLQ bool

var queueCmd = &cobra.Command{
    Use:   "queue <name>",
    Short: "Add SQS queue",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        return addQueue(args[0], queueDLQ)
    },
}

func init() {
    queueCmd.Flags().BoolVar(&queueDLQ, "dlq", false, "Create dead-letter queue")
}
```

**Example - tunnel.go:**

```go
package add

import "github.com/spf13/cobra"

var (
    tunnelHost     string
    tunnelPort     string
    tunnelProvider string
)

var tunnelCmd = &cobra.Command{
    Use:   "tunnel <name>",
    Short: "Add tunnel for external access",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        return addTunnel(args[0], tunnelHost, tunnelPort, tunnelProvider)
    },
}

func init() {
    tunnelCmd.Flags().StringVar(&tunnelHost, "host", "localhost", "Local host")
    tunnelCmd.Flags().StringVar(&tunnelPort, "port", "", "Local port (required)")
    tunnelCmd.Flags().StringVar(&tunnelProvider, "provider", "cloudflared", "Provider")
    tunnelCmd.MarkFlagRequired("port")
}
```

**Test:**
```bash
go run main.go service add postgres testdb
go run main.go service add queue orders --dlq
go run main.go service add tunnel localstack --port 4566
```

**Commit:** `feat(cli): Add individual service add subcommands`

---

### Task 5: Delete Old add.go

**Files:**
- Delete: `internal/cli/add.go`

**Test:** `make test`

**Commit:** `refactor(cli): Remove old add command`

---

### Task 6: Create Config Parent Command

**Files:**
- Create: `internal/cli/configcmd/config.go`

**Implementation:**

```go
// internal/cli/configcmd/config.go
package configcmd

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
    Use:   "config",
    Short: "Manage Grund configuration",
    Long: `Commands for managing Grund global and service configuration.

Examples:
  grund config init           Initialize global config
  grund config show           Show global configuration
  grund config show myservice Show service configuration`,
}

func init() {
    Cmd.AddCommand(initCmd)
    Cmd.AddCommand(showCmd)
}
```

**Commit:** `feat(cli): Add config parent command`

---

### Task 7: Move setup → config init

**Files:**
- Move: `internal/cli/setup.go` → `internal/cli/configcmd/init.go`
- Modify: `internal/cli/root.go`

**Changes:**
1. Rename command from `setup` to `init`
2. Update package and imports
3. Remove setup from root.go

**Test:** `go run main.go config init --help`

**Commit:** `feat(cli): Move setup to config init`

---

### Task 8: Move config → config show

**Files:**
- Move: `internal/cli/config.go` → `internal/cli/configcmd/show.go`
- Modify: `internal/cli/root.go`

**Changes:**
1. Rename command from `config` to `show`
2. Update package and imports

**Test:**
```bash
go run main.go config show
go run main.go config show user-service
```

**Commit:** `feat(cli): Move config to config show`

---

### Task 9: Add service validate Command

**Files:**
- Create: `internal/cli/service/validate.go`

**Implementation:**

```go
package service

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/spf13/cobra"
    "gopkg.in/yaml.v3"
)

var validateCmd = &cobra.Command{
    Use:   "validate",
    Short: "Validate grund.yaml configuration",
    RunE:  runValidate,
}

func runValidate(cmd *cobra.Command, args []string) error {
    data, err := os.ReadFile(filepath.Join(".", "grund.yaml"))
    if err != nil {
        return fmt.Errorf("cannot read grund.yaml: %w", err)
    }

    var config map[string]any
    if err := yaml.Unmarshal(data, &config); err != nil {
        return fmt.Errorf("invalid YAML: %w", err)
    }

    // Validate required fields
    if err := validateRequiredFields(config); err != nil {
        return err
    }

    fmt.Println("✓ grund.yaml is valid")
    return nil
}

func validateRequiredFields(config map[string]any) error {
    svc, ok := config["service"].(map[string]any)
    if !ok {
        return fmt.Errorf("missing: service")
    }
    for _, field := range []string{"name", "type", "port"} {
        if _, ok := svc[field]; !ok {
            return fmt.Errorf("missing: service.%s", field)
        }
    }
    return nil
}
```

**Test:** `go run main.go service validate`

**Commit:** `feat(cli): Add service validate command`

---

### Task 10: Add service list Command

**Files:**
- Create: `internal/cli/service/list.go`

**Implementation:**

```go
package service

import (
    "fmt"

    "github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List all registered services",
    RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
    container, err := getContainer()
    if err != nil {
        return err
    }

    services, err := container.ServiceRepository.FindAll()
    if err != nil {
        return err
    }

    if len(services) == 0 {
        fmt.Println("No services registered")
        return nil
    }

    fmt.Printf("%-20s %-10s %-6s %s\n", "SERVICE", "TYPE", "PORT", "DEPS")
    for _, svc := range services {
        fmt.Printf("%-20s %-10s %-6d %d\n",
            svc.Name, svc.Type, svc.Port.Value(),
            len(svc.Dependencies.Services))
    }
    return nil
}
```

**Test:** `go run main.go service list`

**Commit:** `feat(cli): Add service list command`

---

### Task 11: Update root.go Wiring

**Files:**
- Modify: `internal/cli/root.go`

**Final structure:**

```go
func init() {
    // Global flags
    rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Config path")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

    // Daily operations (root level)
    rootCmd.AddCommand(upCmd)
    rootCmd.AddCommand(downCmd)
    rootCmd.AddCommand(statusCmd)
    rootCmd.AddCommand(logsCmd)
    rootCmd.AddCommand(restartCmd)
    rootCmd.AddCommand(resetCmd)

    // Service management
    rootCmd.AddCommand(service.Cmd)

    // Configuration
    rootCmd.AddCommand(configcmd.Cmd)

    // Secrets
    rootCmd.AddCommand(secretsCmd)
}
```

**Test:** `go run main.go --help`

**Commit:** `refactor(cli): Update root command wiring`

---

### Task 12: Update Documentation

**Files:**
- Modify: `docs/wiki/cli-commands.md`
- Modify: `README.md`

**Update cli-commands.md** with new command structure and examples.

**Update README.md** quick start section.

**Commit:** `docs: Update CLI documentation for new structure`

---

### Task 13: Run Full Test Suite

**Steps:**
```bash
make test
make test-e2e
```

**Manual E2E:**
```bash
# Global config
grund config init

# Service setup
cd /tmp && mkdir test-svc && cd test-svc
grund service init
grund service add postgres testdb
grund service add queue orders --dlq
grund service add bucket uploads
grund service add tunnel localstack --port 4566
grund service validate
grund service list

# Lifecycle
grund up test-svc
grund status
grund down
```

**Commit:** Fix any issues found

---

## Summary

| Task | Description |
|------|-------------|
| 1 | Create service parent command |
| 2 | Move init → service init |
| 3 | Create service add parent command |
| 4 | Create individual add subcommands |
| 5 | Delete old add.go |
| 6 | Create config parent command |
| 7 | Move setup → config init |
| 8 | Move config → config show |
| 9 | Add service validate |
| 10 | Add service list |
| 11 | Update root.go wiring |
| 12 | Update documentation |
| 13 | Run full test suite |

**Estimated files changed:** ~15 files
**New commands:** `service validate`, `service list`
**Moved commands:** `init`, `add`, `setup`, `config`
