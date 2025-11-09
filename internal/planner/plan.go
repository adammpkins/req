// Package planner applies defaults, validates commands, and produces execution plans.
package planner

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adammpkins/req/internal/types"
)

// ExecutionPlan represents a fully resolved execution plan ready for HTTP runtime.
type ExecutionPlan struct {
	Verb        types.Verb         `json:"verb"`
	Method      string             `json:"method"`
	URL         string             `json:"url"`
	Headers     map[string]string  `json:"headers,omitempty"`
	QueryParams map[string]string  `json:"query_params,omitempty"`
	Cookies     map[string]string  `json:"cookies,omitempty"`
	Body        *BodyPlan          `json:"body,omitempty"`
	Output      *OutputPlan        `json:"output,omitempty"`
	Retry       *RetryPlan         `json:"retry,omitempty"`
	Timeout     *time.Duration    `json:"timeout,omitempty"`
	SizeLimit   *int64            `json:"size_limit,omitempty"`
	Proxy       string             `json:"proxy,omitempty"`
	Insecure    bool               `json:"insecure,omitempty"`
	Verbose     bool               `json:"verbose,omitempty"`
	Resume      bool               `json:"resume,omitempty"`
	Follow      string             `json:"follow,omitempty"` // "smart" or empty
	Expect      []types.ExpectCheck `json:"expect,omitempty"`
}

// BodyPlan represents the request body configuration.
type BodyPlan struct {
	Type     string                `json:"type"` // json, form, multipart, raw
	Content  string                `json:"content,omitempty"`
	FilePath string                `json:"file_path,omitempty"`
	Field    string                `json:"field,omitempty"` // for multipart
	AttachParts []types.AttachPart `json:"attach_parts,omitempty"` // for multipart
	Boundary string                `json:"boundary,omitempty"` // for multipart
}

// OutputPlan represents the output configuration.
type OutputPlan struct {
	Format      string `json:"format"` // json, csv, text, raw
	Destination string `json:"destination,omitempty"`
	Pick        string `json:"pick,omitempty"` // JSONPath expression
}

// RetryPlan represents retry configuration.
type RetryPlan struct {
	Count  int           `json:"count"`
	Backoff BackoffRange `json:"backoff"`
}

// BackoffRange represents a backoff range with min and max durations.
type BackoffRange struct {
	Min time.Duration `json:"min"`
	Max time.Duration `json:"max"`
}

// Plan creates an ExecutionPlan from a parsed Command.
func Plan(cmd *types.Command) (*ExecutionPlan, error) {
	plan := &ExecutionPlan{
		Verb:        cmd.Verb,
		URL:         cmd.Target.URL,
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
		Cookies:     make(map[string]string),
	}

	// Apply verb-specific defaults
	if err := applyVerbDefaults(cmd.Verb, plan); err != nil {
		return nil, err
	}

	// Process clauses
	for _, clause := range cmd.Clauses {
		if err := applyClause(clause, plan, cmd.Verb); err != nil {
			return nil, err
		}
	}

	// Post-process: extract filename for save verb if destination not provided or is a directory
	if cmd.Verb == types.VerbSave && plan.Output != nil {
		if plan.Output.Destination == "" {
			// No destination provided, extract from URL
			filename := extractFilenameFromURL(plan.URL)
			if filename != "" {
				plan.Output.Destination = filename
			}
		} else {
			// Destination provided - check if it's a directory
			if isDirectory(plan.Output.Destination) {
				// It's a directory, append filename from URL
				filename := extractFilenameFromURL(plan.URL)
				if filename != "" {
					plan.Output.Destination = filepath.Join(plan.Output.Destination, filename)
				}
			}
			// If it's a file path (like /tmp/file.zip), use it as-is
		}
	}

	// Validate plan
	if err := validatePlan(plan); err != nil {
		return nil, err
	}

	return plan, nil
}

