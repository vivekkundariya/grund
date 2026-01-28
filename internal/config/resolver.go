package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConfigResolver resolves configuration from multiple sources
// Priority order (highest to lowest):
// 1. CLI flag (--config)
// 2. Environment variable (GRUND_CONFIG)
// 3. Default: ~/.grund/services.yaml
type ConfigResolver struct {
	// CLIConfigPath is set via --config flag
	CLIConfigPath string

	// GlobalConfig is the loaded global configuration
	GlobalConfig *GlobalConfig
}

// NewConfigResolver creates a new config resolver
func NewConfigResolver(cliConfigPath string) (*ConfigResolver, error) {
	globalConfig, err := LoadGlobalConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}
	
	return &ConfigResolver{
		CLIConfigPath: cliConfigPath,
		GlobalConfig:  globalConfig,
	}, nil
}

// ResolveServicesFile resolves the path to the services registry file
// Returns the resolved path and the orchestration root directory
func (r *ConfigResolver) ResolveServicesFile() (servicesPath string, orchestrationRoot string, err error) {
	// 1. CLI flag has highest priority
	if r.CLIConfigPath != "" {
		absPath, err := filepath.Abs(r.CLIConfigPath)
		if err != nil {
			return "", "", fmt.Errorf("failed to resolve CLI config path: %w", err)
		}

		if _, err := os.Stat(absPath); err != nil {
			return "", "", fmt.Errorf("config file not found: %s", absPath)
		}

		return absPath, filepath.Dir(absPath), nil
	}

	// 2. Environment variable
	if envPath := os.Getenv(EnvConfigFile); envPath != "" {
		absPath, err := filepath.Abs(envPath)
		if err != nil {
			return "", "", fmt.Errorf("failed to resolve GRUND_CONFIG path: %w", err)
		}

		if _, err := os.Stat(absPath); err != nil {
			return "", "", fmt.Errorf("config file from GRUND_CONFIG not found: %s", absPath)
		}

		return absPath, filepath.Dir(absPath), nil
	}

	// 3. Default: ~/.grund/services.yaml
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to get home directory: %w", err)
	}

	grundDir := filepath.Join(home, ".grund")
	defaultPath := filepath.Join(grundDir, "services.yaml")

	if _, err := os.Stat(defaultPath); err != nil {
		return "", "", fmt.Errorf("services file not found at %s\n\nTo get started:\n  1. Create ~/.grund/services.yaml with your service registry\n  2. Or set GRUND_CONFIG environment variable\n  3. Or use --config flag", defaultPath)
	}

	return defaultPath, grundDir, nil
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if len(path) == 0 {
		return path
	}
	
	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		if len(path) > 1 {
			return filepath.Join(home, path[2:])
		}
		return home
	}
	
	return path
}

// GetServicesBasePath returns the base path for cloning services
func (r *ConfigResolver) GetServicesBasePath() string {
	if r.GlobalConfig.ServicesBasePath != "" {
		return expandPath(r.GlobalConfig.ServicesBasePath)
	}
	
	// Default to ~/projects
	home, err := os.UserHomeDir()
	if err != nil {
		return "./services"
	}
	return filepath.Join(home, "projects")
}

// GetDockerComposeCommand returns the docker compose command
func (r *ConfigResolver) GetDockerComposeCommand() string {
	return r.GlobalConfig.Docker.ComposeCommand
}

// GetLocalStackEndpoint returns the LocalStack endpoint
func (r *ConfigResolver) GetLocalStackEndpoint() string {
	return r.GlobalConfig.LocalStack.Endpoint
}

// GetLocalStackRegion returns the LocalStack AWS region
func (r *ConfigResolver) GetLocalStackRegion() string {
	return r.GlobalConfig.LocalStack.Region
}
