package config

import (
	"fmt"
)

// ValidateServiceConfig validates a service configuration
func ValidateServiceConfig(config *ServiceConfig) error {
	if config.Version == "" {
		return fmt.Errorf("version is required")
	}

	if config.Service.Name == "" {
		return fmt.Errorf("service.name is required")
	}

	if config.Service.Port <= 0 {
		return fmt.Errorf("service.port must be greater than 0")
	}

	if config.Service.Health.Endpoint == "" {
		return fmt.Errorf("service.health.endpoint is required")
	}

	// Validate that either build or run is specified
	if config.Service.Build == nil && config.Service.Run == nil {
		return fmt.Errorf("either service.build or service.run must be specified")
	}

	return nil
}

// ValidateServiceRegistry validates the service registry
func ValidateServiceRegistry(registry *ServiceRegistry) error {
	if registry.Version == "" {
		return fmt.Errorf("version is required")
	}

	if len(registry.Services) == 0 {
		return fmt.Errorf("at least one service must be registered")
	}

	for name, entry := range registry.Services {
		if entry.Repo == "" {
			return fmt.Errorf("service %s: repo is required", name)
		}
		if entry.Path == "" {
			return fmt.Errorf("service %s: path is required", name)
		}
	}

	return nil
}
