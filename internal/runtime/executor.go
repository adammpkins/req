// Package runtime executes HTTP requests based on execution plans.
package runtime

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/adammpkins/req/internal/planner"
	"github.com/adammpkins/req/internal/types"
	"github.com/adammpkins/req/internal/session"
)

// Executor executes HTTP requests.
type Executor struct {
	client *http.Client
}

// NewExecutor creates a new executor.
func NewExecutor(plan *planner.ExecutionPlan) (*Executor, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}

	// Configure TLS if insecure
	if plan.Insecure {
		transport.TLSClientConfig = getInsecureTLSConfig()
		fmt.Fprintf(os.Stderr, "Warning: TLS verification disabled\n")
	}

	// Configure proxy if specified
	if plan.Proxy != "" {
		proxyURL, err := url.Parse(plan.Proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
		Jar:       jar,
	}

	if plan.Timeout != nil {
		client.Timeout = *plan.Timeout
	}

	return &Executor{client: client}, nil
}

// Execute executes an HTTP request based on the plan.
func (e *Executor) Execute(plan *planner.ExecutionPlan) error {
	// Build request URL with query parameters (preserving order)
	reqURL, err := e.buildURL(plan)
	if err != nil {
		return &ExecutionError{Code: 5, Message: fmt.Sprintf("invalid URL: %v", err)}
	}

	// Build request body
	body, contentType, err := e.buildBody(plan)
	if err != nil {
		return &ExecutionError{Code: 5, Message: fmt.Sprintf("failed to build body: %v", err)}
	}

	// Create request
	req, err := http.NewRequest(plan.Method, reqURL, body)
	if err != nil {
		return &ExecutionError{Code: 5, Message: fmt.Sprintf("failed to create request: %v", err)}
	}

	// Set headers
	e.setHeaders(req, plan, contentType)

	// Set cookies
	e.setCookies(req, plan)

	// Auto-apply session if available and not explicitly set
	e.autoApplySession(req, plan)

	// Add Accept-Encoding if not set by user
	if req.Header.Get("Accept-Encoding") == "" {
		req.Header.Set("Accept-Encoding", "gzip, br")
	}

	// Execute request with redirect handling
	// For authenticate verb, we need to capture Set-Cookie from redirect responses
	var resp *http.Response
	var redirectTrace []string
	var bodyBytes []byte
	var decompressed bool
	var allSetCookies []string

	if plan.Verb == types.VerbAuthenticate {
		resp, redirectTrace, allSetCookies, err = e.executeWithRedirectsCapturingCookies(req, plan)
		if err != nil {
			return &ExecutionError{Code: 4, Message: fmt.Sprintf("request failed: %v", err)}
		}
		defer resp.Body.Close()
		// Also include Set-Cookie from final response
		allSetCookies = append(allSetCookies, resp.Header.Values("Set-Cookie")...)
	} else {
		resp, redirectTrace, err = e.executeWithRedirects(req, plan)
		if err != nil {
			return &ExecutionError{Code: 4, Message: fmt.Sprintf("request failed: %v", err)}
		}
		defer resp.Body.Close()
	}

	// Print redirect trace to stderr
	if len(redirectTrace) > 0 {
		for _, trace := range redirectTrace {
			fmt.Fprintf(os.Stderr, "%s\n", trace)
		}
	}

	// Read and decompress response body
	bodyBytes, decompressed, err = e.readAndDecompress(resp)
	if err != nil {
		return &ExecutionError{Code: 4, Message: fmt.Sprintf("failed to read response: %v", err)}
	}

	if decompressed {
		fmt.Fprintf(os.Stderr, "Decompressed response\n")
	}

	// Print meta to stderr
	e.printMeta(resp, reqURL, len(bodyBytes), decompressed)

	// Capture session for authenticate verb
	if plan.Verb == types.VerbAuthenticate {
		host, err := session.ExtractHost(plan.URL)
		if err == nil {
			updatedSession, err := session.UpdateSessionFromResponse(host, allSetCookies, bodyBytes)
			if err == nil && updatedSession != nil {
				if err := session.SaveSession(updatedSession); err == nil {
					fmt.Fprintf(os.Stderr, "Session saved for %s\n", host)
				}
			}
		}
	}

	// Run expect checks
	if len(plan.Expect) > 0 {
		if err := e.runExpectChecks(resp, bodyBytes, plan.Expect); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			return &ExecutionError{Code: 3, Message: "expectation failed"}
		}
	} else {
		// If no expect checks, fail on non-2xx status codes
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return &ExecutionError{Code: 4, Message: fmt.Sprintf("HTTP %d %s", resp.StatusCode, resp.Status)}
		}
	}

	// Handle output based on plan
	if plan.Output != nil && plan.Output.Destination != "" {
		// Save to file - uses io.Copy for efficient writing
		// TODO: Optimize to stream directly from resp.Body when no expect checks
		return e.saveToFile(bytes.NewReader(bodyBytes), plan.Output.Destination)
	}

	// Handle watch verb with TTY detection
	if plan.Verb == types.VerbWatch {
		// TODO: Implement TTY detection
		// TTY: timestamped lines
		// Non-TTY: raw lines
		return e.writeOutput(bodyBytes, plan.Output)
	}

	// Format and write output
	return e.writeOutput(bodyBytes, plan.Output)
}

