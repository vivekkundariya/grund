package configcmd

import "github.com/spf13/cobra"

// Cmd is the parent command for configuration management
var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Grund configuration",
	Long: `Commands for managing Grund global and service configuration.

Examples:
  grund config init           Initialize global config (~/.grund/)
  grund config show           Show global configuration
  grund config show myservice Show service configuration`,
}
