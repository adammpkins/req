// Package runtime executes HTTP requests based on execution plans.
package runtime

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adammpkins/req/internal/planner"
)

// Executor executes HTTP requests.
type Executor struct {
	client *http.Client
}

// NewExecutor creates a new executor.
func NewExecutor(plan *planner.ExecutionPlan) (*Executor, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if plan.Timeout != nil {
		client.Timeout = *plan.Timeout
	}

	// Configure TLS if insecure
	if plan.Insecure {
		transport := &http.Transport{
			TLSClientConfig: getInsecureTLSConfig(),
		}
		client.Transport = transport
	}

	// Configure proxy if specified
	if plan.Proxy != "" {
		proxyURL, err := url.Parse(plan.Proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		if plan.Insecure {
			transport.TLSClientConfig = getInsecureTLSConfig()
		}
		client.Transport = transport
	}

	return &Executor{client: client}, nil
}

// Execute executes an HTTP request based on the plan.
func (e *Executor) Execute(plan *planner.ExecutionPlan) error {
	// Build request URL with query parameters
	reqURL := plan.URL
	if len(plan.QueryParams) > 0 {
		u, err := url.Parse(plan.URL)
		if err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}
		q := u.Query()
		for k, v := range plan.QueryParams {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
		reqURL = u.String()
	}

	// Create request
	var body io.Reader
	if plan.Body != nil && plan.Body.Content != "" {
		body = strings.NewReader(plan.Body.Content)
	}

	req, err := http.NewRequest(plan.Method, reqURL, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for k, v := range plan.Headers {
		req.Header.Set(k, v)
	}

	// Set content type for body
	if plan.Body != nil {
		switch plan.Body.Type {
		case "json":
			req.Header.Set("Content-Type", "application/json")
		case "form":
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	// Execute request
	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d %s", resp.StatusCode, resp.Status)
	}

	// Handle output based on plan
	if plan.Output != nil && plan.Output.Destination != "" {
		// Save to file
		return e.saveToFile(resp.Body, plan.Output.Destination)
	}

	// Default: write to stdout
	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}

// ExecuteWithResponse executes an HTTP request and returns the response body as a string.
// This is useful for TUI mode where we need to capture and format the response.
func (e *Executor) ExecuteWithResponse(plan *planner.ExecutionPlan) (string, error) {
	// Build request URL with query parameters
	reqURL := plan.URL
	if len(plan.QueryParams) > 0 {
		u, err := url.Parse(plan.URL)
		if err != nil {
			return "", fmt.Errorf("invalid URL: %w", err)
		}
		q := u.Query()
		for k, v := range plan.QueryParams {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
		reqURL = u.String()
	}

	// Create request
	var body io.Reader
	if plan.Body != nil && plan.Body.Content != "" {
		body = strings.NewReader(plan.Body.Content)
	}

	req, err := http.NewRequest(plan.Method, reqURL, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for k, v := range plan.Headers {
		req.Header.Set(k, v)
	}

	// Set content type for body
	if plan.Body != nil {
		switch plan.Body.Type {
		case "json":
			req.Header.Set("Content-Type", "application/json")
		case "form":
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	// Execute request
	resp, err := e.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d %s", resp.StatusCode, resp.Status)
	}

	// Handle output based on plan
	if plan.Output != nil && plan.Output.Destination != "" {
		// Save to file - return success message
		err := e.saveToFile(resp.Body, plan.Output.Destination)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("File saved to %s", plan.Output.Destination), nil
	}

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(bodyBytes), nil
}

// saveToFile saves the response body to a file.
func (e *Executor) saveToFile(body io.Reader, destination string) error {
	// Determine the actual file path
	filePath := destination

	// If destination is a directory, extract filename from URL
	if isDirectory(destination) {
		// Extract filename from URL (should be handled by planner, but handle here as fallback)
		// For now, return error - planner should handle this
		return fmt.Errorf("destination is a directory: %s (use a full path like /tmp/file.zip)", destination)
	}

	// Create directory if needed (for paths like /tmp/file.zip)
	dir := filepath.Dir(filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Create file
	file, err := os.Create(filePath)
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

// isDirectory checks if a path is a directory.
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// getInsecureTLSConfig returns an insecure TLS config.
func getInsecureTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
	}
}