// ExecutionError represents an execution error with exit code.
type ExecutionError struct {
	Code    int
	Message string
}

func (e *ExecutionError) Error() string {
	return e.Message
}

// buildURL builds the request URL with query parameters, preserving order.
func (e *Executor) buildURL(plan *planner.ExecutionPlan) (string, error) {
	u, err := url.Parse(plan.URL)
	if err != nil {
		return "", err
	}

	// Merge existing query params with new ones
	existingParams := u.Query()
	for k, v := range plan.QueryParams {
		// Append to preserve order for repeated keys
		existingParams.Add(k, v)
	}
	u.RawQuery = existingParams.Encode()

	return u.String(), nil
}

// buildBody builds the request body.
func (e *Executor) buildBody(plan *planner.ExecutionPlan) (io.Reader, string, error) {
	if plan.Body == nil {
		return nil, "", nil
	}

	// Handle multipart
	if plan.Body.Type == "multipart" {
		return e.buildMultipartBody(plan.Body)
	}

	// Handle file or stdin
	if plan.Body.FilePath != "" {
		if plan.Body.FilePath == "-" {
			// Read from stdin
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return nil, "", fmt.Errorf("failed to read stdin: %w", err)
			}
			plan.Body.Content = string(data)
		} else {
			// Read from file
			data, err := os.ReadFile(plan.Body.FilePath)
			if err != nil {
				return nil, "", fmt.Errorf("failed to read file %s: %w", plan.Body.FilePath, err)
			}
			plan.Body.Content = string(data)
		}
	}

	// Determine content type
	contentType := ""
	if plan.Body.Type == "json" {
		contentType = "application/json"
		// Log JSON inference if it was inferred
		if strings.HasPrefix(strings.TrimSpace(plan.Body.Content), "{") || strings.HasPrefix(strings.TrimSpace(plan.Body.Content), "[") {
			fmt.Fprintf(os.Stderr, "Inferred Content-Type: application/json\n")
		}
	} else if plan.Body.Type == "form" {
		contentType = "application/x-www-form-urlencoded"
	}

	return strings.NewReader(plan.Body.Content), contentType, nil
}

