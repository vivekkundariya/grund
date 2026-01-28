package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vivekkundariya/grund/internal/application/wiring"
	"github.com/vivekkundariya/grund/internal/config"
	"github.com/vivekkundariya/grund/internal/ui"
)

var (
	container      *wiring.Container
	configResolver *config.ConfigResolver
	
	// CLI flags
	configFile string
	verbose    bool
)

var rootCmd = &cobra.Command{
	Use:   "grund",
	Short: "Grund - Local development orchestration tool",
	Long: `Grund is a CLI tool that enables developers to selectively spin up
services and their dependencies with a single command.

Configuration:
  1. --config flag (explicit path)
  2. GRUND_CONFIG environment variable
  3. Default: ~/.grund/services.yaml

Examples:
  grund up service-a service-b    Start services with dependencies
  grund down                      Stop all services
  grund status                    Show service status
  grund secrets list myservice    Check required secrets`,
	Version: "1.0.0",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Set verbose mode on logger
		ui.SetVerbose(verbose)

		// Skip initialization for help and version commands
		if cmd.Name() == "help" || cmd.Name() == "version" {
			return nil
		}

		// Also skip for completion commands
		if cmd.Name() == "completion" {
			return nil
		}

		// Skip for init, setup, and down commands (don't need existing services.yaml)
		// down can work without a project context by stopping all projects
		if cmd.Name() == "init" || cmd.Name() == "setup" || cmd.Name() == "down" {
			return nil
		}

		// Initialize config resolver
		var err error
		configResolver, err = config.NewConfigResolver(configFile)
		if err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}

		// Resolve services file and orchestration root
		servicesPath, orchestrationRoot, err := configResolver.ResolveServicesFile()
		if err != nil {
			return err
		}

		ui.Debug("Using config: %s", servicesPath)
		ui.Debug("Orchestration root: %s", orchestrationRoot)

		// Initialize dependency injection container
		container, err = wiring.NewContainerWithConfig(orchestrationRoot, servicesPath, configResolver)
		if err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}

		return nil
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", 
		"Path to services registry file (default: auto-detect services.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, 
		"Enable verbose output")
	
	// Add subcommands
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(secretsCmd)
	rootCmd.AddCommand(newSetupCmd())
}

// newSetupCmd creates the setup subcommand for global configuration
func newSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Initialize global Grund configuration",
		Long: `Initialize the global Grund configuration at ~/.grund/config.yaml.

This is a one-time setup for your machine. It creates a default configuration
file that you can customize with your preferred paths and settings.

For initializing a service with grund.yaml, use 'grund init' instead.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.InitGlobalConfig(); err != nil {
				return fmt.Errorf("failed to initialize config: %w", err)
			}

			configPath, _ := config.GetGlobalConfigPath()
			fmt.Printf("Global config initialized at: %s\n", configPath)
			fmt.Println("\nYou can customize this file to set default paths and options.")
			return nil
		},
	}
}

// GetConfigResolver returns the current config resolver (for use by subcommands)
func GetConfigResolver() *config.ConfigResolver {
	return configResolver
}
