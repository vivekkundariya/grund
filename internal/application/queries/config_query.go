package queries

import (
	"github.com/yourorg/grund/internal/application/ports"
	"github.com/yourorg/grund/internal/domain/service"
)

// ConfigQuery represents a query for service configuration
type ConfigQuery struct {
	ServiceName string
}

// ConfigQueryHandler handles config queries
type ConfigQueryHandler struct {
	serviceRepo      ports.ServiceRepository
	registryRepo     ports.ServiceRegistryRepository
	envResolver      ports.EnvironmentResolver
}

// NewConfigQueryHandler creates a new config query handler
func NewConfigQueryHandler(
	serviceRepo ports.ServiceRepository,
	registryRepo ports.ServiceRegistryRepository,
	envResolver ports.EnvironmentResolver,
) *ConfigQueryHandler {
	return &ConfigQueryHandler{
		serviceRepo:  serviceRepo,
		registryRepo: registryRepo,
		envResolver:  envResolver,
	}
}

// ResolvedConfig represents the resolved configuration
type ResolvedConfig struct {
	Service         *service.Service
	Environment     map[string]string
	Infrastructure  []string
	Dependencies    []string
}

// Handle executes the config query
func (h *ConfigQueryHandler) Handle(query ConfigQuery) (*ResolvedConfig, error) {
	svc, err := h.serviceRepo.FindByName(service.ServiceName(query.ServiceName))
	if err != nil {
		return nil, err
	}

	// Resolve environment variables
	envContext := ports.EnvironmentContext{
		Infrastructure: make(map[string]ports.InfrastructureContext),
		Services:       make(map[string]ports.ServiceContext),
		Self: ports.ServiceContext{
			Host: "localhost",
			Port: svc.Port.Value(),
		},
	}

	resolvedEnv, err := h.envResolver.Resolve(svc.Environment.References, envContext)
	if err != nil {
		return nil, err
	}

	// Merge with direct env vars
	env := make(map[string]string)
	for k, v := range svc.Environment.Variables {
		env[k] = v
	}
	for k, v := range resolvedEnv {
		env[k] = v
	}

	// Collect infrastructure types
	var infraTypes []string
	if svc.RequiresInfrastructure("postgres") {
		infraTypes = append(infraTypes, "postgres")
	}
	if svc.RequiresInfrastructure("mongodb") {
		infraTypes = append(infraTypes, "mongodb")
	}
	if svc.RequiresInfrastructure("redis") {
		infraTypes = append(infraTypes, "redis")
	}
	if svc.RequiresInfrastructure("localstack") {
		infraTypes = append(infraTypes, "localstack")
	}

	// Collect service dependencies
	var deps []string
	for _, dep := range svc.Dependencies.Services {
		deps = append(deps, dep.String())
	}

	return &ResolvedConfig{
		Service:        svc,
		Environment:    env,
		Infrastructure: infraTypes,
		Dependencies:   deps,
	}, nil
}
