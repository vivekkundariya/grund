package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/vivekkundariya/grund/internal/application/queries"
)

var configCmd = &cobra.Command{
	Use:   "config <service>",
	Short: "Show resolved configuration for a service",
	Long:  `Show the resolved configuration for a service, including environment variables, infrastructure, and dependencies.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if container == nil {
			return fmt.Errorf("container not initialized")
		}

		serviceName := args[0]

		query := queries.ConfigQuery{ServiceName: serviceName}
		config, err := container.ConfigQueryHandler.Handle(query)
		if err != nil {
			return err
		}

		// Service info
		fmt.Printf("\n  Service: %s\n", config.Service.Name)
		fmt.Printf("  Type:    %s\n", config.Service.Type)
		fmt.Printf("  Port:    %d\n", config.Service.Port.Value())
		fmt.Println()

		// Dependencies
		if len(config.Dependencies) > 0 {
			fmt.Printf("  Dependencies: %s\n", strings.Join(config.Dependencies, ", "))
		} else {
			fmt.Println("  Dependencies: none")
		}

		// Infrastructure
		if len(config.Infrastructure) > 0 {
			fmt.Printf("  Infrastructure: %s\n", strings.Join(config.Infrastructure, ", "))
		} else {
			fmt.Println("  Infrastructure: none")
		}
		fmt.Println()

		// Environment variables table
		if len(config.Environment) > 0 {
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.SetStyle(table.StyleRounded)
			t.AppendHeader(table.Row{"Environment Variable", "Value"})

			// Sort keys for consistent output
			keys := make([]string, 0, len(config.Environment))
			for k := range config.Environment {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				v := config.Environment[k]
				// Truncate long values
				if len(v) > 60 {
					v = v[:57] + "..."
				}
				t.AppendRow(table.Row{k, v})
			}

			t.Render()
		} else {
			fmt.Println("  No environment variables configured")
		}
		fmt.Println()

		return nil
	},
}
