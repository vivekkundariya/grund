package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	logsFollow bool
	logsTail   int
)

var logsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "View aggregated or per-service logs",
	Long:  `View logs from all services or a specific service.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement logs command
		service := ""
		if len(args) > 0 {
			service = args[0]
		}
		fmt.Printf("Viewing logs for: %s\n", service)
		if logsFollow {
			fmt.Println("  (following)")
		}
		if logsTail > 0 {
			fmt.Printf("  (last %d lines)\n", logsTail)
		}
		return nil
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")
	logsCmd.Flags().IntVar(&logsTail, "tail", 0, "Number of lines to show from the end")
}
