package tests

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/adammpkins/req/internal/parser"
	"github.com/adammpkins/req/internal/planner"
	"github.com/adammpkins/req/internal/runtime"
)

// TestRuntimeReadDogFacts tests reading from a real API.
// Note: This test requires internet connectivity.
func TestRuntimeReadDogFacts(t *testing.T) {
	// Use httpbin.org as a reliable test API
	// If dog-facts-api comes back online, we can switch back
	cmdStr := "read https://httpbin.org/json"
	
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

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = executor.Execute(plan)
	
	// Close write end and restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Read captured output
	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	// Verify output is JSON
	var data map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, string(output))
	}

	// Verify we got some data
	if len(data) == 0 {
		t.Errorf("Expected non-empty JSON response, got empty map")
	}

	t.Logf("Successfully received JSON response: %v", data)
}

// TestRuntimeReadMultipleDogFacts tests reading from a real API with query params.
// Note: This test requires internet connectivity.
func TestRuntimeReadMultipleDogFacts(t *testing.T) {
	// Use httpbin.org with query params
	cmdStr := "read https://httpbin.org/get?foo=bar&baz=qux"
	
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

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = executor.Execute(plan)
	
	// Close write end and restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Read captured output
	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	// Verify output is JSON
	var data map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, string(output))
	}

	// Verify we got data with query params
	if args, ok := data["args"].(map[string]interface{}); ok {
		if args["foo"] != "bar" || args["baz"] != "qux" {
			t.Errorf("Expected query params foo=bar&baz=qux, got: %v", args)
		}
	} else {
		t.Errorf("Expected 'args' field in response, got: %v", data)
	}

	t.Logf("Successfully received JSON response with query params")
}

// TestRuntimeSaveDogFact tests saving API response to a file.
// Note: This test requires internet connectivity.
func TestRuntimeSaveDogFact(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "api-response-*.json")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmdStr := "read https://httpbin.org/json to=" + tmpFile.Name()
	
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

	err = executor.Execute(plan)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify file was created and contains valid JSON
	fileContent, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(fileContent, &data); err != nil {
		t.Fatalf("File content is not valid JSON: %v\nContent: %s", err, string(fileContent))
	}

	if len(data) == 0 {
		t.Errorf("Expected non-empty JSON in file, got empty map")
	}

	t.Logf("Successfully saved API response to file: %s", tmpFile.Name())
}

// TestRuntimeSaveVideoFile tests downloading a video file using the save command.
// Note: This test requires internet connectivity.
func TestRuntimeSaveVideoFile(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "video-*.mp4")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Use the provided test video URL
	cmdStr := "save http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4 to=" + tmpFile.Name()
	
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

	err = executor.Execute(plan)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify file was created and has content
	fileInfo, err := os.Stat(tmpFile.Name())
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	if fileInfo.Size() == 0 {
		t.Error("Expected non-empty file, got empty file")
	}

	// Video files should be reasonably large (at least a few KB)
	if fileInfo.Size() < 1024 {
		t.Errorf("Expected file size >= 1KB, got %d bytes", fileInfo.Size())
	}

	// Verify file content is binary (not text)
	fileContent, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// Check for MP4 file signature (ftyp box at the beginning)
	// MP4 files typically start with some metadata, but we can check for non-text content
	if len(fileContent) == 0 {
		t.Error("Expected non-empty file content")
	}

	// MP4 files typically start with specific bytes or contain binary data
	// For a simple test, verify it's not plain text by checking for null bytes
	// or non-printable characters in the first few bytes
	hasBinaryContent := false
	for i := 0; i < len(fileContent) && i < 100; i++ {
		if fileContent[i] == 0 || (fileContent[i] < 32 && fileContent[i] != 9 && fileContent[i] != 10 && fileContent[i] != 13) {
			hasBinaryContent = true
			break
		}
	}

	if !hasBinaryContent && len(fileContent) > 100 {
		// If first 100 bytes are all printable, it's likely not a binary video file
		t.Logf("Warning: File content appears to be text, expected binary video file")
	}

	t.Logf("Successfully downloaded video file: %s (size: %d bytes)", tmpFile.Name(), fileInfo.Size())
}

// TestRuntimeReadWithTimeout tests request with timeout.
// Note: This test requires internet connectivity.
func TestRuntimeReadWithTimeout(t *testing.T) {
	cmdStr := "read https://httpbin.org/json under=10s"
	
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

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = executor.Execute(plan)
	
	// Close write end and restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify we got output
	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output")
	}

	t.Logf("Successfully received response with timeout")
}

// TestRuntimeHTTPError tests handling of HTTP errors.
// Note: This test requires internet connectivity.
func TestRuntimeHTTPError(t *testing.T) {
	// Use a non-existent endpoint to trigger 404
	cmdStr := "read https://httpbin.org/status/404"
	
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

	err = executor.Execute(plan)
	
	// Should return an error for non-2xx status
	if err == nil {
		t.Error("Expected error for non-existent endpoint, got nil")
	}

	// Verify error message contains status code
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("Expected error to contain '404', got: %v", err)
	}

	t.Logf("Successfully handled HTTP error: %v", err)
}

// TestRuntimeInvalidURL tests handling of invalid URLs.
func TestRuntimeInvalidURL(t *testing.T) {
	cmdStr := "read https://this-domain-does-not-exist-12345.example.com/nonexistent"
	
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

	err = executor.Execute(plan)
	
	// Should return an error for invalid domain
	if err == nil {
		t.Error("Expected error for invalid domain, got nil")
	}

	t.Logf("Successfully handled invalid URL error: %v", err)
}

// TestRuntimeBasicAuth tests Basic Auth with httpbin.org/basic-auth endpoint.
// Note: This test requires internet connectivity.
func TestRuntimeBasicAuth(t *testing.T) {
	cmdStr := `read https://httpbin.org/basic-auth/user/passwd include='basic: user:passwd' expect=status:200`
	
	cmd, err := parser.Parse(cmdStr)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	plan, err := planner.Plan(cmd)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	// Verify that Authorization header was set correctly
	authHeader, ok := plan.Headers["Authorization"]
	if !ok {
		t.Fatal("Authorization header not set in plan")
	}
	if !strings.HasPrefix(authHeader, "Basic ") {
		t.Errorf("Authorization header should start with 'Basic ', got: %s", authHeader)
	}

	executor, err := runtime.NewExecutor(plan)
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = executor.Execute(plan)
	
	// Close write end and restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Read output
	output, _ := io.ReadAll(r)
	
	// Verify we got a successful response
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	
	authenticated, ok := result["authenticated"].(bool)
	if !ok || !authenticated {
		t.Errorf("Expected authenticated=true, got: %v", result["authenticated"])
	}
	
	user, ok := result["user"].(string)
	if !ok || user != "user" {
		t.Errorf("Expected user='user', got: %v", result["user"])
	}
}

