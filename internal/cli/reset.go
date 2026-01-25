package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/vivekkundariya/grund/internal/infrastructure/docker"
	"github.com/vivekkundariya/grund/internal/ui"
)

var (
	resetVolumes bool
	resetImages  bool
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Stop services and clean up resources",
	Long: `Stop all services and optionally remove volumes and images.

Examples:
  grund reset              # Stop all services
  grund reset -v           # Stop and remove volumes (database data)
  grund reset -v --images  # Stop, remove volumes and images (full cleanup)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if configResolver == nil {
			return fmt.Errorf("config not initialized")
		}

		_, orchestrationRoot, err := configResolver.ResolveServicesFile()
		if err != nil {
			return err
		}

		composeFile := docker.GetComposeFilePath(orchestrationRoot)

		// Check if compose file exists
		if _, err := os.Stat(composeFile); os.IsNotExist(err) {
			ui.Infof("No services to reset (no compose file found)")
			return nil
		}

		// Build docker compose down command
		dockerArgs := []string{"compose", "-f", composeFile, "down"}

		if resetVolumes {
			dockerArgs = append(dockerArgs, "-v")
			ui.Step("Stopping services and removing volumes...")
		} else {
			ui.Step("Stopping services...")
		}

		if resetImages {
			dockerArgs = append(dockerArgs, "--rmi", "local")
			ui.SubStep("Also removing locally built images")
		}

		// Execute docker compose down
		dockerCmd := exec.CommandContext(cmd.Context(), "docker", dockerArgs...)
		dockerCmd.Dir = orchestrationRoot
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr

		if err := dockerCmd.Run(); err != nil {
			return fmt.Errorf("failed to reset: %w", err)
		}

		ui.Successf("Reset complete")
		return nil
	},
}

func init() {
	resetCmd.Flags().BoolVarP(&resetVolumes, "volumes", "v", false, "Remove named volumes (database data will be lost)")
	resetCmd.Flags().BoolVar(&resetImages, "images", false, "Remove locally built images")
}