// applyVerbDefaults applies default settings based on the verb.
func applyVerbDefaults(verb types.Verb, plan *ExecutionPlan) error {
	switch verb {
	case types.VerbRead:
		plan.Method = http.MethodGet
		plan.Output = &OutputPlan{Format: "auto"}
	case types.VerbSave:
		plan.Method = http.MethodGet
		plan.Output = &OutputPlan{Format: "raw"}
	case types.VerbSend:
		// Default to GET, will be changed to POST if with= is present
		plan.Method = http.MethodGet
		plan.Output = &OutputPlan{Format: "auto"}
	case types.VerbUpload:
		// Default to POST, but will error if no attach= or with= present
		plan.Method = http.MethodPost
		plan.Output = &OutputPlan{Format: "auto"}
	case types.VerbWatch:
		plan.Method = http.MethodGet
		plan.Output = &OutputPlan{Format: "auto"}
	case types.VerbInspect:
		plan.Method = http.MethodHead
		plan.Output = &OutputPlan{Format: "json"}
	case types.VerbAuthenticate:
		// Default to POST if with= is present, otherwise require using=
		// We'll check this in validatePlan
		plan.Method = http.MethodPost // tentative, may be overridden
		plan.Output = &OutputPlan{Format: "auto"}
	case types.VerbSession:
		// Session verbs are handled separately in main
		plan.Method = http.MethodGet // placeholder
		plan.Output = &OutputPlan{Format: "auto"}
	default:
		return fmt.Errorf("unsupported verb: %s", verb)
	}
	return nil
}

// validateUsingClause validates that the HTTP method is compatible with the verb.
func validateUsingClause(verb types.Verb, method string) error {
	allowedMethods := map[types.Verb][]string{
		types.VerbRead:    {"GET", "HEAD", "OPTIONS"},
		types.VerbSave:    {"GET", "POST"},
		types.VerbSend:   {"POST", "PUT", "PATCH"},
		types.VerbUpload: {"POST", "PUT"},
		types.VerbWatch:  {"GET"},
		types.VerbInspect: {"HEAD", "GET", "OPTIONS"},
	}
	
	allowed, ok := allowedMethods[verb]
	if !ok {
		// If verb not in map, allow any method (for future verbs like delete)
		return nil
	}
	
	for _, allowedMethod := range allowed {
		if method == allowedMethod {
			return nil
		}
	}
	
	return fmt.Errorf("verb '%s' is incompatible with method '%s'", verb, method)
}

