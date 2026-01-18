package ports

import (
	"github.com/vivekkundariya/grund/internal/domain/infrastructure"
	"github.com/vivekkundariya/grund/internal/domain/service"
)

// ComposeGenerator defines the interface for Docker Compose generation
type ComposeGenerator interface {
	Generate(services []*service.Service, infra infrastructure.InfrastructureRequirements) (string, error)
}

// EnvironmentResolver defines the interface for environment variable resolution
type EnvironmentResolver interface {
	Resolve(envRefs map[string]string, context EnvironmentContext) (map[string]string, error)
}

// EnvironmentContext provides context for resolving environment variables
type EnvironmentContext struct {
	Infrastructure map[string]InfrastructureContext
	Services       map[string]ServiceContext
	Self           ServiceContext
}

// InfrastructureContext provides infrastructure connection details
type InfrastructureContext struct {
	Host string
	Port int
}

// ServiceContext provides service connection details
type ServiceContext struct {
	Host   string
	Port   int
	Config map[string]interface{}
}
