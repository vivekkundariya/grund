package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigResolver_ResolveServicesFile_CLIFlag(t *testing.T) {
	// Create temp services file
	tmpDir := t.TempDir()
	servicesFile := filepath.Join(tmpDir, "custom-services.yaml")
	if err := os.WriteFile(servicesFile, []byte("version: 1\nservices: {}"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	resolver, err := NewConfigResolver(servicesFile)
	if err != nil {
		t.Fatalf("NewConfigResolver() error: %v", err)
	}

	path, root, err := resolver.ResolveServicesFile()
	if err != nil {
		t.Fatalf("ResolveServicesFile() error: %v", err)
	}

	if path != servicesFile {
		t.Errorf("Expected path %s, got %s", servicesFile, path)
	}

	if root != tmpDir {
		t.Errorf("Expected root %s, got %s", tmpDir, root)
	}
}

func TestConfigResolver_ResolveServicesFile_EnvVar(t *testing.T) {
	// Create temp services file
	tmpDir := t.TempDir()
	servicesFile := filepath.Join(tmpDir, "env-services.yaml")
	if err := os.WriteFile(servicesFile, []byte("version: 1\nservices: {}"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Set environment variable
	os.Setenv(EnvConfigFile, servicesFile)
	defer os.Unsetenv(EnvConfigFile)

	resolver, err := NewConfigResolver("") // No CLI flag
	if err != nil {
		t.Fatalf("NewConfigResolver() error: %v", err)
	}

	path, root, err := resolver.ResolveServicesFile()
	if err != nil {
		t.Fatalf("ResolveServicesFile() error: %v", err)
	}

	if path != servicesFile {
		t.Errorf("Expected path %s, got %s", servicesFile, path)
	}

	if root != tmpDir {
		t.Errorf("Expected root %s, got %s", tmpDir, root)
	}
}

func TestConfigResolver_ResolveServicesFile_CLIOverridesEnv(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two different files
	envFile := filepath.Join(tmpDir, "env-services.yaml")
	cliFile := filepath.Join(tmpDir, "cli-services.yaml")

	os.WriteFile(envFile, []byte("version: 1\nservices: {}"), 0644)
	os.WriteFile(cliFile, []byte("version: 1\nservices: {}"), 0644)

	// Set environment variable
	os.Setenv(EnvConfigFile, envFile)
	defer os.Unsetenv(EnvConfigFile)

	// CLI flag should take priority
	resolver, _ := NewConfigResolver(cliFile)
	path, _, err := resolver.ResolveServicesFile()

	if err != nil {
		t.Fatalf("ResolveServicesFile() error: %v", err)
	}

	if path != cliFile {
		t.Errorf("CLI flag should override env var. Expected %s, got %s", cliFile, path)
	}
}

func TestConfigResolver_ResolveServicesFile_LocalFile(t *testing.T) {
	// Create temp directory with services.yaml
	tmpDir := t.TempDir()
	servicesFile := filepath.Join(tmpDir, "services.yaml")
	if err := os.WriteFile(servicesFile, []byte("version: 1\nservices: {}"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Change to temp directory
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	resolver, err := NewConfigResolver("") // No CLI flag, no env var
	if err != nil {
		t.Fatalf("NewConfigResolver() error: %v", err)
	}

	path, root, err := resolver.ResolveServicesFile()
	if err != nil {
		t.Fatalf("ResolveServicesFile() error: %v", err)
	}

	// Resolve symlinks for comparison (macOS /var -> /private/var)
	expectedPath, _ := filepath.EvalSymlinks(servicesFile)
	expectedRoot, _ := filepath.EvalSymlinks(tmpDir)
	actualPath, _ := filepath.EvalSymlinks(path)
	actualRoot, _ := filepath.EvalSymlinks(root)

	if actualPath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, actualPath)
	}

	if actualRoot != expectedRoot {
		t.Errorf("Expected root %s, got %s", expectedRoot, actualRoot)
	}
}

func TestConfigResolver_ResolveServicesFile_AlternativeNames(t *testing.T) {
	tmpDir := t.TempDir()

	// Create grund-services.yaml (alternative name)
	servicesFile := filepath.Join(tmpDir, "grund-services.yaml")
	if err := os.WriteFile(servicesFile, []byte("version: 1\nservices: {}"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	resolver, _ := NewConfigResolver("")
	path, _, err := resolver.ResolveServicesFile()

	if err != nil {
		t.Fatalf("ResolveServicesFile() error: %v", err)
	}

	// Resolve symlinks for comparison
	expectedPath, _ := filepath.EvalSymlinks(servicesFile)
	actualPath, _ := filepath.EvalSymlinks(path)

	if actualPath != expectedPath {
		t.Errorf("Expected alternative name to work. Expected %s, got %s", expectedPath, actualPath)
	}
}

func TestConfigResolver_ResolveServicesFile_ParentDirectory(t *testing.T) {
	// Create parent/child directory structure
	tmpDir := t.TempDir()
	childDir := filepath.Join(tmpDir, "child")
	os.MkdirAll(childDir, 0755)

	// Put services.yaml in parent
	servicesFile := filepath.Join(tmpDir, "services.yaml")
	os.WriteFile(servicesFile, []byte("version: 1\nservices: {}"), 0644)

	// Change to child directory
	originalWd, _ := os.Getwd()
	os.Chdir(childDir)
	defer os.Chdir(originalWd)

	resolver, _ := NewConfigResolver("")
	path, root, err := resolver.ResolveServicesFile()

	if err != nil {
		t.Fatalf("ResolveServicesFile() error: %v", err)
	}

	// Resolve symlinks for comparison
	expectedPath, _ := filepath.EvalSymlinks(servicesFile)
	expectedRoot, _ := filepath.EvalSymlinks(tmpDir)
	actualPath, _ := filepath.EvalSymlinks(path)
	actualRoot, _ := filepath.EvalSymlinks(root)

	if actualPath != expectedPath {
		t.Errorf("Should find services.yaml in parent. Expected %s, got %s", expectedPath, actualPath)
	}

	if actualRoot != expectedRoot {
		t.Errorf("Root should be parent dir. Expected %s, got %s", expectedRoot, actualRoot)
	}
}

func TestConfigResolver_ResolveServicesFile_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	// Override GRUND_HOME to prevent finding global config
	originalHome := os.Getenv(EnvGrundHome)
	os.Setenv(EnvGrundHome, filepath.Join(tmpDir, "nonexistent-grund-home"))
	defer func() {
		if originalHome == "" {
			os.Unsetenv(EnvGrundHome)
		} else {
			os.Setenv(EnvGrundHome, originalHome)
		}
	}()

	resolver, _ := NewConfigResolver("")
	_, _, err := resolver.ResolveServicesFile()

	if err == nil {
		t.Error("Expected error when services file not found")
	}
}

func TestConfigResolver_CLIFlagFileNotFound(t *testing.T) {
	resolver, _ := NewConfigResolver("/nonexistent/services.yaml")
	_, _, err := resolver.ResolveServicesFile()

	if err == nil {
		t.Error("Expected error when CLI flag points to nonexistent file")
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/projects", filepath.Join(home, "projects")},
		{"~", home},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := expandPath(tt.input)
			if result != tt.expected {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDefaultGlobalConfig(t *testing.T) {
	config := DefaultGlobalConfig()

	if config.DefaultServicesFile != "services.yaml" {
		t.Errorf("Expected default services file 'services.yaml', got %q", config.DefaultServicesFile)
	}

	if config.Docker.ComposeCommand != "docker compose" {
		t.Errorf("Expected docker compose command, got %q", config.Docker.ComposeCommand)
	}

	if config.LocalStack.Endpoint != "http://localhost:4566" {
		t.Errorf("Expected localstack endpoint, got %q", config.LocalStack.Endpoint)
	}
}

func TestConfigResolver_GetSettings(t *testing.T) {
	resolver, _ := NewConfigResolver("")

	if endpoint := resolver.GetLocalStackEndpoint(); endpoint != "http://localhost:4566" {
		t.Errorf("Expected default localstack endpoint, got %q", endpoint)
	}

	if region := resolver.GetLocalStackRegion(); region != "us-east-1" {
		t.Errorf("Expected default region, got %q", region)
	}

	if cmd := resolver.GetDockerComposeCommand(); cmd != "docker compose" {
		t.Errorf("Expected default docker compose command, got %q", cmd)
	}
}
