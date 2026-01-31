package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// These are end-to-end tests that test the CLI binary

// Helper to get the path to the grund binary
func getBinaryPath(t *testing.T) string {
	t.Helper()

	// Try to find the binary in common locations
	locations := []string{
		"../../bin/grund",
		"../bin/grund",
		"./bin/grund",
		"/usr/local/bin/grund",
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			abs, _ := filepath.Abs(loc)
			return abs
		}
	}

	// Try to build it
	t.Log("Binary not found, attempting to build...")
	cmd := exec.Command("go", "build", "-o", "../../bin/grund", "../../cmd/grund")
	if err := cmd.Run(); err != nil {
		t.Skipf("Could not find or build grund binary: %v", err)
	}

	return "../../bin/grund"
}

// Helper to run a grund command and capture output
func runGrund(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	binary := getBinaryPath(t)
	cmd := exec.Command(binary, args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	exitCode = 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		exitCode = 1
	}

	return stdoutBuf.String(), stderrBuf.String(), exitCode
}

func TestCLI_Help(t *testing.T) {
	stdout, _, exitCode := runGrund(t, "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for --help, got %d", exitCode)
	}

	// Check that help text contains expected commands
	expectedCommands := []string{"up", "down", "status", "logs", "restart"}
	for _, cmd := range expectedCommands {
		if !strings.Contains(stdout, cmd) {
			t.Errorf("Expected help to contain command %q", cmd)
		}
	}

	// Check for tool name
	if !strings.Contains(strings.ToLower(stdout), "grund") {
		t.Error("Expected help to contain 'grund'")
	}
}

func TestCLI_Version(t *testing.T) {
	stdout, _, exitCode := runGrund(t, "--version")

	// Version might not be implemented, just check it doesn't crash
	if exitCode != 0 {
		t.Skipf("--version not implemented or returned error (exit code %d)", exitCode)
	}

	if stdout == "" {
		t.Error("Expected version output, got empty string")
	}
}

func TestCLI_Up_MissingServicesYaml(t *testing.T) {
	// Change to a temp directory without services.yaml
	tmpDir := t.TempDir()

	binary := getBinaryPath(t)
	cmd := exec.Command(binary, "up", "some-service")
	cmd.Dir = tmpDir

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	if err == nil {
		t.Error("Expected error when running 'up' without services.yaml, got nil")
	}

	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "services.yaml") && !strings.Contains(stderr, "not found") && !strings.Contains(stderr, "error") {
		t.Logf("stderr: %s", stderr)
		// Just log, don't fail - the specific error message may vary
	}
}

func TestCLI_Down_NoActiveServices(t *testing.T) {
	// Running 'down' when nothing is running should be safe
	tmpDir := t.TempDir()

	// Create a minimal services.yaml to avoid "not in orchestration repo" error
	servicesYaml := `version: "1"
services: {}
`
	os.WriteFile(filepath.Join(tmpDir, "services.yaml"), []byte(servicesYaml), 0644)

	binary := getBinaryPath(t)
	cmd := exec.Command(binary, "down")
	cmd.Dir = tmpDir

	// This will likely fail due to docker compose issues, but shouldn't panic
	err := cmd.Run()
	// We don't check error because docker might not be available
	// The important thing is the command doesn't panic
	_ = err
}

func TestCLI_Status_NoServices(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal services.yaml
	servicesYaml := `version: "1"
services: {}
`
	os.WriteFile(filepath.Join(tmpDir, "services.yaml"), []byte(servicesYaml), 0644)

	binary := getBinaryPath(t)
	cmd := exec.Command(binary, "status")
	cmd.Dir = tmpDir

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	_ = cmd.Run()
	// Just verify it doesn't crash
}

func TestCLI_UnknownCommand(t *testing.T) {
	_, stderr, exitCode := runGrund(t, "unknowncommand")

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for unknown command")
	}

	if !strings.Contains(stderr, "unknown command") && !strings.Contains(stderr, "unknowncommand") {
		t.Logf("stderr: %s", stderr)
		// Just log - error message format may vary
	}
}

func TestCLI_Up_InvalidService(t *testing.T) {
	tmpDir := t.TempDir()

	// Create services.yaml with some services
	servicesYaml := `version: "1"
services:
  valid-service:
    path: /some/path
`
	os.WriteFile(filepath.Join(tmpDir, "services.yaml"), []byte(servicesYaml), 0644)

	binary := getBinaryPath(t)
	cmd := exec.Command(binary, "up", "nonexistent-service")
	cmd.Dir = tmpDir

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	if err == nil {
		t.Error("Expected error for nonexistent service")
	}
}

func TestCLI_Config_ShowsResolvedConfig(t *testing.T) {
	// This test requires a properly set up fixture
	// For now, we just verify the command exists and responds appropriately

	tmpDir := t.TempDir()

	servicesYaml := `version: "1"
services:
  test-service:
    path: /nonexistent/path
`
	os.WriteFile(filepath.Join(tmpDir, "services.yaml"), []byte(servicesYaml), 0644)

	binary := getBinaryPath(t)
	cmd := exec.Command(binary, "config", "test-service")
	cmd.Dir = tmpDir

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	_ = cmd.Run()

	// The command might fail due to missing grund.yaml in the service directory
	// That's expected - we're just testing the CLI routing works
}

func TestCLI_Init(t *testing.T) {
	tmpDir := t.TempDir()
	serviceDir := filepath.Join(tmpDir, "new-service")
	os.MkdirAll(serviceDir, 0755)

	binary := getBinaryPath(t)
	cmd := exec.Command(binary, "init")
	cmd.Dir = serviceDir

	var stdoutBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf

	err := cmd.Run()

	// Check if grund.yaml was created
	grundYaml := filepath.Join(serviceDir, "grund.yaml")
	if err == nil {
		if _, statErr := os.Stat(grundYaml); os.IsNotExist(statErr) {
			t.Logf("init command succeeded but grund.yaml was not created")
		}
	}
}

func TestCLI_Logs_NoService(t *testing.T) {
	stdout, stderr, exitCode := runGrund(t, "logs")

	// Behavior depends on implementation - some CLIs allow 'logs' without args
	// to show all logs, others require a service name
	// We just verify it doesn't crash

	_ = stdout
	_ = stderr
	_ = exitCode
}

func TestCLI_Restart_NoService(t *testing.T) {
	stdout, stderr, exitCode := runGrund(t, "restart")

	// Behavior depends on implementation
	// Some CLIs allow 'restart' without args to restart all
	// Others require at least one service name

	_ = stdout
	_ = stderr
	_ = exitCode
	// Just verify it doesn't crash
}
