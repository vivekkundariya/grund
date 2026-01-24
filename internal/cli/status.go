package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vivekkundariya/grund/internal/application/queries"
	"github.com/vivekkundariya/grund/internal/ui"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show running services and health",
	Long:  `Display the status of all running services and infrastructure, including health checks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if container == nil {
			return fmt.Errorf("container not initialized")
		}

		query := queries.StatusQuery{ServiceName: nil}
		statuses, err := container.StatusQueryHandler.Handle(cmd.Context(), query)
		if err != nil {
			return err
		}

		if len(statuses) == 0 {
			ui.Infof("No services found. Run 'grund up <service>' to start services.")
			return nil
		}

		// Print header
		fmt.Printf("\n%-20s %-12s %-10s\n", "SERVICE", "STATUS", "HEALTH")
		fmt.Printf("%-20s %-12s %-10s\n", "-------", "------", "------")

		for _, s := range statuses {
			statusStr := s.Status
			healthStr := s.Health

			// Colorize based on status
			if s.Status == "running" {
				statusStr = ui.Success(s.Status)
			} else if s.Status == "exited" {
				statusStr = ui.Error(s.Status)
			} else {
				statusStr = ui.Warning(s.Status)
			}

			// Colorize health
			if s.Health == "healthy" {
				healthStr = ui.Success(s.Health)
			} else if s.Health == "unhealthy" {
				healthStr = ui.Error(s.Health)
			}

			fmt.Printf("%-20s %-12s %-10s\n", s.Name, statusStr, healthStr)
		}
		fmt.Println()

		return nil
	},
}
