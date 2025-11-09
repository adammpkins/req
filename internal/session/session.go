// Package session manages HTTP sessions (cookies and tokens) per host.
package session

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Session represents a stored session for a host.
type Session struct {
	Host          string            `json:"host"`
	Cookies       map[string]string `json:"cookies,omitempty"`
	Authorization string            `json:"authorization,omitempty"` // Bearer token
}

var (
	stateDir     string
	stateDirOnce sync.Once
)

// getStateDir returns the user state directory for storing sessions.
func getStateDir() string {
	stateDirOnce.Do(func() {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory
			stateDir = ".req"
			return
		}
		stateDir = filepath.Join(homeDir, ".config", "req")
	})
	return stateDir
}

// ensureStateDir ensures the state directory exists with proper permissions.
func ensureStateDir() error {
	dir := getStateDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}
	return nil
}

// getSessionPath returns the file path for a host's session.
func getSessionPath(host string) (string, error) {
	if err := ensureStateDir(); err != nil {
		return "", err
	}
	// Sanitize host name for filename
	safeHost := strings.ReplaceAll(host, ":", "_")
	safeHost = strings.ReplaceAll(safeHost, "/", "_")
	return filepath.Join(getStateDir(), fmt.Sprintf("session_%s.json", safeHost)), nil
}

// LoadSession loads a session for the given host.
func LoadSession(host string) (*Session, error) {
	path, err := getSessionPath(host)
	if err != nil {
		return nil, err
	}

	// Check file permissions - refuse to load if group or world readable
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No session exists
		}
		return nil, fmt.Errorf("failed to stat session file: %w", err)
	}

	mode := info.Mode().Perm()
	// Check if group or others have read permission (044, 004, or any combination)
	if mode&0044 != 0 {
		return nil, fmt.Errorf("session file %s has insecure permissions (%s): group or world readable, refusing to load", path, mode.String())
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read session: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session: %w", err)
	}

	return &session, nil
}

// SaveSession saves a session for the given host.
func SaveSession(session *Session) error {
	path, err := getSessionPath(session.Host)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Write with strict permissions (0600)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write session: %w", err)
	}

	return nil
}

// DeleteSession deletes a session for the given host.
func DeleteSession(host string) error {
	path, err := getSessionPath(host)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// ExtractHost extracts the host from a URL.
func ExtractHost(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	return u.Host, nil
}

// UpdateSessionFromResponse updates a session from an HTTP response.
// Captures Set-Cookie headers and access_token from JSON body.
func UpdateSessionFromResponse(host string, setCookies []string, body []byte) (*Session, error) {
	session, err := LoadSession(host)
	if err != nil {
		return nil, err
	}

	if session == nil {
		session = &Session{
			Host:    host,
			Cookies: make(map[string]string),
		}
	}

	// Parse Set-Cookie headers
	for _, cookieHeader := range setCookies {
		// Simple cookie parsing (just get name=value part)
		parts := strings.Split(cookieHeader, ";")
		if len(parts) > 0 {
			cookiePart := strings.TrimSpace(parts[0])
			eqIdx := strings.Index(cookiePart, "=")
			if eqIdx > 0 {
				name := cookiePart[:eqIdx]
				value := cookiePart[eqIdx+1:]
				session.Cookies[name] = value
			}
		}
	}

	// Try to extract access_token from JSON body
	if len(body) > 0 {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(body, &jsonData); err == nil {
			if token, ok := jsonData["access_token"].(string); ok && token != "" {
				session.Authorization = "Bearer " + token
			}
		}
	}

	return session, nil
}

// ListSessions lists all stored sessions.
func ListSessions() ([]string, error) {
	dir := getStateDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read state directory: %w", err)
	}

	var hosts []string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "session_") && strings.HasSuffix(entry.Name(), ".json") {
			// Extract host from filename
			host := strings.TrimPrefix(entry.Name(), "session_")
			host = strings.TrimSuffix(host, ".json")
			host = strings.ReplaceAll(host, "_", ":")
			hosts = append(hosts, host)
		}
	}

	return hosts, nil
}

// RedactSession creates a redacted version of a session for display.
func RedactSession(session *Session) *Session {
	redacted := &Session{
		Host:          session.Host,
		Cookies:       make(map[string]string),
		Authorization: "",
	}

	// Redact cookies (show only names)
	for name := range session.Cookies {
		redacted.Cookies[name] = "***"
	}

	// Redact authorization
	if session.Authorization != "" {
		redacted.Authorization = "Bearer ***"
	}

	return redacted
}