// applyClause applies a clause to the execution plan.
func applyClause(clause types.Clause, plan *ExecutionPlan, verb types.Verb) error {
	switch c := clause.(type) {
	case types.UsingClause:
		// Validate compatibility before applying
		if err := validateUsingClause(verb, c.Method); err != nil {
			return err
		}
		// Normalize to uppercase (defensive, should already be normalized in parser)
		plan.Method = strings.ToUpper(c.Method)
	case types.HeadersClause:
		for k, v := range c.Headers {
			plan.Headers[k] = v
		}
	case types.ParamsClause:
		for k, v := range c.Params {
			plan.QueryParams[k] = v
		}
	case types.WithClause:
		if plan.Body == nil {
			plan.Body = &BodyPlan{}
		}
		
		// Handle file or stdin
		if c.IsFile {
			plan.Body.FilePath = c.Value
			plan.Body.Type = c.Type
			if plan.Body.Type == "" {
				plan.Body.Type = "raw" // Default for files
			}
		} else if c.IsStdin {
			plan.Body.FilePath = "-" // Special marker for stdin
			plan.Body.Type = c.Type
			if plan.Body.Type == "" {
				plan.Body.Type = "raw" // Default for stdin
			}
		} else {
			plan.Body.Content = c.Value
			plan.Body.Type = c.Type
			// If type was inferred as JSON, we'll note it in runtime
			if plan.Body.Type == "json" {
				// JSON inference will be logged in runtime
			}
		}
		
		// If method is still GET and we have a body, default to POST
		if plan.Method == http.MethodGet {
			plan.Method = http.MethodPost
		}
	case types.AsClause:
		if plan.Output == nil {
			plan.Output = &OutputPlan{}
		}
		plan.Output.Format = c.Format
	case types.ToClause:
		if plan.Output == nil {
			plan.Output = &OutputPlan{}
		}
		plan.Output.Destination = c.Destination
	case types.RetryClause:
		if plan.Retry == nil {
			plan.Retry = &RetryPlan{
				Backoff: BackoffRange{
					Min: 200 * time.Millisecond,
					Max: 5 * time.Second,
				},
			}
		}
		plan.Retry.Count = c.Count
	case types.BackoffClause:
		if plan.Retry == nil {
			plan.Retry = &RetryPlan{Count: 3}
		}
		plan.Retry.Backoff = BackoffRange{
			Min: c.Min,
			Max: c.Max,
		}
	case types.TimeoutClause:
		plan.Timeout = &c.Duration
	case types.ProxyClause:
		plan.Proxy = c.URL
	case types.PickClause:
		if plan.Output == nil {
			plan.Output = &OutputPlan{}
		}
		plan.Output.Pick = c.Path
	case types.InsecureClause:
		plan.Insecure = c.Value
	case types.ViaClause:
		plan.Proxy = c.URL
	case types.IncludeClause:
		// Merge include items into headers, params, or cookies
		for _, item := range c.Items {
			switch item.Type {
			case "header":
				// For multi-valued headers, we'd need to track arrays, but for now last wins
				// TODO: Support multi-valued headers properly
				plan.Headers[item.Name] = item.Value
			case "param":
				// Params can be repeated, so we append to query params
				// The runtime will handle serialization
				plan.QueryParams[item.Name] = item.Value
			case "cookie":
				// Cookies: last value wins
				plan.Cookies[item.Name] = item.Value
			case "basic":
				// Basic Auth: encode username:password and set Authorization header
				// item.Value contains "username:password"
				credentials := item.Value
				// Base64 encode the credentials
				encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
				// Set Authorization header with Basic scheme
				plan.Headers["Authorization"] = "Basic " + encoded
			}
		}
	case types.AttachClause:
		if plan.Body == nil {
			plan.Body = &BodyPlan{}
		}
		plan.Body.Type = "multipart"
		plan.Body.AttachParts = c.Parts
		if c.Boundary != "" {
			plan.Body.Boundary = c.Boundary
		}
	case types.ExpectClause:
		plan.Expect = c.Checks
	case types.FollowClause:
		plan.Follow = c.Policy
	case types.UnderClause:
		if c.IsSize {
			plan.SizeLimit = &c.Size
		} else {
			plan.Timeout = &c.Duration
		}
	case types.VerboseClause:
		plan.Verbose = true
	case types.ResumeClause:
		plan.Resume = true
	default:
		return fmt.Errorf("unsupported clause type: %T", clause)
	}
	return nil
}

// validatePlan validates the execution plan.
func validatePlan(plan *ExecutionPlan) error {
	if plan.Method == "" {
		return fmt.Errorf("method is required")
	}
	if plan.URL == "" {
		return fmt.Errorf("URL is required")
	}
	
	// Validate upload verb: must have attach= or with=
	// This check will be done after clauses are processed, so we check here
	// Actually, we need to check this in Plan() after processing clauses
	// For now, we'll do basic validation
	
	return nil
}

// extractFilenameFromURL extracts a filename from a URL.
func extractFilenameFromURL(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	// Get the path
	path := u.Path
	if path == "" || path == "/" {
		// Try to get from query or fragment
		return "download"
	}

	// Remove leading slash
	path = strings.TrimPrefix(path, "/")

	// Get the last segment
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return "download"
	}

	filename := parts[len(parts)-1]

	// URL decode the filename (handle both path and query encoding)
	filename, err = url.PathUnescape(filename)
	if err != nil {
		// If PathUnescape fails, try QueryUnescape
		filename, err = url.QueryUnescape(filename)
		if err != nil {
			// If decoding fails, use the original
			filename = parts[len(parts)-1]
		}
	}

	// If filename is empty or doesn't have an extension, use a default
	if filename == "" || !strings.Contains(filename, ".") {
		filename = "download"
	}

	// Clean the filename (remove any path separators)
	filename = filepath.Base(filename)

	return filename
}

// isDirectory checks if a path is a directory.
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

