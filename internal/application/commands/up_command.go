package commands

import (
	"context"
	"fmt"

	"github.com/vivekkundariya/grund/internal/application/ports"
	"github.com/vivekkundariya/grund/internal/domain/dependency"
	"github.com/vivekkundariya/grund/internal/domain/infrastructure"
	"github.com/vivekkundariya/grund/internal/domain/service"
)

// UpCommand represents the command to start services
type UpCommand struct {
	ServiceNames []string
	NoDeps       bool
	InfraOnly    bool
	Build        bool
	Local        bool
}

// UpCommandHandler handles the up command
// This follows the Command pattern and Single Responsibility Principle
type UpCommandHandler struct {
	serviceRepo      ports.ServiceRepository
	registryRepo     ports.ServiceRegistryRepository
	orchestrator     ports.ContainerOrchestrator
	provisioner      ports.InfrastructureProvisioner
	composeGenerator ports.ComposeGenerator
	healthChecker    ports.HealthChecker
}

// NewUpCommandHandler creates a new up command handler
func NewUpCommandHandler(
	serviceRepo ports.ServiceRepository,
	registryRepo ports.ServiceRegistryRepository,
	orchestrator ports.ContainerOrchestrator,
	provisioner ports.InfrastructureProvisioner,
	composeGenerator ports.ComposeGenerator,
	healthChecker ports.HealthChecker,
) *UpCommandHandler {
	return &UpCommandHandler{
		serviceRepo:      serviceRepo,
		registryRepo:      registryRepo,
		orchestrator:      orchestrator,
		provisioner:       provisioner,
		composeGenerator:  composeGenerator,
		healthChecker:     healthChecker,
	}
}

// Handle executes the up command
func (h *UpCommandHandler) Handle(ctx context.Context, cmd UpCommand) error {
	// 1. Load services
	services, err := h.loadServices(cmd.ServiceNames)
	if err != nil {
		return fmt.Errorf("failed to load services: %w", err)
	}

	// 2. Build dependency graph
	graph := dependency.NewGraph()
	for _, svc := range services {
		graph.AddService(svc)
	}
	if err := graph.Build(); err != nil {
		return fmt.Errorf("failed to build dependency graph: %w", err)
	}

	// 3. Check for circular dependencies
	for _, svcName := range cmd.ServiceNames {
		if cycle, err := graph.DetectCycle(service.ServiceName(svcName)); err != nil {
			return fmt.Errorf("circular dependency: %w", err)
		} else if len(cycle) > 0 {
			return fmt.Errorf("circular dependency detected")
		}
	}

	// 4. Resolve startup order
	serviceNames := make([]service.ServiceName, len(cmd.ServiceNames))
	for i, name := range cmd.ServiceNames {
		serviceNames[i] = service.ServiceName(name)
	}

	startupOrder, err := graph.TopologicalSort(serviceNames)
	if err != nil {
		return fmt.Errorf("failed to determine startup order: %w", err)
	}

	// 5. Aggregate infrastructure requirements
	infraReqs := h.aggregateInfrastructure(services)

	// 6. Generate docker-compose
	if err := h.generateCompose(services, infraReqs); err != nil {
		return fmt.Errorf("failed to generate compose file: %w", err)
	}

	// 7. Start infrastructure containers first (postgres, redis, localstack, etc.)
	if err := h.orchestrator.StartInfrastructure(ctx); err != nil {
		return fmt.Errorf("failed to start infrastructure: %w", err)
	}

	// 8. Provision infrastructure (create databases, SQS queues, etc.)
	// This runs AFTER containers are up
	if err := h.provisionInfrastructure(ctx, infraReqs); err != nil {
		return fmt.Errorf("failed to provision infrastructure: %w", err)
	}

	// 9. Start services in order
	if !cmd.InfraOnly {
		if err := h.startServices(ctx, startupOrder); err != nil {
			return fmt.Errorf("failed to start services: %w", err)
		}
	}

	return nil
}

func (h *UpCommandHandler) loadServices(names []string) ([]*service.Service, error) {
	var services []*service.Service
	for _, name := range names {
		svc, err := h.serviceRepo.FindByName(service.ServiceName(name))
		if err != nil {
			return nil, fmt.Errorf("service %s: %w", name, err)
		}
		services = append(services, svc)
	}
	return services, nil
}

func (h *UpCommandHandler) aggregateInfrastructure(services []*service.Service) infrastructure.InfrastructureRequirements {
	var reqs []infrastructure.InfrastructureRequirements
	for _, svc := range services {
		reqs = append(reqs, svc.Dependencies.Infrastructure)
	}
	return infrastructure.Aggregate(reqs...)
}

func (h *UpCommandHandler) provisionInfrastructure(ctx context.Context, req infrastructure.InfrastructureRequirements) error {
	if req.Postgres != nil {
		if err := h.provisioner.ProvisionPostgres(ctx, req.Postgres); err != nil {
			return fmt.Errorf("postgres: %w", err)
		}
	}
	if req.MongoDB != nil {
		if err := h.provisioner.ProvisionMongoDB(ctx, req.MongoDB); err != nil {
			return fmt.Errorf("mongodb: %w", err)
		}
	}
	if req.Redis != nil {
		if err := h.provisioner.ProvisionRedis(ctx, req.Redis); err != nil {
			return fmt.Errorf("redis: %w", err)
		}
	}
	if req.SQS != nil || req.SNS != nil || req.S3 != nil {
		if err := h.provisioner.ProvisionLocalStack(ctx, req); err != nil {
			return fmt.Errorf("localstack: %w", err)
		}
	}
	return nil
}

func (h *UpCommandHandler) generateCompose(services []*service.Service, req infrastructure.InfrastructureRequirements) error {
	_, err := h.composeGenerator.Generate(services, req)
	return err
}

func (h *UpCommandHandler) startServices(ctx context.Context, order []service.ServiceName) error {
	return h.orchestrator.StartServices(ctx, order)
}
