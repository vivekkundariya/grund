package executor

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

// StreamLogs streams logs from docker-compose
func StreamLogs(composeFile, workingDir string, services []string, follow bool, tail int) error {
	args := []string{"compose", "-f", composeFile, "logs"}

	if follow {
		args = append(args, "-f")
	}

	if tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tail))
	}

	args = append(args, services...)

	cmd := exec.Command("docker", args...)
	cmd.Dir = workingDir
	cmd.Stdout = nil // TODO: Stream to stdout
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stream logs: %w", err)
	}

	return nil
}

// GetLogsPath returns the path for log files
func GetLogsPath(orchestrationRoot string) string {
	return filepath.Join(orchestrationRoot, "logs")
}
