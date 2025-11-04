package tests

import (
	"net/http"
	"net/http/httptest"
)

// TestServer provides a test HTTP server for integration tests.
type TestServer struct {
	*httptest.Server
}

// NewTestServer creates a new test HTTP server with common endpoints.
func NewTestServer() *TestServer {
	mux := http.NewServeMux()

	// JSON endpoint
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]`))
	})

	// Plain text endpoint
	mux.HandleFunc("/text", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, world!"))
	})

	// Redirect endpoint
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/users", http.StatusFound)
	})

	// SSE endpoint
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: event1\n\n"))
	})

	server := httptest.NewServer(mux)
	return &TestServer{Server: server}
}

