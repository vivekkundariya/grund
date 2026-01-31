package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vivekkundariya/grund/internal/cli/skills"
	"github.com/vivekkundariya/grund/internal/config"
	"github.com/vivekkundariya/grund/internal/ui"
)

var (
	initSkipAI bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Grund and set up AI assistant skills",
	Long: `Initialize Grund global configuration and optionally set up AI assistant skills.

This command performs two main tasks:
  1. Creates or updates the global config at ~/.grund/config.yaml
  2. Installs AI assistant skill files (CLAUDE.md, MCP configs) for supported tools

The AI skill setup helps AI coding assistants understand your Grund configuration
and provide better assistance with local development orchestration.

Examples:
  grund init              # Full initialization with AI skills
  grund init --skip-ai    # Only initialize config, skip AI setup`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initSkipAI, "skip-ai", false, "Skip AI assistant skill setup")
}

func runInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	ui.Header("Grund Initialization")
	fmt.Println()

	// Step 1: Handle config initialization
	if err := handleConfigInit(reader); err != nil {
		return fmt.Errorf("config initialization failed: %w", err)
	}

	// Step 2: Handle AI skills installation (unless skipped)
	if !initSkipAI {
		fmt.Println()
		if err := handleAISkillsInit(reader); err != nil {
			return fmt.Errorf("AI skills setup failed: %w", err)
		}
	} else {
		ui.Infof("Skipping AI assistant skill setup (--skip-ai)")
	}

	fmt.Println()
	ui.Successf("Grund initialization complete!")
	return nil
}

// handleConfigInit handles the global config initialization.
func handleConfigInit(reader *bufio.Reader) error {
	ui.Step("Initializing global configuration...")

	configPath, _ := config.GetGlobalConfigPath()

	if config.GlobalConfigExists() {
		ui.Infof("Global config already exists at: %s", configPath)

		if promptYesNo(reader, "Do you want to re-initialize the config?", false) {
			if err := config.ForceInitGlobalConfig(); err != nil {
				return fmt.Errorf("failed to re-initialize config: %w", err)
			}
			ui.Successf("Config re-initialized at: %s", configPath)
		} else {
			ui.Infof("Keeping existing config")
		}
	} else {
		if err := config.InitGlobalConfig(); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}
		ui.Successf("Config created at: %s", configPath)
	}

	return nil
}

// handleAISkillsInit handles the AI assistant skills installation.
func handleAISkillsInit(reader *bufio.Reader) error {
	ui.Step("Setting up AI assistant skills...")

	// Check what's already installed
	claudeInstalled := skills.IsInstalled(skills.Claude)
	cursorInstalled := skills.IsInstalled(skills.Cursor)

	if claudeInstalled && cursorInstalled {
		ui.Successf("AI assistant skills already configured for Claude Code and Cursor")
		return nil
	}

	// Report already installed
	if claudeInstalled {
		ui.Infof("Claude Code skill already installed")
	}
	if cursorInstalled {
		ui.Infof("Cursor skill already installed")
	}

	// Ask if user wants to set up skills
	if !promptYesNo(reader, "Would you like to set up AI assistant skills for Grund?", true) {
		ui.Infof("Skipping AI assistant setup")
		return nil
	}

	// Determine which assistants to offer
	var toInstall []skills.AIAssistant

	if !claudeInstalled && !cursorInstalled {
		// Both available - ask which
		choice := promptChoice(reader, "Which AI assistant?", []string{"Claude", "Cursor", "Both"}, "Both")
		switch choice {
		case "Claude":
			toInstall = []skills.AIAssistant{skills.Claude}
		case "Cursor":
			toInstall = []skills.AIAssistant{skills.Cursor}
		default: // "Both"
			toInstall = []skills.AIAssistant{skills.Claude, skills.Cursor}
		}
	} else if !claudeInstalled {
		toInstall = []skills.AIAssistant{skills.Claude}
	} else {
		toInstall = []skills.AIAssistant{skills.Cursor}
	}

	// Install selected assistants
	for _, assistant := range toInstall {
		if err := skills.Install(assistant); err != nil {
			return fmt.Errorf("failed to install %s skill: %w", skills.AssistantName(assistant), err)
		}
		path, _ := skills.SkillPaths(assistant)
		ui.Successf("%s skill installed at: %s", skills.AssistantName(assistant), path)
	}

	return nil
}

// promptYesNo prompts the user with a yes/no question and returns their choice.
// The question should not include the [y/n] suffix - it will be added automatically.
// Shows [Y/n] if defaultVal is true, [y/N] if false.
// Returns defaultVal on empty input or error.
func promptYesNo(reader *bufio.Reader, question string, defaultVal bool) bool {
	defaultStr := "y/N"
	if defaultVal {
		defaultStr = "Y/n"
	}

	fmt.Printf("%s [%s]: ", question, defaultStr)

	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultVal
	}

	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return defaultVal
	}

	return input == "y" || input == "yes"
}

// promptChoice prompts the user to select from a list of choices.
// Returns the selected string value, or defaultVal on empty/invalid input.
func promptChoice(reader *bufio.Reader, question string, choices []string, defaultVal string) string {
	fmt.Printf("%s (%s) [%s]: ", question, strings.Join(choices, "/"), defaultVal)

	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultVal
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}

	// Validate choice
	for _, c := range choices {
		if input == c {
			return input
		}
	}

	fmt.Printf("  Invalid choice, using default: %s\n", defaultVal)
	return defaultVal
}
