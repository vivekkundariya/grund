package configcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vivekkundariya/grund/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize global Grund configuration",
	Long: `Initialize the global Grund configuration at ~/.grund/config.yaml.

This is a one-time setup for your machine. It creates a default configuration
file that you can customize with your preferred paths and settings.

For initializing a service with grund.yaml, use 'grund service init' instead.

Example:
  grund config init`,
	RunE: runConfigInit,
}

func init() {
	Cmd.AddCommand(initCmd)
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	if err := config.InitGlobalConfig(); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	configPath, _ := config.GetGlobalConfigPath()
	fmt.Printf("Global config initialized at: %s\n", configPath)
	fmt.Println("\nYou can customize this file to set default paths and options.")
	return nil
}
