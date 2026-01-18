package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadServiceConfig loads a grund.yaml file from the given path
func LoadServiceConfig(path string) (*ServiceConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ServiceConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

// LoadServiceRegistry loads the services.yaml file from the orchestration repo
func LoadServiceRegistry(path string) (*ServiceRegistry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read services.yaml: %w", err)
	}

	var registry ServiceRegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse services.yaml: %w", err)
	}

	return &registry, nil
}
