package cli

import (
	"fmt"
	"strings"

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

		// Define column widths
		const (
			nameWidth   = 20
			statusWidth = 15
		)

		// Print header
		fmt.Println()
		fmt.Printf("  %s%s\n",
			padRight("SERVICE", nameWidth),
			"STATUS")
		fmt.Printf("  %s%s\n",
			strings.Repeat("─", nameWidth-2)+"  ",
			strings.Repeat("─", statusWidth-2))

		for _, s := range statuses {
			// Get status icon and color
			var statusIcon, statusColor string
			switch s.Status {
			case "running":
				statusIcon = "●"
				statusColor = ui.Green
			case "exited":
				statusIcon = "●"
				statusColor = ui.Red
			default:
				statusIcon = "○"
				statusColor = ui.Yellow
			}

			// Print row with proper padding (pad BEFORE colorizing)
			fmt.Printf("  %s%s%s %s%s\n",
				padRight(s.Name, nameWidth),
				statusColor, statusIcon,
				s.Status, ui.Reset)
		}
		fmt.Println()

		return nil
	},
}

// padRight pads a string to the right with spaces
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
