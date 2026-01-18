package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetServicePath returns the local path for a service name
func GetServicePath(registry *ServiceRegistry, serviceName string) (string, error) {
	entry, ok := registry.Services[serviceName]
	if !ok {
		return "", fmt.Errorf("service %s not found in registry", serviceName)
	}

	// Expand ~ to home directory
	path := entry.Path
	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	return path, nil
}

// GetServiceConfigPath returns the path to grund.yaml for a service
func GetServiceConfigPath(registry *ServiceRegistry, serviceName string) (string, error) {
	servicePath, err := GetServicePath(registry, serviceName)
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(servicePath, "grund.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("grund.yaml not found at %s", configPath)
	}

	return configPath, nil
}
