package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var resetService string

var resetCmd = &cobra.Command{
	Use:   "reset [--service SERVICE]",
	Short: "Stop, clean volumes, restart fresh",
	Long:  `Stop all services, remove volumes, and start fresh. Use --service to reset only a specific service's data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement reset command
		if resetService != "" {
			fmt.Printf("Resetting service: %s\n", resetService)
		} else {
			fmt.Println("Resetting all services...")
		}
		return nil
	},
}

func init() {
	resetCmd.Flags().StringVar(&resetService, "service", "", "Reset only specific service data")
}
