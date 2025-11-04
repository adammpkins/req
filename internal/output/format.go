// Package output provides formatting and pretty-printing for execution plans.
package output

import (
	"encoding/json"
	"os"

	"github.com/adammpkins/req/internal/planner"
	"github.com/mattn/go-isatty"
)

// FormatPlan formats an ExecutionPlan as JSON for output.
func FormatPlan(plan *planner.ExecutionPlan) ([]byte, error) {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		// Pretty print when outputting to terminal
		return json.MarshalIndent(plan, "", "  ")
	}
	// Compact JSON when piped
	return json.Marshal(plan)
}