// buildMultipartBody builds a multipart/form-data body.
func (e *Executor) buildMultipartBody(bodyPlan *planner.BodyPlan) (io.Reader, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	boundary := bodyPlan.Boundary
	if boundary == "" {
		boundary = writer.Boundary()
	} else {
		writer.SetBoundary(boundary)
	}

	for _, part := range bodyPlan.AttachParts {
		var partWriter io.Writer
		var err error

		// Create form field
		if part.Filename != "" {
			partWriter, err = writer.CreateFormFile(part.Name, part.Filename)
		} else {
			partWriter, err = writer.CreateFormField(part.Name)
		}
		if err != nil {
			return nil, "", fmt.Errorf("failed to create form field: %w", err)
		}

		// Write part content
		if part.FilePath != "" {
			// Read file
			data, err := os.ReadFile(part.FilePath)
			if err != nil {
				return nil, "", fmt.Errorf("failed to read file %s: %w", part.FilePath, err)
			}
			if _, err := partWriter.Write(data); err != nil {
				return nil, "", fmt.Errorf("failed to write file data: %w", err)
			}
		} else {
			// Write value
			if _, err := partWriter.Write([]byte(part.Value)); err != nil {
				return nil, "", fmt.Errorf("failed to write value: %w", err)
			}
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
	return &buf, contentType, nil
}

// setHeaders sets request headers.
func (e *Executor) setHeaders(req *http.Request, plan *planner.ExecutionPlan, contentType string) {
	// Set user headers first
	for k, v := range plan.Headers {
		req.Header.Set(k, v)
	}

	// Override Content-Type if multipart (user may have set it manually)
	if plan.Body != nil && plan.Body.Type == "multipart" {
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
			// Check if user had set Content-Type manually
			if _, wasSet := plan.Headers["Content-Type"]; wasSet {
				fmt.Fprintf(os.Stderr, "Note: Content-Type overridden for multipart\n")
			}
		}
	} else if contentType != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", contentType)
	}
}

// setCookies sets request cookies.
func (e *Executor) setCookies(req *http.Request, plan *planner.ExecutionPlan) {
	for name, value := range plan.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: value,
		})
	}
}

// autoApplySession automatically applies a stored session if available.
func (e *Executor) autoApplySession(req *http.Request, plan *planner.ExecutionPlan) {
	// Don't auto-apply if Authorization or Cookie headers are explicitly set
	hasAuth := req.Header.Get("Authorization") != ""
	hasCookie := false
	for name := range plan.Cookies {
		if name != "" {
			hasCookie = true
			break
		}
	}
	if hasAuth || hasCookie {
		return
	}

	// Extract host from URL
	host, err := session.ExtractHost(plan.URL)
	if err != nil {
		return
	}

	// Load session
	sess, err := session.LoadSession(host)
	if err != nil || sess == nil {
		return
	}

	// Apply authorization if available
	if sess.Authorization != "" {
		req.Header.Set("Authorization", sess.Authorization)
	}

	// Apply cookies
	for name, value := range sess.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: value,
		})
	}

	fmt.Fprintf(os.Stderr, "Using session for %s\n", host)
}

// executeWithRedirects executes the request with redirect handling.
func (e *Executor) executeWithRedirects(req *http.Request, plan *planner.ExecutionPlan) (*http.Response, []string, error) {
	maxRedirects := 5
	var redirectTrace []string

	// Determine redirect policy based on verb
	shouldFollow := false
	isWriteVerb := plan.Method == "POST" || plan.Method == "PUT" || plan.Method == "PATCH" || plan.Method == "DELETE"

	if plan.Follow == "smart" {
		// Smart follow: only follow 307/308 for write verbs
		shouldFollow = true
	} else {
		// Default: read, save, and authenticate follow, write verbs don't
		if plan.Verb == types.VerbRead || plan.Verb == types.VerbSave || plan.Verb == types.VerbAuthenticate {
			shouldFollow = true
		} else if isWriteVerb {
			shouldFollow = false
		} else {
			// Other verbs (watch, inspect) don't follow by default
			shouldFollow = false
		}
	}

	if !shouldFollow {
		resp, err := e.client.Do(req)
		if err == nil && isWriteVerb && (resp.StatusCode == 301 || resp.StatusCode == 302 || resp.StatusCode == 303) {
			redirectTrace = append(redirectTrace, fmt.Sprintf("Advisory: %d redirect for write verb, not following", resp.StatusCode))
		}
		return resp, redirectTrace, err
	}

	// Follow redirects
	redirects := 0
	client := *e.client
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if redirects >= maxRedirects {
			return fmt.Errorf("stopped after %d redirects", maxRedirects)
		}

		// For smart follow with write verbs, only follow 307/308
		if plan.Follow == "smart" && isWriteVerb {
			statusCode := via[len(via)-1].Response.StatusCode
			if statusCode != 307 && statusCode != 308 {
				return fmt.Errorf("write verb: not following %d redirect (use 307/308)", statusCode)
			}
		}

		redirects++
		statusCode := via[len(via)-1].Response.StatusCode
		redirectTrace = append(redirectTrace, fmt.Sprintf("→ %d %s %s", statusCode, req.Method, req.URL.String()))
		return nil
	}

	resp, err := client.Do(req)
	return resp, redirectTrace, err
}

