package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vivekkundariya/grund/internal/cli/prompts"
	"github.com/vivekkundariya/grund/internal/cli/skills"
	"github.com/vivekkundariya/grund/internal/config"
	"github.com/vivekkundariya/grund/internal/ui"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Grund and set up AI assistant skills",
	Long: `Initialize Grund global configuration and set up AI assistant skills.

This command performs two tasks:
  1. Creates global config at ~/.grund/config.yaml (or re-initializes if exists)
  2. Installs AI assistant skills for Claude Code and/or Cursor

Example:
  grund init`,
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	ui.Header("Grund Initialization")
	fmt.Println()

	// Step 1: Handle config initialization
	if err := handleConfigInit(); err != nil {
		return fmt.Errorf("config initialization failed: %w", err)
	}

	// Step 2: Handle AI skills installation
	fmt.Println()
	if err := handleAISkillsInit(); err != nil {
		return fmt.Errorf("AI skills setup failed: %w", err)
	}

	fmt.Println()
	ui.Successf("Grund initialization complete!")
	return nil
}

// handleConfigInit handles the global config initialization.
func handleConfigInit() error {
	ui.Step("Initializing global configuration...")

	configPath, _ := config.GetGlobalConfigPath()

	if config.GlobalConfigExists() {
		ui.Infof("Global config already exists at: %s", configPath)

		reinit, err := prompts.Confirm("Do you want to re-initialize the config?", false)
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}

		if reinit {
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
func handleAISkillsInit() error {
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
	setupSkills, err := prompts.Confirm("Would you like to set up AI assistant skills for Grund?", true)
	if err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}

	if !setupSkills {
		ui.Infof("Skipping AI assistant setup")
		return nil
	}

	// Determine which assistants to offer
	var toInstall []skills.AIAssistant

	if !claudeInstalled && !cursorInstalled {
		// Both available - ask which
		choice, err := prompts.Select("Which AI assistant?", []string{"Claude Code", "Cursor", "Both"}, "Both")
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}

		switch choice {
		case "Claude Code":
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
