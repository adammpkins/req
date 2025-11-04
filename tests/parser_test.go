package tests

import (
	"testing"

	"github.com/adammpkins/req/internal/parser"
	"github.com/adammpkins/req/internal/types"
)

func TestParseBasicRead(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *types.Command
		wantErr bool
	}{
		{
			name:  "simple read",
			input: "read https://api.example.com/users",
			want: &types.Command{
				Verb:   types.VerbRead,
				Target: types.Target{URL: "https://api.example.com/users"},
			},
			wantErr: false,
		},
		{
			name:  "read with as clause",
			input: "read https://api.example.com/users as=json",
			want: &types.Command{
				Verb:   types.VerbRead,
				Target: types.Target{URL: "https://api.example.com/users"},
				Clauses: []types.Clause{
					types.AsClause{Format: "json"},
				},
			},
			wantErr: false,
		},
		{
			name:  "read with multiple clauses",
			input: "read https://api.example.com/users as=json verbose",
			want: &types.Command{
				Verb:   types.VerbRead,
				Target: types.Target{URL: "https://api.example.com/users"},
				Clauses: []types.Clause{
					types.AsClause{Format: "json"},
					types.VerboseClause{},
				},
			},
			wantErr: false,
		},
		{
			name:  "send with json body",
			input: "send https://api.example.com/users with=json:'{\"name\":\"Ada\"}'",
			want: &types.Command{
				Verb:   types.VerbSend,
				Target: types.Target{URL: "https://api.example.com/users"},
				Clauses: []types.Clause{
					types.WithClause{Type: "json", Value: "'{\"name\":\"Ada\"}'"},
				},
			},
			wantErr: false,
		},
		{
			name:  "save with destination",
			input: "save https://example.com/file.zip to=file.zip",
			want: &types.Command{
				Verb:   types.VerbSave,
				Target: types.Target{URL: "https://example.com/file.zip"},
				Clauses: []types.Clause{
					types.ToClause{Destination: "file.zip"},
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid verb",
			input:   "invalid https://example.com",
			wantErr: true,
		},
		{
			name:    "missing target",
			input:   "read",
			wantErr: true,
		},
		{
			name:  "read with insecure flag",
			input: "read https://api.example.com/users insecure",
			want: &types.Command{
				Verb:   types.VerbRead,
				Target: types.Target{URL: "https://api.example.com/users"},
				Clauses: []types.Clause{
					types.InsecureClause{},
				},
			},
			wantErr: false,
		},
		{
			name:  "read with timeout",
			input: "read https://api.example.com/users timeout=5s",
			want: &types.Command{
				Verb:   types.VerbRead,
				Target: types.Target{URL: "https://api.example.com/users"},
				Clauses: []types.Clause{
					types.TimeoutClause{Duration: 5000000000}, // 5s in nanoseconds
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Verb != tt.want.Verb {
				t.Errorf("Parse() Verb = %v, want %v", got.Verb, tt.want.Verb)
			}
			if got.Target.URL != tt.want.Target.URL {
				t.Errorf("Parse() Target.URL = %v, want %v", got.Target.URL, tt.want.Target.URL)
			}
			if len(got.Clauses) != len(tt.want.Clauses) {
				t.Errorf("Parse() Clauses length = %v, want %v", len(got.Clauses), len(tt.want.Clauses))
			}
		})
	}
}

func TestParseErrorSuggestions(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantSuggestion string
	}{
		{
			name: "typo in verb",
			input: "reed https://api.example.com/users",
			wantSuggestion: "read",
		},
		{
			name: "typo in clause",
			input: "read https://api.example.com/users ass=json",
			wantSuggestion: "as",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse(tt.input)
			if err == nil {
				t.Errorf("Parse() expected error but got none")
				return
			}
			parseErr, ok := err.(*parser.ParseError)
			if !ok {
				t.Errorf("Parse() error is not a ParseError: %T", err)
				return
			}
			if parseErr.Suggest != tt.wantSuggestion {
				t.Errorf("Parse() Suggest = %v, want %v", parseErr.Suggest, tt.wantSuggestion)
			}
		})
	}
}

