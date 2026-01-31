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
	if config.Version != "1" {
		t.Errorf("Expected default version '1', got %q", config.Version)
	}
	if config.Services == nil {
		t.Error("Expected Services map to be initialized")
	}
}

func TestSaveAndLoadGlobalConfig(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv(EnvGrundHome, tmpDir)
	defer os.Unsetenv(EnvGrundHome)

	// Create custom config
	config := &GlobalConfig{
		Version: "1",
		Services: map[string]ServiceEntry{
			"user-service": {
				Path: "~/projects/user-service",
				Repo: "git@github.com:org/user-service.git",
			},
			"order-service": {
				Path: "~/projects/order-service",
			},
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
	if loaded.Version != "1" {
		t.Errorf("Version = %q, want '1'", loaded.Version)
	}
	if len(loaded.Services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(loaded.Services))
	}
	if entry, ok := loaded.Services["user-service"]; !ok {
		t.Error("Expected user-service to be present")
	} else {
		if entry.Path != "~/projects/user-service" {
			t.Errorf("user-service Path = %q, want '~/projects/user-service'", entry.Path)
		}
		if entry.Repo != "git@github.com:org/user-service.git" {
			t.Errorf("user-service Repo = %q, want 'git@github.com:org/user-service.git'", entry.Repo)
		}
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

func TestDefaultGlobalConfig(t *testing.T) {
	config := DefaultGlobalConfig()

	if config.Version != "1" {
		t.Errorf("Version should be '1', got %q", config.Version)
	}
	if config.Services == nil {
		t.Error("Services should be initialized")
	}
	if len(config.Services) != 0 {
		t.Errorf("Services should be empty, got %d entries", len(config.Services))
	}
}

func TestGettersWithDefaults(t *testing.T) {
	// Use temp directory with no config file
	tmpDir := t.TempDir()
	os.Setenv(EnvGrundHome, tmpDir)
	defer os.Unsetenv(EnvGrundHome)

	// Test that getter functions return default values when no config exists
	if GetLocalStackEndpoint() != "http://localhost:4566" {
		t.Errorf("GetLocalStackEndpoint() = %q, want 'http://localhost:4566'", GetLocalStackEndpoint())
	}
	if GetLocalStackRegion() != "us-east-1" {
		t.Errorf("GetLocalStackRegion() = %q, want 'us-east-1'", GetLocalStackRegion())
	}
	if GetDockerComposeCommand() != "docker compose" {
		t.Errorf("GetDockerComposeCommand() = %q, want 'docker compose'", GetDockerComposeCommand())
	}
}

func TestGettersWithCustomConfig(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv(EnvGrundHome, tmpDir)
	defer os.Unsetenv(EnvGrundHome)

	// Create config with custom values
	config := &GlobalConfig{
		Version:  "1",
		Services: make(map[string]ServiceEntry),
		Docker: &DockerConfig{
			ComposeCommand: "podman-compose",
		},
		LocalStack: &LocalStackConfig{
			Endpoint: "http://custom-localstack:4566",
			Region:   "eu-west-1",
		},
	}
	if err := SaveGlobalConfig(config); err != nil {
		t.Fatalf("SaveGlobalConfig() error: %v", err)
	}

	// Test that getter functions return custom values
	if GetLocalStackEndpoint() != "http://custom-localstack:4566" {
		t.Errorf("GetLocalStackEndpoint() = %q, want 'http://custom-localstack:4566'", GetLocalStackEndpoint())
	}
	if GetLocalStackRegion() != "eu-west-1" {
		t.Errorf("GetLocalStackRegion() = %q, want 'eu-west-1'", GetLocalStackRegion())
	}
	if GetDockerComposeCommand() != "podman-compose" {
		t.Errorf("GetDockerComposeCommand() = %q, want 'podman-compose'", GetDockerComposeCommand())
	}
}

func TestGettersWithPartialConfig(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv(EnvGrundHome, tmpDir)
	defer os.Unsetenv(EnvGrundHome)

	// Create config with only some values set
	config := &GlobalConfig{
		Version:  "1",
		Services: make(map[string]ServiceEntry),
		Docker: &DockerConfig{
			ComposeCommand: "docker-compose", // custom
		},
		// LocalStack not set - should use defaults
	}
	if err := SaveGlobalConfig(config); err != nil {
		t.Fatalf("SaveGlobalConfig() error: %v", err)
	}

	// Docker should be custom
	if GetDockerComposeCommand() != "docker-compose" {
		t.Errorf("GetDockerComposeCommand() = %q, want 'docker-compose'", GetDockerComposeCommand())
	}

	// LocalStack should use defaults
	if GetLocalStackEndpoint() != "http://localhost:4566" {
		t.Errorf("GetLocalStackEndpoint() = %q, want 'http://localhost:4566'", GetLocalStackEndpoint())
	}
	if GetLocalStackRegion() != "us-east-1" {
		t.Errorf("GetLocalStackRegion() = %q, want 'us-east-1'", GetLocalStackRegion())
	}
}

func TestGlobalConfigMethodGetters(t *testing.T) {
	// Test the method-based getters on GlobalConfig
	config := &GlobalConfig{
		Version:  "1",
		Services: make(map[string]ServiceEntry),
		Docker: &DockerConfig{
			ComposeCommand: "custom-compose",
		},
		LocalStack: &LocalStackConfig{
			Endpoint: "http://custom:4566",
			Region:   "ap-south-1",
		},
	}

	if config.GetDockerComposeCommand() != "custom-compose" {
		t.Errorf("GetDockerComposeCommand() = %q, want 'custom-compose'", config.GetDockerComposeCommand())
	}
	if config.GetLocalStackEndpoint() != "http://custom:4566" {
		t.Errorf("GetLocalStackEndpoint() = %q, want 'http://custom:4566'", config.GetLocalStackEndpoint())
	}
	if config.GetLocalStackRegion() != "ap-south-1" {
		t.Errorf("GetLocalStackRegion() = %q, want 'ap-south-1'", config.GetLocalStackRegion())
	}

	// Test with nil values - should return defaults
	configNil := &GlobalConfig{
		Version:  "1",
		Services: make(map[string]ServiceEntry),
	}

	if configNil.GetDockerComposeCommand() != "docker compose" {
		t.Errorf("GetDockerComposeCommand() = %q, want 'docker compose'", configNil.GetDockerComposeCommand())
	}
	if configNil.GetLocalStackEndpoint() != "http://localhost:4566" {
		t.Errorf("GetLocalStackEndpoint() = %q, want 'http://localhost:4566'", configNil.GetLocalStackEndpoint())
	}
	if configNil.GetLocalStackRegion() != "us-east-1" {
		t.Errorf("GetLocalStackRegion() = %q, want 'us-east-1'", configNil.GetLocalStackRegion())
	}
}

func TestGlobalConfig_AddService(t *testing.T) {
	config := DefaultGlobalConfig()

	config.AddService("test-service", ServiceEntry{
		Path: "~/projects/test-service",
		Repo: "git@github.com:org/test-service.git",
	})

	if len(config.Services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(config.Services))
	}

	entry, ok := config.Services["test-service"]
	if !ok {
		t.Fatal("Expected test-service to be present")
	}
	if entry.Path != "~/projects/test-service" {
		t.Errorf("Path = %q, want '~/projects/test-service'", entry.Path)
	}
}

func TestGlobalConfig_GetService(t *testing.T) {
	config := DefaultGlobalConfig()
	config.Services["test-service"] = ServiceEntry{
		Path: "~/projects/test-service",
	}

	// Test getting existing service
	entry, ok := config.GetService("test-service")
	if !ok {
		t.Error("Expected to find test-service")
	}
	if entry.Path != "~/projects/test-service" {
		t.Errorf("Path = %q, want '~/projects/test-service'", entry.Path)
	}

	// Test getting non-existing service
	_, ok = config.GetService("non-existing")
	if ok {
		t.Error("Expected not to find non-existing service")
	}
}

func TestGlobalConfigExists(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv(EnvGrundHome, tmpDir)
	defer os.Unsetenv(EnvGrundHome)

	// Should not exist initially
	if GlobalConfigExists() {
		t.Error("Config should not exist initially")
	}

	// Create config
	if err := InitGlobalConfig(); err != nil {
		t.Fatalf("InitGlobalConfig() error: %v", err)
	}

	// Should exist now
	if !GlobalConfigExists() {
		t.Error("Config should exist after init")
	}
}

func TestForceInitGlobalConfig(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv(EnvGrundHome, tmpDir)
	defer os.Unsetenv(EnvGrundHome)

	// Create config with services
	config := &GlobalConfig{
		Version: "1",
		Services: map[string]ServiceEntry{
			"existing-service": {Path: "~/projects/existing"},
		},
	}
	if err := SaveGlobalConfig(config); err != nil {
		t.Fatalf("SaveGlobalConfig() error: %v", err)
	}

	// Force re-init should overwrite
	if err := ForceInitGlobalConfig(); err != nil {
		t.Fatalf("ForceInitGlobalConfig() error: %v", err)
	}

	// Load and verify it was reset
	loaded, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("LoadGlobalConfig() error: %v", err)
	}
	if len(loaded.Services) != 0 {
		t.Errorf("Expected empty services after force init, got %d", len(loaded.Services))
	}
}
