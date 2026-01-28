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
// 3. Local file in current directory (services.yaml, grund-services.yaml)
// 4. Parent directory traversal
// 5. Global config: ~/.grund/services.yaml
type ConfigResolver struct {
	// CLIConfigPath is set via --config flag
	CLIConfigPath string

	// GlobalConfig is the loaded global configuration
	GlobalConfig *GlobalConfig
}

// serviceFileNames are the filenames to look for in order of preference
var serviceFileNames = []string{
	"services.yaml",
	"grund-services.yaml",
	"services.yml",
	"grund-services.yml",
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

	// 3. Search current directory and parent directories
	if path, root, found := r.searchLocalFiles(); found {
		return path, root, nil
	}

	// 4. Global config: ~/.grund/services.yaml (or GRUND_HOME)
	grundDir, err := GetGrundHome()
	if err != nil {
		return "", "", fmt.Errorf("failed to get grund home directory: %w", err)
	}

	defaultPath := filepath.Join(grundDir, "services.yaml")

	if _, err := os.Stat(defaultPath); err != nil {
		return "", "", fmt.Errorf("services file not found\n\nSearched:\n  - Current directory and parents for: %v\n  - Global config at: %s\n\nTo get started:\n  1. Create services.yaml in your project root\n  2. Or create ~/.grund/services.yaml\n  3. Or set GRUND_CONFIG environment variable\n  4. Or use --config flag", serviceFileNames, defaultPath)
	}

	return defaultPath, grundDir, nil
}

// searchLocalFiles searches for service files in current and parent directories
func (r *ConfigResolver) searchLocalFiles() (path string, root string, found bool) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", false
	}

	dir := cwd
	for {
		// Try each filename in current directory
		for _, name := range serviceFileNames {
			candidate := filepath.Join(dir, name)
			if _, err := os.Stat(candidate); err == nil {
				return candidate, dir, true
			}
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root, stop searching
			break
		}
		dir = parent
	}

	return "", "", false
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
