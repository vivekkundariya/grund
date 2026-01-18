package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vivekkundariya/grund/internal/application/queries"
)

var configCmd = &cobra.Command{
	Use:   "config [service]",
	Short: "Validate and show resolved config",
	Long:  `Show the resolved configuration for a service, including all environment variables and dependencies.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if container == nil {
			return fmt.Errorf("container not initialized")
		}

		service := ""
		if len(args) > 0 {
			service = args[0]
		}

		query := queries.ConfigQuery{ServiceName: service}
		config, err := container.ConfigQueryHandler.Handle(query)
		if err != nil {
			return err
		}

		// TODO: Format and display the config
		_ = config
		return nil
	},
}
