package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vivekkundariya/grund/internal/application/commands"
)

var restartBuild bool

var restartCmd = &cobra.Command{
	Use:   "restart [services...]",
	Short: "Restart specific service(s)",
	Long:  `Restart one or more services while keeping infrastructure running.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if container == nil {
			return fmt.Errorf("container not initialized")
		}

		restartCmd := commands.RestartCommand{
			ServiceNames: args,
			Build:        restartBuild,
		}
		return container.RestartCommandHandler.Handle(cmd.Context(), restartCmd)
	},
}

func init() {
	restartCmd.Flags().BoolVar(&restartBuild, "build", false, "Rebuild containers before restarting")
}
