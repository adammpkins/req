package tests

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/adammpkins/req/internal/parser"
	"github.com/adammpkins/req/internal/planner"
	"github.com/adammpkins/req/internal/runtime"
)

// TestSaveStreaming tests that save verb uses efficient file writing.
// Note: Current implementation uses io.Copy which is efficient, but reads body into memory first.
// Future optimization: stream directly from resp.Body when no expect checks are present.
func TestSaveStreaming(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	// Create a large response endpoint
	ts.mux.HandleFunc("/large", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		// Write 1MB of data
		data := make([]byte, 1024*1024)
		for i := range data {
			data[i] = byte(i % 256)
		}
		w.Write(data)
	})

	tmpFile := filepath.Join(t.TempDir(), "output.bin")
	cmdStr := "save " + ts.URL() + "/large to=" + tmpFile
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

	// Verify file was created and has correct size
	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}

	expectedSize := int64(1024 * 1024)
	if info.Size() != expectedSize {
		t.Errorf("Expected file size %d, got %d", expectedSize, info.Size())
	}
}

// TestWatchTTYDetection tests TTY detection for watch verb.
// Note: Watch verb TTY detection is not yet implemented.
func TestWatchTTYDetection(t *testing.T) {
	t.Skip("Watch verb TTY detection not yet implemented")
}

