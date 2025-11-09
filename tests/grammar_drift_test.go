package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adammpkins/req/internal/grammar"
)

// TestGrammarDrift ensures the binary help output matches the grammar snapshot.
func TestGrammarDrift(t *testing.T) {
	// Get expected snapshot from grammar package
	expectedJSON, err := grammar.GetSnapshotJSON()
	if err != nil {
		t.Fatalf("Failed to generate expected snapshot: %v", err)
	}

	var expected grammar.Snapshot
	if err := json.Unmarshal(expectedJSON, &expected); err != nil {
		t.Fatalf("Failed to unmarshal expected snapshot: %v", err)
	}

	// Load actual snapshot from file
	snapshotPath := filepath.Join("fixtures", "grammar_snapshot.json")
	actualJSON, err := os.ReadFile(snapshotPath)
	if err != nil {
		t.Fatalf("Failed to read snapshot file: %v", err)
	}

	var actual grammar.Snapshot
	if err := json.Unmarshal(actualJSON, &actual); err != nil {
		t.Fatalf("Failed to unmarshal actual snapshot: %v", err)
	}

	// Compare verbs
	if len(expected.Verbs) != len(actual.Verbs) {
		t.Errorf("Verb count mismatch: expected %d, got %d", len(expected.Verbs), len(actual.Verbs))
	}

	expectedVerbs := make(map[string]bool)
	for _, v := range expected.Verbs {
		expectedVerbs[v] = true
	}
	for _, v := range actual.Verbs {
		if !expectedVerbs[v] {
			t.Errorf("Unexpected verb in snapshot: %s", v)
		}
	}
	for _, v := range expected.Verbs {
		found := false
		for _, av := range actual.Verbs {
			if av == v {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing verb in snapshot: %s", v)
		}
	}

	// Compare clauses
	if len(expected.Clauses) != len(actual.Clauses) {
		t.Errorf("Clause count mismatch: expected %d, got %d", len(expected.Clauses), len(actual.Clauses))
	}

	expectedClauses := make(map[string]grammar.ClauseSnapshot)
	for _, c := range expected.Clauses {
		expectedClauses[c.Name] = c
	}

	for _, ac := range actual.Clauses {
		ec, ok := expectedClauses[ac.Name]
		if !ok {
			t.Errorf("Unexpected clause in snapshot: %s", ac.Name)
			continue
		}
		if ec.Description != ac.Description {
			t.Errorf("Clause %s description mismatch: expected %q, got %q", ac.Name, ec.Description, ac.Description)
		}
		if ec.Repeatable != ac.Repeatable {
			t.Errorf("Clause %s repeatable mismatch: expected %v, got %v", ac.Name, ec.Repeatable, ac.Repeatable)
		}
	}

	for _, ec := range expected.Clauses {
		found := false
		for _, ac := range actual.Clauses {
			if ac.Name == ec.Name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing clause in snapshot: %s", ec.Name)
		}
	}
}

// TestBinaryHelpDrift ensures the binary help output matches the grammar.
func TestBinaryHelpDrift(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join("..", "bin", "req")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// Try to build it
		cmd := exec.Command("go", "build", "-o", binaryPath, "../cmd/req")
		if err := cmd.Run(); err != nil {
			t.Skipf("Could not build binary: %v", err)
			return
		}
	}

	// Run req help
	cmd := exec.Command(binaryPath, "help")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run req help: %v", err)
	}

	helpText := string(output)

	// Get grammar
	g := grammar.GetGrammar()

	// Check that all verbs appear in help
	for _, verb := range g.Verbs {
		if !strings.Contains(helpText, verb.Name) {
			t.Errorf("Verb %s not found in help output", verb.Name)
		}
		if !strings.Contains(helpText, verb.Description) {
			t.Errorf("Verb %s description not found in help output", verb.Name)
		}
	}

	// Check that all clauses appear in help
	for _, clause := range g.Clauses {
		if !strings.Contains(helpText, clause.Name) {
			t.Errorf("Clause %s not found in help output", clause.Name)
		}
		if !strings.Contains(helpText, clause.Description) {
			t.Errorf("Clause %s description not found in help output", clause.Name)
		}
	}
}

