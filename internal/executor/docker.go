package executor

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

// DockerCompose wraps docker-compose commands
type DockerCompose struct {
	ComposeFile string
	WorkingDir  string
}

// NewDockerCompose creates a new DockerCompose executor
func NewDockerCompose(composeFile, workingDir string) *DockerCompose {
	return &DockerCompose{
		ComposeFile: composeFile,
		WorkingDir:  workingDir,
	}
}

// Up starts services using docker-compose
func (dc *DockerCompose) Up(services ...string) error {
	args := []string{"compose", "-f", dc.ComposeFile, "up", "-d"}
	args = append(args, services...)

	cmd := exec.Command("docker", args...)
	cmd.Dir = dc.WorkingDir
	cmd.Stdout = nil // TODO: Stream output
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	return nil
}

// Down stops all services
func (dc *DockerCompose) Down() error {
	args := []string{"compose", "-f", dc.ComposeFile, "down"}

	cmd := exec.Command("docker", args...)
	cmd.Dir = dc.WorkingDir

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	return nil
}

// Restart restarts specific services
func (dc *DockerCompose) Restart(services ...string) error {
	args := []string{"compose", "-f", dc.ComposeFile, "restart"}
	args = append(args, services...)

	cmd := exec.Command("docker", args...)
	cmd.Dir = dc.WorkingDir

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restart services: %w", err)
	}

	return nil
}

// GetComposeFilePath returns the full path to the compose file
func GetComposeFilePath(orchestrationRoot string) string {
	return filepath.Join(orchestrationRoot, "docker-compose.generated.yaml")
}
