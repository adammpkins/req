// Package types provides shared types and enums used across the req package.
package types

import "time"

// Verb represents the action verb in a req command.
type Verb string

const (
	VerbRead         Verb = "read"
	VerbSave         Verb = "save"
	VerbSend         Verb = "send"
	VerbUpload       Verb = "upload"
	VerbWatch        Verb = "watch"
	VerbInspect      Verb = "inspect"
	VerbAuthenticate Verb = "authenticate"
	VerbSession      Verb = "session"
)

// Command represents a parsed req command AST.
type Command struct {
	Verb    Verb
	Target  Target
	Clauses []Clause
	// For session verb, subcommand (show, clear, use)
	SessionSubcommand string
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
	Value    string // inline value, file path, or "-" for stdin
	Type     string // json, form, etc. (inferred if empty)
	IsFile   bool   // true if value starts with @
	IsStdin  bool   // true if value is @-
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

// UsingClause represents a "using=" clause for HTTP method override.
type UsingClause struct {
	Method string // GET, POST, PUT, PATCH, DELETE, etc.
}

func (UsingClause) clause() {}

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

// VerboseClause represents the "verbose" flag.
type VerboseClause struct{}

func (VerboseClause) clause() {}

// ResumeClause represents the "resume" flag for resumable downloads.
type ResumeClause struct{}

func (ResumeClause) clause() {}

// IncludeClause represents an "include=" clause for headers, params, and cookies.
type IncludeClause struct {
	Items []IncludeItem
}

func (IncludeClause) clause() {}

// IncludeItem represents a single item in an include clause.
type IncludeItem struct {
	Type  string // "header", "param", "cookie", "basic"
	Name  string // header name, param key, or cookie key (empty for basic)
	Value string // header value, param value, cookie value, or username:password for basic
}

// AttachClause represents an "attach=" clause for multipart form data.
type AttachClause struct {
	Parts    []AttachPart
	Boundary string // optional explicit boundary
}

func (AttachClause) clause() {}

// AttachPart represents a single part in an attach clause.
type AttachPart struct {
	Name     string // required
	FilePath string // file=@path (mutually exclusive with Value)
	Value    string // value=... (mutually exclusive with FilePath)
	Filename string // optional filename
	Type     string // optional Content-Type
}

// ExpectClause represents an "expect=" clause for response assertions.
type ExpectClause struct {
	Checks []ExpectCheck
}

func (ExpectClause) clause() {}

// ExpectCheck represents a single assertion check.
type ExpectCheck struct {
	Type  string // "status", "header", "contains", "jsonpath", "matches"
	Name  string // for header checks, the header name
	Value string // the value to check against
	Path  string // for jsonpath, the JSONPath expression
	Regex string // for matches, the regex pattern
}

// FollowClause represents a "follow=" clause for redirect policy.
type FollowClause struct {
	Policy string // "smart" or empty for default
}

func (FollowClause) clause() {}

// UnderClause represents an "under=" clause for timeout or size limit.
type UnderClause struct {
	Duration time.Duration // if it's a duration
	Size     int64         // if it's a size (in bytes)
	IsSize   bool          // true if it's a size limit, false if duration
}

func (UnderClause) clause() {}

// ViaClause represents a "via=" clause for proxy URL.
type ViaClause struct {
	URL string
}

func (ViaClause) clause() {}

// InsecureClause represents an "insecure=" clause (updated to support true/false).
type InsecureClause struct {
	Value bool // true or false
}

func (InsecureClause) clause() {}

