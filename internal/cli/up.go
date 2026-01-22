package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vivekkundariya/grund/internal/application/commands"
)

var (
	upNoDeps      bool
	upInfraOnly   bool
	upBuild       bool
	upLocal       bool
)

var upCmd = &cobra.Command{
	Use:   "up [services...]",
	Short: "Start services and dependencies",
	Long: `Start one or more services along with all their dependencies.
Infrastructure will be started first, followed by services in dependency order.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if container == nil {
			return fmt.Errorf("container not initialized")
		}

		upCmd := commands.UpCommand{
			ServiceNames: args,
			NoDeps:       upNoDeps,
			InfraOnly:    upInfraOnly,
			Build:        upBuild,
			Local:        upLocal,
		}

		return container.UpCommandHandler.Handle(cmd.Context(), upCmd)
	},
}

func init() {
	upCmd.Flags().BoolVar(&upNoDeps, "no-deps", false, "Only start specified services, no dependencies")
	upCmd.Flags().BoolVar(&upInfraOnly, "infra-only", false, "Only start infrastructure, no services")
	upCmd.Flags().BoolVar(&upBuild, "build", false, "Force rebuild containers")
	upCmd.Flags().BoolVar(&upLocal, "local", false, "Run service locally (not in container)")
}
