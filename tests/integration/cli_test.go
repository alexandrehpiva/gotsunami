package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLIVersion(t *testing.T) {
	// Build the binary first
	cmd := exec.Command("go", "build", "-o", "gotsunami-test", "./cmd/gotsunami")
	err := cmd.Run()
	require.NoError(t, err)
	defer os.Remove("gotsunami-test")

	// Test version command
	cmd = exec.Command("./gotsunami-test", "version")
	output, err := cmd.Output()
	require.NoError(t, err)

	outputStr := string(output)
	assert.Contains(t, outputStr, "GoTsunami")
	assert.Contains(t, outputStr, "Build Time")
	assert.Contains(t, outputStr, "Go Version")
}

func TestCLIHelp(t *testing.T) {
	// Build the binary first
	cmd := exec.Command("go", "build", "-o", "gotsunami-test", "./cmd/gotsunami")
	err := cmd.Run()
	require.NoError(t, err)
	defer os.Remove("gotsunami-test")

	// Test help command
	cmd = exec.Command("./gotsunami-test", "help")
	output, err := cmd.Output()
	require.NoError(t, err)

	outputStr := string(output)
	assert.Contains(t, outputStr, "GoTsunami")
	assert.Contains(t, outputStr, "run")
	assert.Contains(t, outputStr, "validate")
	assert.Contains(t, outputStr, "version")
}

func TestCLIValidateScenario(t *testing.T) {
	// Build the binary first
	cmd := exec.Command("go", "build", "-o", "gotsunami-test", "./cmd/gotsunami")
	err := cmd.Run()
	require.NoError(t, err)
	defer os.Remove("gotsunami-test")

	// Test validate command with example scenario
	scenarioPath := filepath.Join("..", "..", "examples", "scenarios", "basic_get.json")
	cmd = exec.Command("./gotsunami-test", "validate", scenarioPath)
	output, err := cmd.Output()
	require.NoError(t, err)

	outputStr := string(output)
	assert.Contains(t, outputStr, "Validating scenario file")
	assert.Contains(t, outputStr, "JSON syntax is valid")
	assert.Contains(t, outputStr, "Scenario is ready for execution")
}

func TestCLIValidateNonExistentScenario(t *testing.T) {
	// Build the binary first
	cmd := exec.Command("go", "build", "-o", "gotsunami-test", "./cmd/gotsunami")
	err := cmd.Run()
	require.NoError(t, err)
	defer os.Remove("gotsunami-test")

	// Test validate command with non-existent scenario
	cmd = exec.Command("./gotsunami-test", "validate", "non-existent.json")
	err = cmd.Run()
	assert.Error(t, err)
}

func TestCLIRunWithInvalidScenario(t *testing.T) {
	// Build the binary first
	cmd := exec.Command("go", "build", "-o", "gotsunami-test", "./cmd/gotsunami")
	err := cmd.Run()
	require.NoError(t, err)
	defer os.Remove("gotsunami-test")

	// Test run command with non-existent scenario
	cmd = exec.Command("./gotsunami-test", "run", "non-existent.json")
	err = cmd.Run()
	assert.Error(t, err)
}

func TestCLIRunWithValidScenario(t *testing.T) {
	// Build the binary first
	cmd := exec.Command("go", "build", "-o", "gotsunami-test", "./cmd/gotsunami")
	err := cmd.Run()
	require.NoError(t, err)
	defer os.Remove("gotsunami-test")

	// Test run command with valid scenario (short duration for testing)
	scenarioPath := filepath.Join("..", "..", "examples", "scenarios", "basic_get.json")
	cmd = exec.Command("./gotsunami-test", "run", scenarioPath, "--vus", "1", "--duration", "1s", "--quiet")
	output, err := cmd.Output()

	// The command might fail due to network issues, but it should not fail due to CLI issues
	outputStr := string(output)
	if err != nil {
		// If it fails, it should be due to network/HTTP issues, not CLI issues
		assert.Contains(t, err.Error(), "load test failed")
	} else {
		// If it succeeds, it should produce some output
		assert.NotEmpty(t, outputStr)
	}
}
