package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillPaths(t *testing.T) {
	tests := []struct {
		assistant AIAssistant
		contains  string
	}{
		{Claude, ".claude/skills/using-grund"},
		{Cursor, ".cursor/skills-cursor/using-grund"},
	}

	for _, tt := range tests {
		path, err := SkillPaths(tt.assistant)
		if err != nil {
			t.Errorf("SkillPaths(%v) returned error: %v", tt.assistant, err)
			continue
		}
		if !filepath.IsAbs(path) {
			t.Errorf("SkillPaths(%v) returned non-absolute path: %s", tt.assistant, path)
		}
		if !strings.Contains(path, tt.contains) {
			t.Errorf("SkillPaths(%v) = %s, expected to contain %s", tt.assistant, path, tt.contains)
		}
	}
}

func TestSkillPathsUnknownAssistant(t *testing.T) {
	unknownAssistant := AIAssistant(999)
	_, err := SkillPaths(unknownAssistant)
	if err == nil {
		t.Error("SkillPaths() with unknown assistant should return error")
	}
	if !strings.Contains(err.Error(), "unknown assistant type") {
		t.Errorf("expected 'unknown assistant type' error, got: %v", err)
	}
}

func TestAssistantName(t *testing.T) {
	tests := []struct {
		assistant AIAssistant
		expected  string
	}{
		{Claude, "Claude Code"},
		{Cursor, "Cursor"},
	}

	for _, tt := range tests {
		name := AssistantName(tt.assistant)
		if name != tt.expected {
			t.Errorf("AssistantName(%v) = %s, expected %s", tt.assistant, name, tt.expected)
		}
	}
}

func TestAssistantNameUnknown(t *testing.T) {
	unknownAssistant := AIAssistant(999)
	name := AssistantName(unknownAssistant)
	if name != "Unknown" {
		t.Errorf("AssistantName(unknown) = %s, expected 'Unknown'", name)
	}
}

func TestInstallAndUninstall(t *testing.T) {
	// Create a temp directory to simulate home
	tmpDir := t.TempDir()

	// Save original home and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tmpDir)

	// Test install for Claude
	err := Install(Claude)
	if err != nil {
		t.Fatalf("Install(Claude) returned error: %v", err)
	}

	// Verify file was created
	expectedPath := filepath.Join(tmpDir, ".claude", "skills", "using-grund", "SKILL.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Install(Claude) did not create SKILL.md at %s", expectedPath)
	}

	// Verify content
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read installed SKILL.md: %v", err)
	}
	if string(content) != GrundSkillContent {
		t.Error("installed SKILL.md content does not match GrundSkillContent")
	}

	// Verify IsInstalled returns true
	if !IsInstalled(Claude) {
		t.Error("IsInstalled(Claude) returned false after Install")
	}

	// Test uninstall
	err = Uninstall(Claude)
	if err != nil {
		t.Fatalf("Uninstall(Claude) returned error: %v", err)
	}

	// Verify directory was removed
	if _, err := os.Stat(expectedPath); !os.IsNotExist(err) {
		t.Error("Uninstall(Claude) did not remove SKILL.md")
	}

	// Verify IsInstalled returns false
	if IsInstalled(Claude) {
		t.Error("IsInstalled(Claude) returned true after Uninstall")
	}
}

func TestUninstallWhenNotInstalled(t *testing.T) {
	// Create a temp directory to simulate home
	tmpDir := t.TempDir()

	// Save original home and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tmpDir)

	// Uninstall when not installed should not error
	err := Uninstall(Claude)
	if err != nil {
		t.Errorf("Uninstall(Claude) when not installed returned error: %v", err)
	}
}

func TestIsInstalledNotInstalled(t *testing.T) {
	// Create a temp directory to simulate home
	tmpDir := t.TempDir()

	// Save original home and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tmpDir)

	// Should return false when not installed
	if IsInstalled(Claude) {
		t.Error("IsInstalled(Claude) returned true when not installed")
	}
	if IsInstalled(Cursor) {
		t.Error("IsInstalled(Cursor) returned true when not installed")
	}
}

func TestAllAssistants(t *testing.T) {
	assistants := AllAssistants()
	if len(assistants) != 2 {
		t.Errorf("AllAssistants() returned %d assistants, expected 2", len(assistants))
	}

	// Verify both Claude and Cursor are included
	hasClaude := false
	hasCursor := false
	for _, a := range assistants {
		if a == Claude {
			hasClaude = true
		}
		if a == Cursor {
			hasCursor = true
		}
	}

	if !hasClaude {
		t.Error("AllAssistants() missing Claude")
	}
	if !hasCursor {
		t.Error("AllAssistants() missing Cursor")
	}
}

func TestGrundSkillContent(t *testing.T) {
	// Verify the skill content has expected sections
	if !strings.Contains(GrundSkillContent, "name: using-grund") {
		t.Error("GrundSkillContent missing frontmatter name")
	}
	if !strings.Contains(GrundSkillContent, "grund up") {
		t.Error("GrundSkillContent missing 'grund up' command")
	}
	if !strings.Contains(GrundSkillContent, "grund service init") {
		t.Error("GrundSkillContent missing 'grund service init' command")
	}
	if !strings.Contains(GrundSkillContent, "grund down") {
		t.Error("GrundSkillContent missing 'grund down' command")
	}
	if !strings.Contains(GrundSkillContent, "grund status") {
		t.Error("GrundSkillContent missing 'grund status' command")
	}
	if !strings.Contains(GrundSkillContent, "grund secrets") {
		t.Error("GrundSkillContent missing 'grund secrets' command")
	}
	if !strings.Contains(GrundSkillContent, "allowed-tools:") {
		t.Error("GrundSkillContent missing frontmatter allowed-tools")
	}
	if !strings.Contains(GrundSkillContent, "Bash(grund *)") {
		t.Error("GrundSkillContent missing Bash tool permission")
	}
}

func TestInstallForBothAssistants(t *testing.T) {
	// Create a temp directory to simulate home
	tmpDir := t.TempDir()

	// Save original home and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tmpDir)

	// Install for both assistants
	for _, assistant := range AllAssistants() {
		err := Install(assistant)
		if err != nil {
			t.Errorf("Install(%v) returned error: %v", assistant, err)
			continue
		}

		if !IsInstalled(assistant) {
			t.Errorf("IsInstalled(%v) returned false after Install", assistant)
		}
	}

	// Verify both paths exist
	claudePath := filepath.Join(tmpDir, ".claude", "skills", "using-grund", "SKILL.md")
	cursorPath := filepath.Join(tmpDir, ".cursor", "skills-cursor", "using-grund", "SKILL.md")

	if _, err := os.Stat(claudePath); os.IsNotExist(err) {
		t.Errorf("Claude SKILL.md not found at %s", claudePath)
	}
	if _, err := os.Stat(cursorPath); os.IsNotExist(err) {
		t.Errorf("Cursor SKILL.md not found at %s", cursorPath)
	}
}

func TestInstallIdempotent(t *testing.T) {
	// Create a temp directory to simulate home
	tmpDir := t.TempDir()

	// Save original home and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tmpDir)

	// Install twice should not error
	err := Install(Claude)
	if err != nil {
		t.Fatalf("First Install(Claude) returned error: %v", err)
	}

	err = Install(Claude)
	if err != nil {
		t.Errorf("Second Install(Claude) returned error: %v", err)
	}

	// Should still be installed
	if !IsInstalled(Claude) {
		t.Error("IsInstalled(Claude) returned false after double Install")
	}
}
