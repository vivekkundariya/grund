package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize grund in current service",
	Long:  `Create a grund.yaml file in the current directory with an interactive setup.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement interactive init
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		configPath := filepath.Join(wd, "grund.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("grund.yaml already exists in %s", wd)
		}

		fmt.Printf("Initializing grund in: %s\n", wd)
		fmt.Println("Interactive setup - to be implemented")
		return nil
	},
}
