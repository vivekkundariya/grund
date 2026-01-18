package wiring

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vivekkundariya/grund/internal/application/commands"
	"github.com/vivekkundariya/grund/internal/application/queries"
	"github.com/vivekkundariya/grund/internal/infrastructure/aws"
	"github.com/vivekkundariya/grund/internal/infrastructure/config"
	"github.com/vivekkundariya/grund/internal/infrastructure/docker"
	"github.com/vivekkundariya/grund/internal/infrastructure/generator"
)

// Container holds all dependencies (Dependency Injection Container)
// This follows the Dependency Inversion Principle
type Container struct {
	// Repositories
	ServiceRepo  interface{} // ports.ServiceRepository
	RegistryRepo interface{} // ports.ServiceRegistryRepository

	// Infrastructure
	Orchestrator     interface{} // ports.ContainerOrchestrator
	Provisioner      interface{} // ports.InfrastructureProvisioner
	ComposeGenerator interface{} // ports.ComposeGenerator
	EnvResolver      interface{} // ports.EnvironmentResolver
	HealthChecker    interface{} // ports.HealthChecker

	// Command Handlers
	UpCommandHandler      *commands.UpCommandHandler
	DownCommandHandler    *commands.DownCommandHandler
	RestartCommandHandler *commands.RestartCommandHandler

	// Query Handlers
	StatusQueryHandler *queries.StatusQueryHandler
	ConfigQueryHandler *queries.ConfigQueryHandler
}

// NewContainer creates a new dependency injection container
func NewContainer(orchestrationRoot string) (*Container, error) {
	// Find services.yaml
	servicesPath := filepath.Join(orchestrationRoot, "services.yaml")
	if _, err := os.Stat(servicesPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("services.yaml not found at %s", servicesPath)
	}

	// Initialize repositories
	registryRepo, err := config.NewServiceRegistryRepository(servicesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry repository: %w", err)
	}

	serviceRepo := config.NewServiceRepository(registryRepo)

	// Initialize infrastructure adapters
	composeFile := docker.GetComposeFilePath(orchestrationRoot)
	orchestrator := docker.NewDockerOrchestrator(composeFile, orchestrationRoot)
	healthChecker := docker.NewHTTPHealthChecker()

	// Initialize provisioners
	postgresProvisioner := docker.NewPostgresProvisioner()
	localstackProvisioner := aws.NewLocalStackProvisioner("http://localhost:4566")
	provisioner := docker.NewCompositeInfrastructureProvisioner(
		postgresProvisioner,
		localstackProvisioner,
	)

	// Initialize generators
	composeGenerator := generator.NewComposeGenerator(composeFile)
	envResolver := generator.NewEnvironmentResolver()

	// Initialize command handlers
	upHandler := commands.NewUpCommandHandler(
		serviceRepo,
		registryRepo,
		orchestrator,
		provisioner,
		composeGenerator,
		healthChecker,
	)

	downHandler := commands.NewDownCommandHandler(orchestrator)
	restartHandler := commands.NewRestartCommandHandler(orchestrator)

	// Initialize query handlers
	statusHandler := queries.NewStatusQueryHandler(orchestrator)
	configHandler := queries.NewConfigQueryHandler(
		serviceRepo,
		registryRepo,
		envResolver,
	)

	return &Container{
		ServiceRepo:           serviceRepo,
		RegistryRepo:          registryRepo,
		Orchestrator:          orchestrator,
		Provisioner:           provisioner,
		ComposeGenerator:      composeGenerator,
		EnvResolver:           envResolver,
		HealthChecker:         healthChecker,
		UpCommandHandler:      upHandler,
		DownCommandHandler:    downHandler,
		RestartCommandHandler: restartHandler,
		StatusQueryHandler:    statusHandler,
		ConfigQueryHandler:    configHandler,
	}, nil
}
