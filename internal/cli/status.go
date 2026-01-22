package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vivekkundariya/grund/internal/application/queries"
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

		// TODO: Format and display statuses
		_ = statuses
		return nil
	},
}
