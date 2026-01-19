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
// 3. Local file (./services.yaml or ./grund-services.yaml)
// 4. Global config default path
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
	
	// 3. Search for local file in current and parent directories
	localPath, localRoot, found := r.findLocalServicesFile()
	if found {
		return localPath, localRoot, nil
	}
	
	// 4. Check global config default orchestration repo
	if r.GlobalConfig.DefaultOrchestrationRepo != "" {
		expandedPath := expandPath(r.GlobalConfig.DefaultOrchestrationRepo)
		servicesPath := filepath.Join(expandedPath, r.GlobalConfig.DefaultServicesFile)
		
		if _, err := os.Stat(servicesPath); err == nil {
			return servicesPath, expandedPath, nil
		}
	}
	
	return "", "", fmt.Errorf("services file not found. Use --config flag, set GRUND_CONFIG env var, or run from orchestration repo")
}

// findLocalServicesFile searches for services file in current and parent directories
func (r *ConfigResolver) findLocalServicesFile() (string, string, bool) {
	wd, err := os.Getwd()
	if err != nil {
		return "", "", false
	}
	
	// Files to search for (in priority order)
	searchFiles := []string{
		r.GlobalConfig.DefaultServicesFile, // services.yaml (configurable)
		"grund-services.yaml",              // Alternative name
		"grund.services.yaml",              // Alternative name
	}
	
	// Search up to 5 parent directories
	path := wd
	for i := 0; i < 5; i++ {
		for _, fileName := range searchFiles {
			servicesPath := filepath.Join(path, fileName)
			if _, err := os.Stat(servicesPath); err == nil {
				return servicesPath, path, true
			}
		}
		
		parent := filepath.Dir(path)
		if parent == path || parent == "/" || parent == "." {
			break
		}
		path = parent
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
