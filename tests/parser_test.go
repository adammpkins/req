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
			input: "send https://api.example.com/users with='{\"name\":\"Ada\"}'",
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
			name:  "read with insecure clause",
			input: "read https://api.example.com/users insecure=true",
			want: &types.Command{
				Verb:   types.VerbRead,
				Target: types.Target{URL: "https://api.example.com/users"},
				Clauses: []types.Clause{
					types.InsecureClause{Value: true},
				},
			},
			wantErr: false,
		},
		{
			name:  "read with timeout",
			input: "read https://api.example.com/users under=5s",
			want: &types.Command{
				Verb:   types.VerbRead,
				Target: types.Target{URL: "https://api.example.com/users"},
				Clauses: []types.Clause{
					types.UnderClause{Duration: 5000000000, IsSize: false}, // 5s in nanoseconds
				},
			},
			wantErr: false,
		},
		{
			name:  "send with using=PUT",
			input: "send https://api.example.com/users using=PUT",
			want: &types.Command{
				Verb:   types.VerbSend,
				Target: types.Target{URL: "https://api.example.com/users"},
				Clauses: []types.Clause{
					types.UsingClause{Method: "PUT"},
				},
			},
			wantErr: false,
		},
		{
			name:  "send with using=patch (normalize to uppercase)",
			input: "send https://api.example.com/users using=patch",
			want: &types.Command{
				Verb:   types.VerbSend,
				Target: types.Target{URL: "https://api.example.com/users"},
				Clauses: []types.Clause{
					types.UsingClause{Method: "PATCH"},
				},
			},
			wantErr: false,
		},
		{
			name:  "read with using=HEAD",
			input: "read https://api.example.com/users using=HEAD",
			want: &types.Command{
				Verb:   types.VerbRead,
				Target: types.Target{URL: "https://api.example.com/users"},
				Clauses: []types.Clause{
					types.UsingClause{Method: "HEAD"},
				},
			},
			wantErr: false,
		},
		{
			name:    "using= with invalid method",
			input:   "read https://api.example.com/users using=INVALID",
			wantErr: true,
		},
		{
			name:    "old method= syntax should error",
			input:   "read https://api.example.com/users method=PUT",
			wantErr: true,
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
				return
			}
			// Check UsingClause if present
			for i, clause := range got.Clauses {
				if usingClause, ok := clause.(types.UsingClause); ok {
					if i >= len(tt.want.Clauses) {
						t.Errorf("Parse() UsingClause found but not expected")
						continue
					}
					if wantUsingClause, ok := tt.want.Clauses[i].(types.UsingClause); ok {
						if usingClause.Method != wantUsingClause.Method {
							t.Errorf("Parse() UsingClause.Method = %v, want %v", usingClause.Method, wantUsingClause.Method)
						}
					}
				}
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

func TestParseQuotedStringsWithSemicolons(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *types.Command)
	}{
		{
			name:    "include with semicolon in header value",
			input:   `read https://api.example.com/users include="header: Content-Type: application/json; charset=UTF-8"`,
			wantErr: false,
			check: func(t *testing.T, cmd *types.Command) {
				if len(cmd.Clauses) != 1 {
					t.Fatalf("expected 1 clause, got %d", len(cmd.Clauses))
				}
				includeClause, ok := cmd.Clauses[0].(types.IncludeClause)
				if !ok {
					t.Fatalf("expected IncludeClause, got %T", cmd.Clauses[0])
				}
				if len(includeClause.Items) != 1 {
					t.Fatalf("expected 1 include item, got %d", len(includeClause.Items))
				}
				item := includeClause.Items[0]
				if item.Type != "header" {
					t.Errorf("expected header type, got %s", item.Type)
				}
				if item.Name != "Content-Type" {
					t.Errorf("expected header name Content-Type, got %s", item.Name)
				}
				if item.Value != "application/json; charset=UTF-8" {
					t.Errorf("expected header value 'application/json; charset=UTF-8', got %s", item.Value)
				}
			},
		},
		{
			name:    "expect with semicolon in header value",
			input:   `read https://api.example.com/users expect="status:200, header:Content-Type=application/json; charset=utf-8"`,
			wantErr: false,
			check: func(t *testing.T, cmd *types.Command) {
				if len(cmd.Clauses) != 1 {
					t.Fatalf("expected 1 clause, got %d", len(cmd.Clauses))
				}
				expectClause, ok := cmd.Clauses[0].(types.ExpectClause)
				if !ok {
					t.Fatalf("expected ExpectClause, got %T", cmd.Clauses[0])
				}
				if len(expectClause.Checks) != 2 {
					t.Fatalf("expected 2 expect checks, got %d", len(expectClause.Checks))
				}
				headerCheck := expectClause.Checks[1]
				if headerCheck.Type != "header" {
					t.Errorf("expected header check type, got %s", headerCheck.Type)
				}
				if headerCheck.Name != "Content-Type" {
					t.Errorf("expected header name Content-Type, got %s", headerCheck.Name)
				}
				if headerCheck.Value != "application/json; charset=utf-8" {
					t.Errorf("expected header value 'application/json; charset=utf-8', got %s", headerCheck.Value)
				}
			},
		},
		{
			name:    "multiple include items with semicolons",
			input:   `read https://api.example.com/users include="header: Accept: application/json; q=0.9; param: q=test"`,
			wantErr: false,
			check: func(t *testing.T, cmd *types.Command) {
				includeClause := cmd.Clauses[0].(types.IncludeClause)
				if len(includeClause.Items) != 2 {
					t.Fatalf("expected 2 include items, got %d", len(includeClause.Items))
				}
				headerItem := includeClause.Items[0]
				if headerItem.Value != "application/json; q=0.9" {
					t.Errorf("expected header value 'application/json; q=0.9', got %s", headerItem.Value)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := parser.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && tt.check != nil {
				tt.check(t, cmd)
			}
		})
	}
}

