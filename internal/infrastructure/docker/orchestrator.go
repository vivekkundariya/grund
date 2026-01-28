package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vivekkundariya/grund/internal/application/ports"
	"github.com/vivekkundariya/grund/internal/config"
	"github.com/vivekkundariya/grund/internal/domain/service"
	"github.com/vivekkundariya/grund/internal/ui"
)

// dockerComposeService represents the JSON output from docker compose ps
type dockerComposeService struct {
	Name       string `json:"Name"`
	State      string `json:"State"`
	Health     string `json:"Health"`
	Publishers []struct {
		URL           string `json:"URL"`
		TargetPort    int    `json:"TargetPort"`
		PublishedPort int    `json:"PublishedPort"`
		Protocol      string `json:"Protocol"`
	} `json:"Publishers"`
}

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
	ui.Step("Stopping services...")

	args := []string{"compose", "-f", d.composeFile, "down"}
	ui.Debug("Running: docker %s", strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = d.workingDir
	output, err := cmd.CombinedOutput()

	// Show output regardless of error
	if len(output) > 0 {
		// Print each line with proper formatting
		for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
			if line != "" {
				ui.SubStep("%s", line)
			}
		}
	}

	if err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	ui.Successf("All services stopped")
	return nil
}

// RestartService restarts a specific service
func (d *DockerOrchestrator) RestartService(ctx context.Context, name service.ServiceName) error {
	ui.Step("Restarting %s...", name.String())

	args := []string{"compose", "-f", d.composeFile, "restart", name.String()}
	ui.Debug("Running: docker %s", strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = d.workingDir
	output, err := cmd.CombinedOutput()

	// Show output regardless of error
	if len(output) > 0 {
		for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
			if line != "" {
				ui.SubStep("%s", line)
			}
		}
	}

	if err != nil {
		return fmt.Errorf("failed to restart service %s: %w", name, err)
	}

	ui.Successf("Service %s restarted", name.String())
	return nil
}

// GetServiceStatus gets the status of a service
func (d *DockerOrchestrator) GetServiceStatus(ctx context.Context, name service.ServiceName) (ports.ServiceStatus, error) {
	// Use docker compose ps to get status
	args := []string{"compose", "-f", d.composeFile, "ps", "--format", "json", name.String()}
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = d.workingDir
	output, err := cmd.Output()
	if err != nil {
		return ports.ServiceStatus{
			Name:   name.String(),
			Status: "not running",
			Health: "unknown",
		}, nil
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return ports.ServiceStatus{
			Name:   name.String(),
			Status: "not running",
			Health: "unknown",
		}, nil
	}

	// Parse JSON output
	var svc dockerComposeService
	if err := json.Unmarshal([]byte(outputStr), &svc); err != nil {
		// Fallback to simple parsing if JSON fails
		return d.parseStatusFallback(name.String(), outputStr), nil
	}

	status := ports.ServiceStatus{
		Name:   name.String(),
		Status: strings.ToLower(svc.State),
		Health: svc.Health,
	}

	if status.Health == "" {
		status.Health = "-"
	}

	// Build localhost URL from publishers
	if len(svc.Publishers) > 0 {
		for _, pub := range svc.Publishers {
			if pub.PublishedPort > 0 {
				status.Endpoint = fmt.Sprintf("http://localhost:%d", pub.PublishedPort)
				break // Use first published port
			}
		}
	}

	return status, nil
}

// parseStatusFallback handles cases where JSON parsing fails
func (d *DockerOrchestrator) parseStatusFallback(name, outputStr string) ports.ServiceStatus {
	status := ports.ServiceStatus{
		Name:   name,
		Status: "unknown",
		Health: "unknown",
	}

	if strings.Contains(outputStr, `"State":"running"`) || strings.Contains(outputStr, `"State": "running"`) {
		status.Status = "running"
	} else if strings.Contains(outputStr, `"State":"exited"`) || strings.Contains(outputStr, `"State": "exited"`) {
		status.Status = "exited"
	}

	if strings.Contains(outputStr, `"Health":"healthy"`) || strings.Contains(outputStr, `"Health": "healthy"`) {
		status.Health = "healthy"
	} else if strings.Contains(outputStr, `"Health":"unhealthy"`) || strings.Contains(outputStr, `"Health": "unhealthy"`) {
		status.Health = "unhealthy"
	} else if strings.Contains(outputStr, `"Health":""`) || strings.Contains(outputStr, `"Health": ""`) {
		status.Health = "-"
	}

	return status
}

// GetAllServiceStatuses gets status of all services in the compose file
func (d *DockerOrchestrator) GetAllServiceStatuses(ctx context.Context) ([]ports.ServiceStatus, error) {
	// Get all services from compose file
	configCmd := exec.CommandContext(ctx, "docker", "compose", "-f", d.composeFile, "config", "--services")
	configCmd.Dir = d.workingDir
	configOutput, err := configCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to read compose config: %w", err)
	}

	var statuses []ports.ServiceStatus
	for _, line := range strings.Split(string(configOutput), "\n") {
		svcName := strings.TrimSpace(line)
		if svcName == "" {
			continue
		}
		status, _ := d.GetServiceStatus(ctx, service.ServiceName(svcName))
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// GetLogs gets logs for a service
func (d *DockerOrchestrator) GetLogs(ctx context.Context, name service.ServiceName, follow bool, tail int) (ports.LogStream, error) {
	// TODO: Implement log streaming
	return nil, fmt.Errorf("not implemented")
}

// GetComposeFilePath returns the path to the generated compose file
// Files are stored in ~/.grund/tmp/<project-name>/ for easy cleanup
// Project name is derived from the orchestration root directory name
func GetComposeFilePath(orchestrationRoot string) string {
	grundHome, err := config.GetGrundHome()
	if err != nil {
		// Fallback to orchestration root if we can't get grund home
		return filepath.Join(orchestrationRoot, "docker-compose.generated.yaml")
	}

	// Get project name from orchestration root directory
	absRoot, err := filepath.Abs(orchestrationRoot)
	if err != nil {
		return filepath.Join(orchestrationRoot, "docker-compose.generated.yaml")
	}
	projectName := filepath.Base(absRoot)

	// Create project-specific tmp directory
	projectTmpDir := filepath.Join(grundHome, "tmp", projectName)

	// Ensure project tmp directory exists
	if err := os.MkdirAll(projectTmpDir, 0755); err != nil {
		// Fallback to orchestration root if we can't create tmp dir
		return filepath.Join(orchestrationRoot, "docker-compose.generated.yaml")
	}

	return filepath.Join(projectTmpDir, "docker-compose.generated.yaml")
}
