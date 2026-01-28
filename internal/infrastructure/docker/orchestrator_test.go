package docker

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/vivekkundariya/grund/internal/config"
)

func TestNewDockerOrchestrator(t *testing.T) {
	orchestrator := NewDockerOrchestrator("/path/to/docker-compose.yaml", "/working/dir")
	
	if orchestrator == nil {
		t.Fatal("NewDockerOrchestrator() returned nil")
	}
	
	// Verify it implements the interface by type assertion
	_, ok := orchestrator.(*DockerOrchestrator)
	if !ok {
		t.Error("NewDockerOrchestrator() should return *DockerOrchestrator")
	}
}

func TestGetComposeFilePath(t *testing.T) {
	// GetComposeFilePath should return path in ~/.grund/tmp/<project-name>/
	grundHome, err := config.GetGrundHome()
	if err != nil {
		t.Fatalf("Failed to get grund home: %v", err)
	}

	tests := []struct {
		name              string
		orchestrationRoot string
		expectedProject   string
	}{
		{
			name:              "root directory",
			orchestrationRoot: "/project",
			expectedProject:   "project",
		},
		{
			name:              "nested directory",
			orchestrationRoot: "/home/user/projects/myapp",
			expectedProject:   "myapp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetComposeFilePath(tt.orchestrationRoot)
			expectedPath := filepath.Join(grundHome, "tmp", tt.expectedProject, "docker-compose.generated.yaml")

			if result != expectedPath {
				t.Errorf("GetComposeFilePath(%q) = %q, want %q", tt.orchestrationRoot, result, expectedPath)
			}
			// Verify it's in the tmp directory with project name
			if !strings.Contains(result, filepath.Join("tmp", tt.expectedProject)) {
				t.Errorf("GetComposeFilePath should return path in tmp/<project>, got %q", result)
			}
		})
	}
}

func TestDockerOrchestrator_CommandConstruction(t *testing.T) {
	// Test that the orchestrator is constructed with correct paths
	composeFile := "/test/docker-compose.yaml"
	workingDir := "/test/workdir"
	
	orchestrator := NewDockerOrchestrator(composeFile, workingDir).(*DockerOrchestrator)
	
	if orchestrator.composeFile != composeFile {
		t.Errorf("Expected composeFile %q, got %q", composeFile, orchestrator.composeFile)
	}
	if orchestrator.workingDir != workingDir {
		t.Errorf("Expected workingDir %q, got %q", workingDir, orchestrator.workingDir)
	}
}

// Note: Integration tests for StartServices, StopServices, RestartService
// would require Docker to be available. These are better suited for
// end-to-end tests or tests with mocked exec.Command.
// Below we document what those tests would verify:

// TestDockerOrchestrator_StartServices_Integration would verify:
// - Correct docker compose command is executed
// - Services are passed as arguments
// - -d flag is used for detached mode
// - Errors from docker are propagated

// TestDockerOrchestrator_StopServices_Integration would verify:
// - docker compose down is executed
// - Correct compose file is specified
// - Working directory is correct

// TestDockerOrchestrator_RestartService_Integration would verify:
// - docker compose restart is executed with service name
// - Errors are handled appropriately