func TestParseBasicAuth(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *types.Command)
	}{
		{
			name:    "basic auth with username:password",
			input:   `read https://httpbin.org/basic-auth/user/passwd include='basic: user:passwd'`,
			wantErr: false,
			check: func(t *testing.T, cmd *types.Command) {
				if len(cmd.Clauses) != 1 {
					t.Fatalf("expected 1 clause, got %d", len(cmd.Clauses))
				}
				includeClause, ok := cmd.Clauses[0].(types.IncludeClause)
				if !ok {
					t.Fatalf("expected IncludeClause, got %T", cmd.Clauses[0])
				}
				if len(includeClause.Items) != 1 {
					t.Fatalf("expected 1 include item, got %d", len(includeClause.Items))
				}
				item := includeClause.Items[0]
				if item.Type != "basic" {
					t.Errorf("expected basic type, got %s", item.Type)
				}
				if item.Value != "user:passwd" {
					t.Errorf("expected value 'user:passwd', got %s", item.Value)
				}
			},
		},
		{
			name:    "basic auth with quoted credentials",
			input:   `read https://httpbin.org/basic-auth/user/passwd include="basic: user:passwd"`,
			wantErr: false,
			check: func(t *testing.T, cmd *types.Command) {
				includeClause := cmd.Clauses[0].(types.IncludeClause)
				item := includeClause.Items[0]
				if item.Type != "basic" {
					t.Errorf("expected basic type, got %s", item.Type)
				}
				if item.Value != "user:passwd" {
					t.Errorf("expected value 'user:passwd', got %s", item.Value)
				}
			},
		},
		{
			name:    "basic auth combined with other include items",
			input:   `read https://httpbin.org/basic-auth/user/passwd include='basic: user:passwd; header: Accept: application/json'`,
			wantErr: false,
			check: func(t *testing.T, cmd *types.Command) {
				includeClause := cmd.Clauses[0].(types.IncludeClause)
				if len(includeClause.Items) != 2 {
					t.Fatalf("expected 2 include items, got %d", len(includeClause.Items))
				}
				basicItem := includeClause.Items[0]
				if basicItem.Type != "basic" {
					t.Errorf("expected first item to be basic, got %s", basicItem.Type)
				}
				headerItem := includeClause.Items[1]
				if headerItem.Type != "header" {
					t.Errorf("expected second item to be header, got %s", headerItem.Type)
				}
			},
		},
		{
			name:    "basic auth missing colon",
			input:   `read https://httpbin.org/basic-auth/user/passwd include='basic: userpass'`,
			wantErr: true,
		},
		{
			name:    "basic auth with empty username",
			input:   `read https://httpbin.org/basic-auth/user/passwd include='basic: :passwd'`,
			wantErr: false,
			check: func(t *testing.T, cmd *types.Command) {
				includeClause := cmd.Clauses[0].(types.IncludeClause)
				item := includeClause.Items[0]
				if item.Value != ":passwd" {
					t.Errorf("expected value ':passwd', got %s", item.Value)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := parser.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && tt.check != nil {
				tt.check(t, cmd)
			}
		})
	}
}

