// Package types provides shared types and enums used across the req package.
package types

import "time"

// Verb represents the action verb in a req command.
type Verb string

const (
	VerbRead    Verb = "read"
	VerbSave    Verb = "save"
	VerbSend    Verb = "send"
	VerbUpload  Verb = "upload"
	VerbWatch   Verb = "watch"
	VerbInspect Verb = "inspect"
	VerbAuth    Verb = "auth"
	VerbSession Verb = "session"
	VerbProfile Verb = "profile"
)

// Command represents a parsed req command AST.
type Command struct {
	Verb    Verb
	Target  Target
	Clauses []Clause
}

// Target represents the URL or resource being acted upon.
type Target struct {
	URL string
}

// Clause represents a modifier clause in the command.
// This is a sum type that will be expanded as we add more clause types.
type Clause interface {
	clause()
}

// WithClause represents a "with=" clause for request body.
type WithClause struct {
	Value string
	Type  string // json, form, etc.
}

func (WithClause) clause() {}

// HeadersClause represents a "headers=" clause.
type HeadersClause struct {
	Headers map[string]string
}

func (HeadersClause) clause() {}

// ParamsClause represents a "params=" clause for query parameters.
type ParamsClause struct {
	Params map[string]string
}

func (ParamsClause) clause() {}

// AsClause represents an "as=" clause for output format.
type AsClause struct {
	Format string // json, csv, text, raw
}

func (AsClause) clause() {}

// ToClause represents a "to=" clause for destination.
type ToClause struct {
	Destination string
}

func (ToClause) clause() {}

// MethodClause represents a "method=" clause.
type MethodClause struct {
	Method string // GET, POST, PUT, DELETE, etc.
}

func (MethodClause) clause() {}

// RetryClause represents a "retry=" clause.
type RetryClause struct {
	Count int
}

func (RetryClause) clause() {}

// BackoffClause represents a "backoff=" clause.
type BackoffClause struct {
	Min time.Duration
	Max time.Duration
}

func (BackoffClause) clause() {}

// TimeoutClause represents a "timeout=" clause.
type TimeoutClause struct {
	Duration time.Duration
}

func (TimeoutClause) clause() {}

// ProxyClause represents a "proxy=" clause.
type ProxyClause struct {
	URL string
}

func (ProxyClause) clause() {}

// PickClause represents a "pick=" clause for JSON path selection.
type PickClause struct {
	Path string // JSONPath expression
}

func (PickClause) clause() {}

// EveryClause represents an "every=" clause for polling.
type EveryClause struct {
	Interval time.Duration
}

func (EveryClause) clause() {}

// UntilClause represents an "until=" clause for conditional polling.
type UntilClause struct {
	Predicate string
}

func (UntilClause) clause() {}

// FieldClause represents a "field=" clause for multipart uploads.
type FieldClause struct {
	Name  string
	Value string
}

func (FieldClause) clause() {}

// InsecureClause represents the "insecure" flag.
type InsecureClause struct{}

func (InsecureClause) clause() {}

// VerboseClause represents the "verbose" flag.
type VerboseClause struct{}

func (VerboseClause) clause() {}

// ResumeClause represents the "resume" flag for resumable downloads.
type ResumeClause struct{}

func (ResumeClause) clause() {}

