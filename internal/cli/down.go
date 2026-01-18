package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourorg/grund/internal/application/commands"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop all running services",
	Long:  `Stop all services and infrastructure that were started by grund.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if container == nil {
			return fmt.Errorf("container not initialized")
		}

		downCmd := commands.DownCommand{}
		return container.DownCommandHandler.Handle(cmd.Context(), downCmd)
	},
}
