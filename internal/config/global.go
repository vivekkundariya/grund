package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultConfigFileName is the default name for the services registry file
	DefaultConfigFileName = "services.yaml"

	// GlobalConfigDir is the directory for global Grund configuration
	GlobalConfigDir = ".grund"

	// GlobalConfigFile is the global configuration file name
	GlobalConfigFile = "config.yaml"

	// EnvConfigFile is the environment variable for custom config file path
	EnvConfigFile = "GRUND_CONFIG"

	// EnvGrundHome is the environment variable for Grund home directory
	EnvGrundHome = "GRUND_HOME"
)

// GlobalConfig represents the global Grund configuration
// stored at ~/.grund/config.yaml
type GlobalConfig struct {
	// DefaultServicesFile is the default path to services.yaml
	// Can be absolute or relative to current directory
	DefaultServicesFile string `yaml:"default_services_file,omitempty"`

	// DefaultOrchestrationRepo is the default path to the orchestration repo
	DefaultOrchestrationRepo string `yaml:"default_orchestration_repo,omitempty"`

	// ServicesBasePath is the base path where services are cloned
	ServicesBasePath string `yaml:"services_base_path,omitempty"`

	// Docker configuration
	Docker DockerConfig `yaml:"docker,omitempty"`

	// LocalStack configuration
	LocalStack LocalStackConfig `yaml:"localstack,omitempty"`
}

// DockerConfig holds Docker-related configuration
type DockerConfig struct {
	// ComposeCommand is the docker compose command (default: "docker compose")
	ComposeCommand string `yaml:"compose_command,omitempty"`
}

// LocalStackConfig holds LocalStack-related configuration
type LocalStackConfig struct {
	// Endpoint is the LocalStack endpoint (default: "http://localhost:4566")
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

	// Apply defaults for unset values
	config.applyDefaults()

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
		DefaultServicesFile: DefaultConfigFileName,
		Docker: DockerConfig{
			ComposeCommand: "docker compose",
		},
		LocalStack: LocalStackConfig{
			Endpoint: "http://localhost:4566",
			Region:   "us-east-1",
		},
	}
}

// applyDefaults applies default values to unset fields
func (c *GlobalConfig) applyDefaults() {
	defaults := DefaultGlobalConfig()

	if c.DefaultServicesFile == "" {
		c.DefaultServicesFile = defaults.DefaultServicesFile
	}
	if c.Docker.ComposeCommand == "" {
		c.Docker.ComposeCommand = defaults.Docker.ComposeCommand
	}
	if c.LocalStack.Endpoint == "" {
		c.LocalStack.Endpoint = defaults.LocalStack.Endpoint
	}
	if c.LocalStack.Region == "" {
		c.LocalStack.Region = defaults.LocalStack.Region
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