// executeWithRedirectsCapturingCookies executes the request with redirect handling,
// capturing Set-Cookie headers from all redirect responses.
// This is needed for authenticate verb to capture cookies from redirect responses.
func (e *Executor) executeWithRedirectsCapturingCookies(req *http.Request, plan *planner.ExecutionPlan) (*http.Response, []string, []string, error) {
	maxRedirects := 5
	var redirectTrace []string
	var allSetCookies []string

	// Use the client's CookieJar to automatically handle cookies during redirects
	// Then extract cookies from the jar after redirects complete
	client := *e.client
	client.CheckRedirect = nil // Disable automatic redirect following

	originalURL := req.URL
	for i := 0; i < maxRedirects; i++ {
		resp, err := client.Do(req)
		if err != nil {
			return nil, redirectTrace, allSetCookies, err
		}

		// Capture Set-Cookie headers from this response
		setCookies := resp.Header.Values("Set-Cookie")
		allSetCookies = append(allSetCookies, setCookies...)

		// Check if this is a redirect
		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			location := resp.Header.Get("Location")
			if location == "" {
				resp.Body.Close()
				return resp, redirectTrace, allSetCookies, nil
			}

			// Parse the location URL
			redirectURL, err := url.Parse(location)
			if err != nil {
				resp.Body.Close()
				return resp, redirectTrace, allSetCookies, fmt.Errorf("invalid redirect URL: %w", err)
			}

			// Make location absolute if needed
			if !redirectURL.IsAbs() {
				baseURL, _ := url.Parse(req.URL.String())
				redirectURL = baseURL.ResolveReference(redirectURL)
			}

			// Create new request for redirect
			redirectTrace = append(redirectTrace, fmt.Sprintf("→ %d %s %s", resp.StatusCode, req.Method, redirectURL.String()))
			resp.Body.Close()

			// Create new request for redirect (preserve method for 307/308)
			method := req.Method
			if resp.StatusCode == 301 || resp.StatusCode == 302 || resp.StatusCode == 303 {
				// Change to GET for these redirects
				method = "GET"
			}

			newReq, err := http.NewRequest(method, redirectURL.String(), nil)
			if err != nil {
				return nil, redirectTrace, allSetCookies, fmt.Errorf("failed to create redirect request: %w", err)
			}

			// Copy headers from original request
			for k, v := range req.Header {
				newReq.Header[k] = v
			}
			req = newReq
			continue
		}

		// Not a redirect, return the response
		// Extract cookies from CookieJar for the original host
		if client.Jar != nil {
			hostURL, err := url.Parse(originalURL.Scheme + "://" + originalURL.Host)
			if err == nil {
				jarCookies := client.Jar.Cookies(hostURL)
				// Convert jar cookies to Set-Cookie format strings
				for _, cookie := range jarCookies {
					allSetCookies = append(allSetCookies, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
				}
			}
		}
		return resp, redirectTrace, allSetCookies, nil
	}

	return nil, redirectTrace, allSetCookies, fmt.Errorf("stopped after %d redirects", maxRedirects)
}

// readAndDecompress reads and decompresses the response body.
func (e *Executor) readAndDecompress(resp *http.Response) ([]byte, bool, error) {
	var reader io.Reader = resp.Body
	decompressed := false

	// Check if compressed - handle case-insensitive and multiple encodings
	encoding := resp.Header.Get("Content-Encoding")
	if encoding != "" {
		// Split by comma and check each encoding (case-insensitive)
		// Process encodings in reverse order (last encoding applied first)
		encodings := strings.Split(encoding, ",")
		for i := len(encodings) - 1; i >= 0; i-- {
			enc := strings.TrimSpace(strings.ToLower(encodings[i]))
			if enc == "gzip" {
				gzipReader, err := gzip.NewReader(reader)
				if err != nil {
					return nil, false, fmt.Errorf("failed to create gzip reader: %w", err)
				}
				defer gzipReader.Close()
				reader = gzipReader
				decompressed = true
			} else if enc == "br" {
				// brotli.Reader implements io.Reader
				reader = brotli.NewReader(reader)
				decompressed = true
			}
		}
	}

	data, err := io.ReadAll(reader)
	return data, decompressed, err
}

