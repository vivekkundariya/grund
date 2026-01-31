package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// GlobalConfigDir is the directory for global Grund configuration
	GlobalConfigDir = ".grund"

	// GlobalConfigFile is the global configuration file name
	GlobalConfigFile = "config.yaml"

	// EnvConfigFile is the environment variable for custom config file path
	EnvConfigFile = "GRUND_CONFIG"

	// EnvGrundHome is the environment variable for Grund home directory
	EnvGrundHome = "GRUND_HOME"

	// Default LocalStack settings
	DefaultLocalStackEndpoint = "http://localhost:4566"
	DefaultLocalStackRegion   = "us-east-1"

	// Default Docker settings
	DefaultDockerComposeCommand = "docker compose"
)

// GlobalConfig represents the unified Grund configuration
// stored at ~/.grund/config.yaml
type GlobalConfig struct {
	// Version of the config schema
	Version string `yaml:"version"`

	// Services is the registry of services
	Services map[string]ServiceEntry `yaml:"services,omitempty"`

	// Docker configuration (optional, defaults applied if not set)
	Docker *DockerConfig `yaml:"docker,omitempty"`

	// LocalStack configuration (optional, defaults applied if not set)
	LocalStack *LocalStackConfig `yaml:"localstack,omitempty"`
}

// ServiceEntry represents a service in the registry
type ServiceEntry struct {
	// Path to the service directory
	Path string `yaml:"path"`

	// Repo is the git repository URL (optional)
	Repo string `yaml:"repo,omitempty"`
}

// DockerConfig holds Docker-related settings
type DockerConfig struct {
	// ComposeCommand is the command to run docker compose (default: "docker compose")
	ComposeCommand string `yaml:"compose_command,omitempty"`
}

// LocalStackConfig holds LocalStack-related settings
type LocalStackConfig struct {
	// Endpoint is the LocalStack endpoint URL (default: "http://localhost:4566")
	Endpoint string `yaml:"endpoint,omitempty"`

	// Region is the AWS region for LocalStack (default: "us-east-1")
	Region string `yaml:"region,omitempty"`
}

// GetGrundHome returns the Grund home directory
// Priority: GRUND_HOME env var > ~/.grund
func GetGrundHome() (string, error) {
	if home := os.Getenv(EnvGrundHome); home != "" {
		return home, nil
	}

	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(userHome, GlobalConfigDir), nil
}

// GetGlobalConfigPath returns the path to the global config file
func GetGlobalConfigPath() (string, error) {
	grundHome, err := GetGrundHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(grundHome, GlobalConfigFile), nil
}

// LoadGlobalConfig loads the global configuration
// Returns default config if file doesn't exist
func LoadGlobalConfig() (*GlobalConfig, error) {
	configPath, err := GetGlobalConfigPath()
	if err != nil {
		return DefaultGlobalConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultGlobalConfig(), nil
		}
		return nil, fmt.Errorf("failed to read global config: %w", err)
	}

	var config GlobalConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse global config: %w", err)
	}

	// Ensure services map is initialized
	if config.Services == nil {
		config.Services = make(map[string]ServiceEntry)
	}

	return &config, nil
}

// SaveGlobalConfig saves the global configuration
func SaveGlobalConfig(config *GlobalConfig) error {
	grundHome, err := GetGrundHome()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(grundHome, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(grundHome, GlobalConfigFile)

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// DefaultGlobalConfig returns the default global configuration
func DefaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Version:  "1",
		Services: make(map[string]ServiceEntry),
	}
}

// InitGlobalConfig initializes the global config directory and file
func InitGlobalConfig() error {
	configPath, err := GetGlobalConfigPath()
	if err != nil {
		return err
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // Already exists
	}

	// Create default config
	return SaveGlobalConfig(DefaultGlobalConfig())
}

// GlobalConfigExists checks if global config file exists
func GlobalConfigExists() bool {
	configPath, err := GetGlobalConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(configPath)
	return err == nil
}

// ForceInitGlobalConfig initializes the global config, overwriting if exists
func ForceInitGlobalConfig() error {
	return SaveGlobalConfig(DefaultGlobalConfig())
}

// GetLocalStackEndpoint returns the LocalStack endpoint
// Reads from config if set, otherwise returns default
func GetLocalStackEndpoint() string {
	if cfg, err := LoadGlobalConfig(); err == nil && cfg.LocalStack != nil && cfg.LocalStack.Endpoint != "" {
		return cfg.LocalStack.Endpoint
	}
	return DefaultLocalStackEndpoint
}

// GetLocalStackRegion returns the LocalStack region
// Reads from config if set, otherwise returns default
func GetLocalStackRegion() string {
	if cfg, err := LoadGlobalConfig(); err == nil && cfg.LocalStack != nil && cfg.LocalStack.Region != "" {
		return cfg.LocalStack.Region
	}
	return DefaultLocalStackRegion
}

// GetDockerComposeCommand returns the docker compose command
// Reads from config if set, otherwise returns default
func GetDockerComposeCommand() string {
	if cfg, err := LoadGlobalConfig(); err == nil && cfg.Docker != nil && cfg.Docker.ComposeCommand != "" {
		return cfg.Docker.ComposeCommand
	}
	return DefaultDockerComposeCommand
}

// GetLocalStackEndpointFromConfig returns the LocalStack endpoint from a config instance
func (c *GlobalConfig) GetLocalStackEndpoint() string {
	if c.LocalStack != nil && c.LocalStack.Endpoint != "" {
		return c.LocalStack.Endpoint
	}
	return DefaultLocalStackEndpoint
}

// GetLocalStackRegionFromConfig returns the LocalStack region from a config instance
func (c *GlobalConfig) GetLocalStackRegion() string {
	if c.LocalStack != nil && c.LocalStack.Region != "" {
		return c.LocalStack.Region
	}
	return DefaultLocalStackRegion
}

// GetDockerComposeCommandFromConfig returns the docker compose command from a config instance
func (c *GlobalConfig) GetDockerComposeCommand() string {
	if c.Docker != nil && c.Docker.ComposeCommand != "" {
		return c.Docker.ComposeCommand
	}
	return DefaultDockerComposeCommand
}

// AddService adds a service to the config
func (c *GlobalConfig) AddService(name string, entry ServiceEntry) {
	if c.Services == nil {
		c.Services = make(map[string]ServiceEntry)
	}
	c.Services[name] = entry
}

// GetService returns a service entry by name
func (c *GlobalConfig) GetService(name string) (ServiceEntry, bool) {
	entry, ok := c.Services[name]
	return entry, ok
}
