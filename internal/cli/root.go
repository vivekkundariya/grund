package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yourorg/grund/internal/application/wiring"
)

var container *wiring.Container

var rootCmd = &cobra.Command{
	Use:   "grund",
	Short: "Grund - Local development orchestration tool",
	Long: `Grund is a CLI tool that enables developers to selectively spin up 
services and their dependencies with a single command.`,
	Version: "0.1.0",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize dependency injection container
		root, err := getOrchestrationRoot()
		if err != nil {
			return err
		}

		container, err = wiring.NewContainer(root)
		if err != nil {
			return fmt.Errorf("failed to initialize container: %w", err)
		}

		return nil
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(cloneCmd)
}

// getOrchestrationRoot returns the path to the orchestration repo root
func getOrchestrationRoot() (string, error) {
	// Try to find services.yaml in current directory or parent
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Look for services.yaml in current and parent directories
	path := wd
	for i := 0; i < 5; i++ {
		servicesPath := filepath.Join(path, "services.yaml")
		if _, err := os.Stat(servicesPath); err == nil {
			return path, nil
		}
		path = filepath.Dir(path)
		if path == "/" || path == "." {
			break
		}
	}

	return "", fmt.Errorf("services.yaml not found. Are you in the orchestration repo?")
}
