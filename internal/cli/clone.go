package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone all registered service repos",
	Long:  `Clone all service repositories listed in services.yaml to their configured paths.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement clone command
		fmt.Println("Cloning registered services...")
		return nil
	},
}