// runExpectChecks runs expectation checks on the response.
func (e *Executor) runExpectChecks(resp *http.Response, body []byte, checks []types.ExpectCheck) error {
	for _, check := range checks {
		if err := e.runExpectCheck(resp, body, check); err != nil {
			return err
		}
	}
	return nil
}

// runExpectCheck runs a single expectation check.
func (e *Executor) runExpectCheck(resp *http.Response, body []byte, check types.ExpectCheck) error {
	switch check.Type {
	case "status":
		expected := check.Value
		actual := fmt.Sprintf("%d", resp.StatusCode)
		if actual != expected {
			return fmt.Errorf("expected status %s, got %s", expected, actual)
		}

	case "header":
		actual := resp.Header.Get(check.Name)
		if actual != check.Value {
			return fmt.Errorf("expected header %s=%s, got %s", check.Name, check.Value, actual)
		}

	case "contains":
		if !strings.Contains(string(body), check.Value) {
			return fmt.Errorf("expected body to contain %q", check.Value)
		}

	case "jsonpath":
		// Simple JSON path extraction (basic implementation)
		var data interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
		// TODO: Implement proper JSONPath evaluation
		// For now, just check if JSON is valid
		_ = data

	case "matches":
		matched, err := regexp.MatchString(check.Regex, string(body))
		if err != nil {
			return fmt.Errorf("invalid regex: %w", err)
		}
		if !matched {
			return fmt.Errorf("body does not match regex %q", check.Regex)
		}

	default:
		return fmt.Errorf("unknown expect check type: %s", check.Type)
	}

	return nil
}

// printMeta prints metadata to stderr.
func (e *Executor) printMeta(resp *http.Response, url string, bodySize int, decompressed bool) {
	fmt.Fprintf(os.Stderr, "HTTP %d\n", resp.StatusCode)
	fmt.Fprintf(os.Stderr, "URL: %s\n", url)
	fmt.Fprintf(os.Stderr, "Size: %d bytes\n", bodySize)
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		fmt.Fprintf(os.Stderr, "Content-Type: %s\n", ct)
	}
}

// writeOutput formats and writes output to stdout.
func (e *Executor) writeOutput(body []byte, output *planner.OutputPlan) error {
	if output == nil {
		// Default: raw output
		_, err := os.Stdout.Write(body)
		return err
	}

	switch output.Format {
	case "json":
		// Pretty print JSON
		var data interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			// Not JSON, output as-is
			_, err := os.Stdout.Write(body)
			return err
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(data)

	case "text":
		// Output as text
		_, err := os.Stdout.Write(body)
		return err

	case "raw":
		// Raw output
		_, err := os.Stdout.Write(body)
		return err

	case "csv":
		// CSV output (basic - would need proper CSV parsing)
		_, err := os.Stdout.Write(body)
		return err

	default:
		// Default: raw
		_, err := os.Stdout.Write(body)
		return err
	}
}

// saveToFile saves the response body to a file.
func (e *Executor) saveToFile(body io.Reader, destination string) error {
	// Create directory if needed
	dir := filepath.Dir(destination)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Create file
	file, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy response body to file
	_, err = io.Copy(file, body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// getInsecureTLSConfig returns an insecure TLS config.
func getInsecureTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
	}
}

// ExecuteWithResponse executes an HTTP request and returns the response body as a string.
// This is useful for TUI mode where we need to capture and format the response.
func (e *Executor) ExecuteWithResponse(plan *planner.ExecutionPlan) (string, error) {
	// Build request URL with query parameters
	reqURL, err := e.buildURL(plan)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Build request body
	body, contentType, err := e.buildBody(plan)
	if err != nil {
		return "", fmt.Errorf("failed to build body: %w", err)
	}

	// Create request
	req, err := http.NewRequest(plan.Method, reqURL, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	e.setHeaders(req, plan, contentType)

	// Set cookies
	e.setCookies(req, plan)

	// Execute request
	resp, _, err := e.executeWithRedirects(req, plan)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read and decompress response body
	bodyBytes, _, err := e.readAndDecompress(resp)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(bodyBytes), nil
}
