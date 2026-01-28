package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/vivekkundariya/grund/internal/infrastructure/docker"
)

var (
	logsFollow bool
	logsTail   int
)

var logsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "View aggregated or per-service logs",
	Long:  `View logs from all services or a specific service.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if configResolver == nil {
			return fmt.Errorf("config not initialized")
		}

		_, orchestrationRoot, err := configResolver.ResolveServicesFile()
		if err != nil {
			return err
		}

		composeFile := docker.GetComposeFilePath(orchestrationRoot)
		projectName := docker.GetProjectName(orchestrationRoot)

		// Build docker compose logs command with project name
		dockerArgs := []string{"compose", "-p", projectName, "-f", composeFile, "logs"}

		if logsFollow {
			dockerArgs = append(dockerArgs, "-f")
		}

		if logsTail > 0 {
			dockerArgs = append(dockerArgs, "--tail", strconv.Itoa(logsTail))
		}

		// Add service name if specified
		if len(args) > 0 {
			dockerArgs = append(dockerArgs, args[0])
		}

		// Execute docker compose logs with stdout/stderr connected
		dockerCmd := exec.CommandContext(cmd.Context(), "docker", dockerArgs...)
		dockerCmd.Dir = orchestrationRoot
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr
		dockerCmd.Stdin = os.Stdin

		return dockerCmd.Run()
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")
	logsCmd.Flags().IntVar(&logsTail, "tail", 100, "Number of lines to show from the end (default 100)")
}
