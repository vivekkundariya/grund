package docker

import (
	"testing"
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
	tests := []struct {
		name            string
		orchestrationRoot string
		expected        string
	}{
		{
			name:              "root directory",
			orchestrationRoot: "/project",
			expected:          "/project/docker-compose.generated.yaml",
		},
		{
			name:              "nested directory",
			orchestrationRoot: "/home/user/projects/myapp",
			expected:          "/home/user/projects/myapp/docker-compose.generated.yaml",
		},
		{
			name:              "relative path",
			orchestrationRoot: ".",
			expected:          "docker-compose.generated.yaml",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetComposeFilePath(tt.orchestrationRoot)
			if result != tt.expected {
				t.Errorf("GetComposeFilePath(%q) = %q, want %q", tt.orchestrationRoot, result, tt.expected)
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
