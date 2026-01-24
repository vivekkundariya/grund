package docker

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vivekkundariya/grund/internal/application/ports"
	"github.com/vivekkundariya/grund/internal/domain/service"
	"github.com/vivekkundariya/grund/internal/ui"
)

// DockerOrchestrator implements ContainerOrchestrator using Docker Compose
// This is an infrastructure adapter
type DockerOrchestrator struct {
	composeFile string
	workingDir  string
}

// NewDockerOrchestrator creates a new Docker orchestrator
func NewDockerOrchestrator(composeFile, workingDir string) ports.ContainerOrchestrator {
	return &DockerOrchestrator{
		composeFile: composeFile,
		workingDir:  workingDir,
	}
}

// infrastructureServices are the services that should be started first
var infrastructureServices = []string{"postgres", "mongodb", "redis", "localstack"}

// StartInfrastructure starts infrastructure containers and waits for them to be healthy
func (d *DockerOrchestrator) StartInfrastructure(ctx context.Context) error {
	ui.Debug("Reading compose file: %s", d.composeFile)

	// First, get the list of services defined in the compose file
	configCmd := exec.CommandContext(ctx, "docker", "compose", "-f", d.composeFile, "config", "--services")
	configCmd.Dir = d.workingDir
	configOutput, err := configCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to read compose config: %w", err)
	}

	// Parse available services
	availableServices := make(map[string]bool)
	for _, line := range strings.Split(string(configOutput), "\n") {
		svc := strings.TrimSpace(line)
		if svc != "" {
			availableServices[svc] = true
		}
	}
	ui.Debug("Available services in compose: %v", availableServices)

	// Filter infrastructure services to only those in compose file
	var servicesToStart []string
	for _, svc := range infrastructureServices {
		if availableServices[svc] {
			servicesToStart = append(servicesToStart, svc)
		}
	}

	if len(servicesToStart) == 0 {
		ui.Debug("No infrastructure services to start")
		return nil
	}

	ui.SubStep("Starting: %s", strings.Join(servicesToStart, ", "))

	// Start infrastructure services with --wait to ensure they're healthy
	args := []string{"compose", "-f", d.composeFile, "up", "-d", "--wait"}
	args = append(args, servicesToStart...)

	ui.Debug("Running: docker %s", strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = d.workingDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		ui.Errorf("Docker compose output:\n%s", string(output))
		return fmt.Errorf("failed to start infrastructure: %w", err)
	}

	ui.Debug("Infrastructure started successfully")
	return nil
}

// StartServices starts services using docker-compose
func (d *DockerOrchestrator) StartServices(ctx context.Context, services []service.ServiceName) error {
	serviceNames := make([]string, len(services))
	for i, svc := range services {
		serviceNames[i] = svc.String()
	}

	ui.SubStep("Building and starting: %s", strings.Join(serviceNames, ", "))

	args := []string{"compose", "-f", d.composeFile, "up", "-d", "--build"}
	args = append(args, serviceNames...)

	ui.Debug("Running: docker %s", strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = d.workingDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		ui.Errorf("Docker compose output:\n%s", string(output))
		return fmt.Errorf("failed to start services: %w", err)
	}

	return nil
}

// StopServices stops all services
func (d *DockerOrchestrator) StopServices(ctx context.Context) error {
	args := []string{"compose", "-f", d.composeFile, "down"}

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = d.workingDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop services: %w\n%s", err, output)
	}

	return nil
}

// RestartService restarts a specific service
func (d *DockerOrchestrator) RestartService(ctx context.Context, name service.ServiceName) error {
	args := []string{"compose", "-f", d.composeFile, "restart", name.String()}

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = d.workingDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart service %s: %w\n%s", name, err, output)
	}

	return nil
}

// GetServiceStatus gets the status of a service
func (d *DockerOrchestrator) GetServiceStatus(ctx context.Context, name service.ServiceName) (ports.ServiceStatus, error) {
	// TODO: Implement actual status checking using docker ps
	return ports.ServiceStatus{
		Name:     name.String(),
		Status:   "unknown",
		Endpoint: "",
		Health:   "unknown",
	}, nil
}

// GetLogs gets logs for a service
func (d *DockerOrchestrator) GetLogs(ctx context.Context, name service.ServiceName, follow bool, tail int) (ports.LogStream, error) {
	// TODO: Implement log streaming
	return nil, fmt.Errorf("not implemented")
}

// GetComposeFilePath returns the path to the generated compose file
func GetComposeFilePath(orchestrationRoot string) string {
	return filepath.Join(orchestrationRoot, "docker-compose.generated.yaml")
}
