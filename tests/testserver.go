// Package tests provides a local HTTP test server for integration tests.
package tests

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

// TestServer provides a test HTTP server with various endpoints.
type TestServer struct {
	server *httptest.Server
	mux    *http.ServeMux
}

// NewTestServer creates a new test server.
func NewTestServer() *TestServer {
	mux := http.NewServeMux()
	ts := &TestServer{
		mux: mux,
	}

	// Echo endpoint - returns request details
	mux.HandleFunc("/echo", ts.handleEcho)

	// Headers endpoint - returns request headers
	mux.HandleFunc("/headers", ts.handleHeaders)

	// Cookies endpoint - sets and returns cookies
	mux.HandleFunc("/cookies", ts.handleCookies)

	// Query endpoint - returns query parameters
	mux.HandleFunc("/query", ts.handleQuery)

	// Gzip endpoint - returns gzipped content
	mux.HandleFunc("/gzip", ts.handleGzip)

	// Redirect endpoints
	mux.HandleFunc("/redirect/301", ts.handleRedirect(301))
	mux.HandleFunc("/redirect/302", ts.handleRedirect(302))
	mux.HandleFunc("/redirect/303", ts.handleRedirect(303))
	mux.HandleFunc("/redirect/307", ts.handleRedirect(307))
	mux.HandleFunc("/redirect/308", ts.handleRedirect(308))
	mux.HandleFunc("/final", ts.handleFinal)

	// Multipart endpoint - echoes multipart form data
	mux.HandleFunc("/multipart", ts.handleMultipart)

	// JSON endpoint - returns JSON
	mux.HandleFunc("/json", ts.handleJSON)

	// Status endpoint - returns specific status code
	mux.HandleFunc("/status/", ts.handleStatus)

	ts.server = httptest.NewServer(mux)
	return ts
}

// URL returns the base URL of the test server.
func (ts *TestServer) URL() string {
	return ts.server.URL
}

// Close shuts down the test server.
func (ts *TestServer) Close() {
	ts.server.Close()
}

// handleEcho returns request details as JSON.
func (ts *TestServer) handleEcho(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
  "method": "%s",
  "url": "%s",
  "headers": %s,
  "body": "%s"
}`, r.Method, r.URL.String(), ts.headersJSON(r), ts.bodyString(r))
}

// handleHeaders returns request headers as JSON.
func (ts *TestServer) handleHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"headers": %s}`, ts.headersJSON(r))
}

// handleCookies sets and returns cookies.
func (ts *TestServer) handleCookies(w http.ResponseWriter, r *http.Request) {
	// Set a test cookie
	http.SetCookie(w, &http.Cookie{
		Name:  "session",
		Value: "test-session-value",
		Path:  "/",
	})

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"cookies": %s}`, ts.cookiesJSON(r))
}

// handleQuery returns query parameters as JSON.
func (ts *TestServer) handleQuery(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"query": %s}`, ts.queryJSON(r))
}

// handleGzip returns gzipped content.
func (ts *TestServer) handleGzip(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Type", "text/plain")

	gz := gzip.NewWriter(w)
	defer gz.Close()
	fmt.Fprint(gz, "This is gzipped content")
}

// handleFinal is the final destination for redirects.
func (ts *TestServer) handleFinal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "Final destination")
}

// handleRedirect returns a redirect response.
func (ts *TestServer) handleRedirect(code int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", ts.server.URL+"/final")
		w.WriteHeader(code)
	}
}

// handleMultipart echoes multipart form data.
func (ts *TestServer) handleMultipart(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error parsing multipart: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
  "form": %s,
  "files": %s
}`, ts.formJSON(r), ts.filesJSON(r))
}

// handleJSON returns JSON data.
func (ts *TestServer) handleJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"message": "Hello, World!", "items": [1, 2, 3]}`)
}

// handleStatus returns a specific HTTP status code.
func (ts *TestServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	// Extract status code from path
	path := strings.TrimPrefix(r.URL.Path, "/status/")
	var code int
	fmt.Sscanf(path, "%d", &code)
	if code == 0 {
		code = 200
	}
	w.WriteHeader(code)
	fmt.Fprintf(w, "Status %d", code)
}

// Helper functions

func (ts *TestServer) headersJSON(r *http.Request) string {
	headers := make(map[string]string)
	for k, v := range r.Header {
		headers[k] = strings.Join(v, ", ")
	}
	return fmt.Sprintf(`{"%s": "%s"}`, "User-Agent", headers["User-Agent"])
}

func (ts *TestServer) cookiesJSON(r *http.Request) string {
	cookies := make([]string, 0)
	for _, cookie := range r.Cookies() {
		cookies = append(cookies, fmt.Sprintf(`"%s=%s"`, cookie.Name, cookie.Value))
	}
	return "[" + strings.Join(cookies, ", ") + "]"
}

func (ts *TestServer) queryJSON(r *http.Request) string {
	params := make([]string, 0)
	for k, v := range r.URL.Query() {
		for _, val := range v {
			params = append(params, fmt.Sprintf(`"%s": "%s"`, k, val))
		}
	}
	return "{" + strings.Join(params, ", ") + "}"
}

func (ts *TestServer) bodyString(r *http.Request) string {
	body, _ := io.ReadAll(r.Body)
	return strings.ReplaceAll(string(body), "\"", "\\\"")
}

func (ts *TestServer) formJSON(r *http.Request) string {
	form := make([]string, 0)
	for k, v := range r.MultipartForm.Value {
		for _, val := range v {
			form = append(form, fmt.Sprintf(`"%s": "%s"`, k, val))
		}
	}
	return "{" + strings.Join(form, ", ") + "}"
}

func (ts *TestServer) filesJSON(r *http.Request) string {
	files := make([]string, 0)
	for k, v := range r.MultipartForm.File {
		for _, fh := range v {
			files = append(files, fmt.Sprintf(`"%s": "%s"`, k, fh.Filename))
		}
	}
	return "{" + strings.Join(files, ", ") + "}"
}
