package docker

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/yourorg/grund/internal/application/ports"
	"github.com/yourorg/grund/internal/domain/service"
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

// StartServices starts services using docker-compose
func (d *DockerOrchestrator) StartServices(ctx context.Context, services []service.ServiceName) error {
	args := []string{"compose", "-f", d.composeFile, "up", "-d"}
	for _, svc := range services {
		args = append(args, svc.String())
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = d.workingDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start services: %w\n%s", err, output)
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
