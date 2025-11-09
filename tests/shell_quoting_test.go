package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestShellQuotingBash tests bash quoting behavior.
func TestShellQuotingBash(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
		return
	}

	// Test command with include clause containing commas and semicolons
	testCmd := `req read https://api.example.com/search include='header: Accept: application/json, application/problem+json; param: q=test;value; cookie: session=abc' as=json`

	// Create a test script that prints argv
	script := `#!/bin/bash
exec "$@" --argv-test
`
	tmpScript := filepath.Join(t.TempDir(), "test.sh")
	if err := os.WriteFile(tmpScript, []byte(script), 0755); err != nil {
		t.Fatalf("Failed to write test script: %v", err)
	}

	// Run through bash and capture output
	cmd := exec.Command("bash", "-c", testCmd+" --argv-test")
	output, _ := cmd.CombinedOutput()

	// Load golden file
	goldenPath := filepath.Join("fixtures", "shell_quoting_bash.golden")
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			os.WriteFile(goldenPath, output, 0644)
			t.Logf("Created golden file: %s", goldenPath)
			return
		}
		t.Fatalf("Failed to read golden file: %v", err)
	}

	// Compare
	actualNorm := strings.ReplaceAll(string(output), "\r\n", "\n")
	expectedNorm := strings.ReplaceAll(string(expected), "\r\n", "\n")

	if actualNorm != expectedNorm {
		t.Errorf("Bash quoting mismatch:\nExpected:\n%s\nGot:\n%s", expectedNorm, actualNorm)
	}
}

// TestShellQuotingZsh tests zsh quoting behavior.
func TestShellQuotingZsh(t *testing.T) {
	if _, err := exec.LookPath("zsh"); err != nil {
		t.Skip("zsh not available")
		return
	}

	testCmd := `req read https://api.example.com/search include='header: Accept: application/json, application/problem+json; param: q=test;value; cookie: session=abc' as=json`

	cmd := exec.Command("zsh", "-c", testCmd+" --argv-test")
	output, _ := cmd.CombinedOutput()

	goldenPath := filepath.Join("fixtures", "shell_quoting_zsh.golden")
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			os.WriteFile(goldenPath, output, 0644)
			t.Logf("Created golden file: %s", goldenPath)
			return
		}
		t.Fatalf("Failed to read golden file: %v", err)
	}

	actualNorm := strings.ReplaceAll(string(output), "\r\n", "\n")
	expectedNorm := strings.ReplaceAll(string(expected), "\r\n", "\n")

	if actualNorm != expectedNorm {
		t.Errorf("Zsh quoting mismatch:\nExpected:\n%s\nGot:\n%s", expectedNorm, actualNorm)
	}
}

// TestShellQuotingFish tests fish shell quoting behavior.
func TestShellQuotingFish(t *testing.T) {
	if _, err := exec.LookPath("fish"); err != nil {
		t.Skip("fish not available")
		return
	}

	testCmd := `req read https://api.example.com/search include='header: Accept: application/json, application/problem+json; param: q=test;value; cookie: session=abc' as=json`

	cmd := exec.Command("fish", "-c", testCmd+" --argv-test")
	output, _ := cmd.CombinedOutput()

	goldenPath := filepath.Join("fixtures", "shell_quoting_fish.golden")
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			os.WriteFile(goldenPath, output, 0644)
			t.Logf("Created golden file: %s", goldenPath)
			return
		}
		t.Fatalf("Failed to read golden file: %v", err)
	}

	actualNorm := strings.ReplaceAll(string(output), "\r\n", "\n")
	expectedNorm := strings.ReplaceAll(string(expected), "\r\n", "\n")

	if actualNorm != expectedNorm {
		t.Errorf("Fish quoting mismatch:\nExpected:\n%s\nGot:\n%s", expectedNorm, actualNorm)
	}
}

// TestShellQuotingPowerShell tests PowerShell quoting behavior.
func TestShellQuotingPowerShell(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		if _, err := exec.LookPath("powershell"); err != nil {
			t.Skip("PowerShell not available")
			return
		}
	}

	testCmd := `req read https://api.example.com/search include='header: Accept: application/json, application/problem+json; param: q=test;value; cookie: session=abc' as=json`

	var cmd *exec.Cmd
	if _, err := exec.LookPath("pwsh"); err == nil {
		cmd = exec.Command("pwsh", "-Command", testCmd+" --argv-test")
	} else {
		cmd = exec.Command("powershell", "-Command", testCmd+" --argv-test")
	}
	output, _ := cmd.CombinedOutput()

	goldenPath := filepath.Join("fixtures", "shell_quoting_powershell.golden")
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			os.WriteFile(goldenPath, output, 0644)
			t.Logf("Created golden file: %s", goldenPath)
			return
		}
		t.Fatalf("Failed to read golden file: %v", err)
	}

	actualNorm := strings.ReplaceAll(string(output), "\r\n", "\n")
	expectedNorm := strings.ReplaceAll(string(expected), "\r\n", "\n")

	if actualNorm != expectedNorm {
		t.Errorf("PowerShell quoting mismatch:\nExpected:\n%s\nGot:\n%s", expectedNorm, actualNorm)
	}
}

// Note: These tests require the binary to support --argv-test flag that prints argv.
// For now, they will skip if shells are not available, which is acceptable.
// The golden files can be created manually by running the commands and capturing output.

