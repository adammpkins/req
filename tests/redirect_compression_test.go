package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adammpkins/req/internal/parser"
	"github.com/adammpkins/req/internal/planner"
	"github.com/adammpkins/req/internal/runtime"
)

// TestRedirectTrace tests that redirect traces appear in stderr.
func TestRedirectTrace(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	// Test redirect with read verb (should follow by default)
	cmdStr := "read " + ts.URL() + "/redirect/301"
	cmd, err := parser.Parse(cmdStr)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	plan, err := planner.Plan(cmd)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	executor, err := runtime.NewExecutor(plan)
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err = executor.Execute(plan)
	os.Stderr = oldStderr
	w.Close()

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Read stderr
	var buf bytes.Buffer
	buf.ReadFrom(r)
	stderr := buf.String()

	// Check for redirect trace
	if !strings.Contains(stderr, "→") {
		t.Errorf("Expected redirect trace in stderr, got: %s", stderr)
	}
	if !strings.Contains(stderr, "301") {
		t.Errorf("Expected status code 301 in redirect trace, got: %s", stderr)
	}
}

// TestCompressionTrace tests that decompression notes appear in stderr.
func TestCompressionTrace(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	cmdStr := "read " + ts.URL() + "/gzip"
	cmd, err := parser.Parse(cmdStr)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	plan, err := planner.Plan(cmd)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	executor, err := runtime.NewExecutor(plan)
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err = executor.Execute(plan)
	os.Stderr = oldStderr
	w.Close()

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Read stderr
	var buf bytes.Buffer
	buf.ReadFrom(r)
	stderr := buf.String()

	// Check for decompression note
	if !strings.Contains(stderr, "Decompressed") {
		t.Errorf("Expected decompression note in stderr, got: %s", stderr)
	}
}

// TestRedirectTraceGolden tests redirect trace output against golden file.
func TestRedirectTraceGolden(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	cmdStr := "read " + ts.URL() + "/redirect/307"
	cmd, err := parser.Parse(cmdStr)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	plan, err := planner.Plan(cmd)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	executor, err := runtime.NewExecutor(plan)
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	var stderrBuf bytes.Buffer
	done := make(chan bool)
	go func() {
		stderrBuf.ReadFrom(r)
		done <- true
	}()

	err = executor.Execute(plan)
	w.Close()
	os.Stderr = oldStderr
	<-done

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	stderr := stderrBuf.String()

	// Load golden file
	goldenPath := filepath.Join("fixtures", "redirect_trace.golden")
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		// Create golden file if it doesn't exist
		if os.IsNotExist(err) {
			os.WriteFile(goldenPath, []byte(stderr), 0644)
			t.Logf("Created golden file: %s", goldenPath)
			return
		}
		t.Fatalf("Failed to read golden file: %v", err)
	}

	// Compare (normalize line endings)
	actualNorm := strings.ReplaceAll(stderr, "\r\n", "\n")
	expectedNorm := strings.ReplaceAll(string(expected), "\r\n", "\n")

	if actualNorm != expectedNorm {
		t.Errorf("Redirect trace mismatch:\nExpected:\n%s\nGot:\n%s", expectedNorm, actualNorm)
	}
}

// TestCompressionTraceGolden tests compression trace output against golden file.
func TestCompressionTraceGolden(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	cmdStr := "read " + ts.URL() + "/gzip"
	cmd, err := parser.Parse(cmdStr)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	plan, err := planner.Plan(cmd)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	executor, err := runtime.NewExecutor(plan)
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	var stderrBuf bytes.Buffer
	done := make(chan bool)
	go func() {
		stderrBuf.ReadFrom(r)
		done <- true
	}()

	err = executor.Execute(plan)
	w.Close()
	os.Stderr = oldStderr
	<-done

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	stderr := stderrBuf.String()

	// Load golden file
	goldenPath := filepath.Join("fixtures", "compression_trace.golden")
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		// Create golden file if it doesn't exist
		if os.IsNotExist(err) {
			os.WriteFile(goldenPath, []byte(stderr), 0644)
			t.Logf("Created golden file: %s", goldenPath)
			return
		}
		t.Fatalf("Failed to read golden file: %v", err)
	}

	// Compare (normalize line endings)
	actualNorm := strings.ReplaceAll(stderr, "\r\n", "\n")
	expectedNorm := strings.ReplaceAll(string(expected), "\r\n", "\n")

	if actualNorm != expectedNorm {
		t.Errorf("Compression trace mismatch:\nExpected:\n%s\nGot:\n%s", expectedNorm, actualNorm)
	}
}

// TestBinaryRedirectTrace tests redirect trace using the built binary.
func TestBinaryRedirectTrace(t *testing.T) {
	binaryPath := filepath.Join("..", "bin", "req")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// Try to build it
		cmd := exec.Command("go", "build", "-o", binaryPath, "../cmd/req")
		if err := cmd.Run(); err != nil {
			t.Skipf("Could not build binary: %v", err)
			return
		}
	}

	ts := NewTestServer()
	defer ts.Close()

	// Run req with redirect
	cmd := exec.Command(binaryPath, "read", ts.URL()+"/redirect/302")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Run() // Don't care about exit code, just want stderr

	output := stderr.String()
	if !strings.Contains(output, "→") {
		t.Errorf("Expected redirect trace in binary output, got: %s", output)
	}
}

// TestBinaryCompressionTrace tests compression trace using the built binary.
func TestBinaryCompressionTrace(t *testing.T) {
	binaryPath := filepath.Join("..", "bin", "req")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// Try to build it
		cmd := exec.Command("go", "build", "-o", binaryPath, "../cmd/req")
		if err := cmd.Run(); err != nil {
			t.Skipf("Could not build binary: %v", err)
			return
		}
	}

	ts := NewTestServer()
	defer ts.Close()

	// Run req with gzip endpoint
	cmd := exec.Command(binaryPath, "read", ts.URL()+"/gzip")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Run() // Don't care about exit code, just want stderr

	output := stderr.String()
	if !strings.Contains(output, "Decompressed") {
		t.Errorf("Expected decompression note in binary output, got: %s", output)
	}
}

