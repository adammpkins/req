package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/adammpkins/req/internal/planner"
	"github.com/adammpkins/req/internal/types"
)

func TestPlanRead(t *testing.T) {
	cmd := &types.Command{
		Verb:   types.VerbRead,
		Target: types.Target{URL: "https://api.example.com/users"},
		Clauses: []types.Clause{
			types.AsClause{Format: "json"},
		},
	}

	plan, err := planner.Plan(cmd)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	if plan.Method != http.MethodGet {
		t.Errorf("Plan() Method = %v, want %v", plan.Method, http.MethodGet)
	}
	if plan.URL != "https://api.example.com/users" {
		t.Errorf("Plan() URL = %v, want %v", plan.URL, "https://api.example.com/users")
	}
	if plan.Output == nil {
		t.Errorf("Plan() Output is nil")
	} else if plan.Output.Format != "json" {
		t.Errorf("Plan() Output.Format = %v, want json", plan.Output.Format)
	}
}

func TestPlanSend(t *testing.T) {
	cmd := &types.Command{
		Verb:   types.VerbSend,
		Target: types.Target{URL: "https://api.example.com/users"},
		Clauses: []types.Clause{
			types.WithClause{Type: "json", Value: "{\"name\":\"Ada\"}"},
		},
	}

	plan, err := planner.Plan(cmd)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	// Send with with= should default to POST
	if plan.Method != http.MethodPost {
		t.Errorf("Plan() Method = %v, want %v", plan.Method, http.MethodPost)
	}
	if plan.Body == nil {
		t.Errorf("Plan() Body is nil")
	} else if plan.Body.Type != "json" {
		t.Errorf("Plan() Body.Type = %v, want json", plan.Body.Type)
	}
}

func TestPlanSave(t *testing.T) {
	cmd := &types.Command{
		Verb:   types.VerbSave,
		Target: types.Target{URL: "https://example.com/file.zip"},
		Clauses: []types.Clause{
			types.ToClause{Destination: "file.zip"},
		},
	}

	plan, err := planner.Plan(cmd)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	if plan.Method != http.MethodGet {
		t.Errorf("Plan() Method = %v, want %v", plan.Method, http.MethodGet)
	}
	if plan.Output == nil {
		t.Errorf("Plan() Output is nil")
	} else if plan.Output.Destination != "file.zip" {
		t.Errorf("Plan() Output.Destination = %v, want file.zip", plan.Output.Destination)
	}
}

func TestPlanWithTimeout(t *testing.T) {
	cmd := &types.Command{
		Verb:   types.VerbRead,
		Target: types.Target{URL: "https://api.example.com/users"},
		Clauses: []types.Clause{
			types.TimeoutClause{Duration: 5 * time.Second},
		},
	}

	plan, err := planner.Plan(cmd)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	if plan.Timeout == nil {
		t.Errorf("Plan() Timeout is nil")
	} else if *plan.Timeout != 5*time.Second {
		t.Errorf("Plan() Timeout = %v, want %v", *plan.Timeout, 5*time.Second)
	}
}

func TestPlanWithRetry(t *testing.T) {
	cmd := &types.Command{
		Verb:   types.VerbRead,
		Target: types.Target{URL: "https://api.example.com/users"},
		Clauses: []types.Clause{
			types.RetryClause{Count: 3},
			types.BackoffClause{
				Min: 200 * time.Millisecond,
				Max: 5 * time.Second,
			},
		},
	}

	plan, err := planner.Plan(cmd)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	if plan.Retry == nil {
		t.Errorf("Plan() Retry is nil")
	} else {
		if plan.Retry.Count != 3 {
			t.Errorf("Plan() Retry.Count = %v, want 3", plan.Retry.Count)
		}
		if plan.Retry.Backoff.Min != 200*time.Millisecond {
			t.Errorf("Plan() Retry.Backoff.Min = %v, want %v", plan.Retry.Backoff.Min, 200*time.Millisecond)
		}
		if plan.Retry.Backoff.Max != 5*time.Second {
			t.Errorf("Plan() Retry.Backoff.Max = %v, want %v", plan.Retry.Backoff.Max, 5*time.Second)
		}
	}
}

func TestPlanWithMethodOverride(t *testing.T) {
	cmd := &types.Command{
		Verb:   types.VerbRead,
		Target: types.Target{URL: "https://api.example.com/users"},
		Clauses: []types.Clause{
			types.MethodClause{Method: http.MethodDelete},
		},
	}

	plan, err := planner.Plan(cmd)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	// Method override should take precedence
	if plan.Method != http.MethodDelete {
		t.Errorf("Plan() Method = %v, want %v", plan.Method, http.MethodDelete)
	}
}

