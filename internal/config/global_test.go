package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetGrundHome_Default(t *testing.T) {
	// Clear env var to test default
	os.Unsetenv(EnvGrundHome)

	home, err := GetGrundHome()
	if err != nil {
		t.Fatalf("GetGrundHome() error: %v", err)
	}

	userHome, _ := os.UserHomeDir()
	expected := filepath.Join(userHome, GlobalConfigDir)

	if home != expected {
		t.Errorf("GetGrundHome() = %q, want %q", home, expected)
	}
}

func TestGetGrundHome_EnvVar(t *testing.T) {
	customHome := "/custom/grund/home"
	os.Setenv(EnvGrundHome, customHome)
	defer os.Unsetenv(EnvGrundHome)

	home, err := GetGrundHome()
	if err != nil {
		t.Fatalf("GetGrundHome() error: %v", err)
	}

	if home != customHome {
		t.Errorf("GetGrundHome() = %q, want %q", home, customHome)
	}
}

func TestLoadGlobalConfig_NotExists(t *testing.T) {
	// Use temp directory as GRUND_HOME
	tmpDir := t.TempDir()
	os.Setenv(EnvGrundHome, tmpDir)
	defer os.Unsetenv(EnvGrundHome)

	config, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("LoadGlobalConfig() error: %v", err)
	}

	// Should return defaults when file doesn't exist
	if config.DefaultServicesFile != DefaultConfigFileName {
		t.Errorf("Expected default services file, got %q", config.DefaultServicesFile)
	}
}

func TestSaveAndLoadGlobalConfig(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv(EnvGrundHome, tmpDir)
	defer os.Unsetenv(EnvGrundHome)

	// Create custom config
	config := &GlobalConfig{
		DefaultServicesFile:      "my-services.yaml",
		DefaultOrchestrationRepo: "~/my-orchestration",
		ServicesBasePath:         "~/my-projects",
		Docker: DockerConfig{
			ComposeCommand: "docker-compose",
		},
		LocalStack: LocalStackConfig{
			Endpoint: "http://localhost:4567",
			Region:   "eu-west-1",
		},
	}

	// Save it
	if err := SaveGlobalConfig(config); err != nil {
		t.Fatalf("SaveGlobalConfig() error: %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(tmpDir, GlobalConfigFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load it back
	loaded, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("LoadGlobalConfig() error: %v", err)
	}

	// Verify values
	if loaded.DefaultServicesFile != "my-services.yaml" {
		t.Errorf("DefaultServicesFile = %q, want 'my-services.yaml'", loaded.DefaultServicesFile)
	}
	if loaded.DefaultOrchestrationRepo != "~/my-orchestration" {
		t.Errorf("DefaultOrchestrationRepo = %q, want '~/my-orchestration'", loaded.DefaultOrchestrationRepo)
	}
	if loaded.Docker.ComposeCommand != "docker-compose" {
		t.Errorf("Docker.ComposeCommand = %q, want 'docker-compose'", loaded.Docker.ComposeCommand)
	}
	if loaded.LocalStack.Endpoint != "http://localhost:4567" {
		t.Errorf("LocalStack.Endpoint = %q, want 'http://localhost:4567'", loaded.LocalStack.Endpoint)
	}
}

func TestInitGlobalConfig(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv(EnvGrundHome, tmpDir)
	defer os.Unsetenv(EnvGrundHome)

	if err := InitGlobalConfig(); err != nil {
		t.Fatalf("InitGlobalConfig() error: %v", err)
	}

	// Verify file was created
	configPath := filepath.Join(tmpDir, GlobalConfigFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Calling again should be safe (idempotent)
	if err := InitGlobalConfig(); err != nil {
		t.Fatalf("InitGlobalConfig() second call error: %v", err)
	}
}

func TestGlobalConfig_ApplyDefaults(t *testing.T) {
	config := &GlobalConfig{
		// Leave everything empty
	}

	config.applyDefaults()

	if config.DefaultServicesFile != DefaultConfigFileName {
		t.Errorf("DefaultServicesFile should have default, got %q", config.DefaultServicesFile)
	}
	if config.Docker.ComposeCommand != "docker compose" {
		t.Errorf("Docker.ComposeCommand should have default, got %q", config.Docker.ComposeCommand)
	}
	if config.LocalStack.Endpoint != "http://localhost:4566" {
		t.Errorf("LocalStack.Endpoint should have default, got %q", config.LocalStack.Endpoint)
	}
	if config.LocalStack.Region != "us-east-1" {
		t.Errorf("LocalStack.Region should have default, got %q", config.LocalStack.Region)
	}
}

func TestGlobalConfig_PartialConfig(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv(EnvGrundHome, tmpDir)
	defer os.Unsetenv(EnvGrundHome)

	// Write partial config (only some fields)
	configPath := filepath.Join(tmpDir, GlobalConfigFile)
	partialConfig := `
default_services_file: custom.yaml
docker:
  compose_command: podman-compose
`
	os.WriteFile(configPath, []byte(partialConfig), 0644)

	config, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("LoadGlobalConfig() error: %v", err)
	}

	// Custom values should be loaded
	if config.DefaultServicesFile != "custom.yaml" {
		t.Errorf("Expected custom services file, got %q", config.DefaultServicesFile)
	}
	if config.Docker.ComposeCommand != "podman-compose" {
		t.Errorf("Expected custom compose command, got %q", config.Docker.ComposeCommand)
	}

	// Unset values should have defaults
	if config.LocalStack.Endpoint != "http://localhost:4566" {
		t.Errorf("Expected default localstack endpoint, got %q", config.LocalStack.Endpoint)
	}
}
