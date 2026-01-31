package skills

import (
	"fmt"
	"os"
	"path/filepath"
)

// AIAssistant represents a supported AI assistant type
type AIAssistant int

const (
	// Claude represents Claude Code assistant
	Claude AIAssistant = iota
	// Cursor represents Cursor assistant
	Cursor
)

// SkillPaths returns the installation path for the given assistant
func SkillPaths(assistant AIAssistant) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	switch assistant {
	case Claude:
		return filepath.Join(home, ".claude", "skills", "using-grund"), nil
	case Cursor:
		return filepath.Join(home, ".cursor", "skills-cursor", "using-grund"), nil
	default:
		return "", fmt.Errorf("unknown assistant type: %d", assistant)
	}
}

// IsInstalled checks if skill is already installed for the given assistant
func IsInstalled(assistant AIAssistant) bool {
	path, err := SkillPaths(assistant)
	if err != nil {
		return false
	}
	skillFile := filepath.Join(path, "SKILL.md")
	_, err = os.Stat(skillFile)
	return err == nil
}

// Install installs the skill for the given assistant
func Install(assistant AIAssistant) error {
	path, err := SkillPaths(assistant)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	skillFile := filepath.Join(path, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte(GrundSkillContent), 0644); err != nil {
		return fmt.Errorf("failed to write skill file: %w", err)
	}

	return nil
}

// Uninstall removes the skill for the given assistant
func Uninstall(assistant AIAssistant) error {
	path, err := SkillPaths(assistant)
	if err != nil {
		return err
	}

	if !IsInstalled(assistant) {
		return nil // Already uninstalled
	}

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove skill directory: %w", err)
	}

	return nil
}

// AssistantName returns human-readable name for the assistant
func AssistantName(assistant AIAssistant) string {
	switch assistant {
	case Claude:
		return "Claude Code"
	case Cursor:
		return "Cursor"
	default:
		return "Unknown"
	}
}

// AllAssistants returns all supported AI assistants
func AllAssistants() []AIAssistant {
	return []AIAssistant{Claude, Cursor}
}
