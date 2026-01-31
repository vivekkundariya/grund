package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTestServicesYaml(t *testing.T, dir string) string {
	t.Helper()

	content := `version: "1"

services:
  service-a:
    repo: git@github.com:org/service-a.git
    path: /path/to/service-a
  service-b:
    repo: git@github.com:org/service-b.git
    path: ~/projects/service-b

path_defaults:
  base: /default/path
`

	filePath := filepath.Join(dir, "services.yaml")
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write services.yaml: %v", err)
	}

	return filePath
}

func TestServiceRegistryRepository_GetServicePath(t *testing.T) {
	tmpDir := t.TempDir()
	servicesPath := createTestServicesYaml(t, tmpDir)

	repo, err := NewServiceRegistryRepository(servicesPath)
	if err != nil {
		t.Fatalf("NewServiceRegistryRepository() returned error: %v", err)
	}

	path, err := repo.GetServicePath("service-a")
	if err != nil {
		t.Fatalf("GetServicePath() returned error: %v", err)
	}

	if path != "/path/to/service-a" {
		t.Errorf("Expected path '/path/to/service-a', got %q", path)
	}
}

func TestServiceRegistryRepository_PathExpansion(t *testing.T) {
	tmpDir := t.TempDir()
	servicesPath := createTestServicesYaml(t, tmpDir)

	repo, err := NewServiceRegistryRepository(servicesPath)
	if err != nil {
		t.Fatalf("NewServiceRegistryRepository() returned error: %v", err)
	}

	path, err := repo.GetServicePath("service-b")
	if err != nil {
		t.Fatalf("GetServicePath() returned error: %v", err)
	}

	// Path should have ~ expanded to home directory
	home, _ := os.UserHomeDir()
	expectedPath := filepath.Join(home, "projects/service-b")

	if path != expectedPath {
		t.Errorf("Expected path %q, got %q", expectedPath, path)
	}

	// Should not start with ~
	if strings.HasPrefix(path, "~") {
		t.Errorf("Path should have ~ expanded, got %q", path)
	}
}

func TestServiceRegistryRepository_GetServicePath_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	servicesPath := createTestServicesYaml(t, tmpDir)

	repo, err := NewServiceRegistryRepository(servicesPath)
	if err != nil {
		t.Fatalf("NewServiceRegistryRepository() returned error: %v", err)
	}

	_, err = repo.GetServicePath("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent service, got nil")
	}
}

func TestServiceRegistryRepository_GetAllServices(t *testing.T) {
	tmpDir := t.TempDir()
	servicesPath := createTestServicesYaml(t, tmpDir)

	repo, err := NewServiceRegistryRepository(servicesPath)
	if err != nil {
		t.Fatalf("NewServiceRegistryRepository() returned error: %v", err)
	}

	services, err := repo.GetAllServices()
	if err != nil {
		t.Fatalf("GetAllServices() returned error: %v", err)
	}

	if len(services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(services))
	}

	// Check service-a
	entryA, ok := services["service-a"]
	if !ok {
		t.Error("Expected service-a in registry")
	} else {
		if entryA.Repo != "git@github.com:org/service-a.git" {
			t.Errorf("Expected repo 'git@github.com:org/service-a.git', got %q", entryA.Repo)
		}
		if entryA.Path != "/path/to/service-a" {
			t.Errorf("Expected path '/path/to/service-a', got %q", entryA.Path)
		}
	}

	// Check service-b
	_, ok = services["service-b"]
	if !ok {
		t.Error("Expected service-b in registry")
	}
}

func TestServiceRegistryRepository_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Write invalid YAML
	invalidPath := filepath.Join(tmpDir, "services.yaml")
	err := os.WriteFile(invalidPath, []byte("invalid: yaml: content:"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = NewServiceRegistryRepository(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestServiceRegistryRepository_FileNotFound(t *testing.T) {
	_, err := NewServiceRegistryRepository("/nonexistent/services.yaml")
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
}

func TestServiceRegistryRepository_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Write empty YAML
	emptyPath := filepath.Join(tmpDir, "services.yaml")
	err := os.WriteFile(emptyPath, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	repo, err := NewServiceRegistryRepository(emptyPath)
	if err != nil {
		t.Fatalf("NewServiceRegistryRepository() returned error: %v", err)
	}

	services, err := repo.GetAllServices()
	if err != nil {
		t.Fatalf("GetAllServices() returned error: %v", err)
	}

	if len(services) != 0 {
		t.Errorf("Expected 0 services for empty file, got %d", len(services))
	}
}
