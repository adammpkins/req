package grammar

import "encoding/json"

// Snapshot represents a snapshot of the grammar for drift detection.
type Snapshot struct {
	Verbs   []string `json:"verbs"`
	Clauses []ClauseSnapshot `json:"clauses"`
}

// ClauseSnapshot represents a clause in the snapshot.
type ClauseSnapshot struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Repeatable  bool   `json:"repeatable"`
}

// GetSnapshot returns a JSON-serializable snapshot of the grammar.
func GetSnapshot() Snapshot {
	g := GetGrammar()
	
	verbs := make([]string, len(g.Verbs))
	for i, v := range g.Verbs {
		verbs[i] = v.Name
	}
	
	clauses := make([]ClauseSnapshot, len(g.Clauses))
	for i, c := range g.Clauses {
		clauses[i] = ClauseSnapshot{
			Name:        c.Name,
			Description: c.Description,
			Repeatable:  c.Repeatable,
		}
	}
	
	return Snapshot{
		Verbs:   verbs,
		Clauses: clauses,
	}
}

// GetSnapshotJSON returns the snapshot as JSON bytes.
func GetSnapshotJSON() ([]byte, error) {
	snapshot := GetSnapshot()
	return json.MarshalIndent(snapshot, "", "  ")
}

