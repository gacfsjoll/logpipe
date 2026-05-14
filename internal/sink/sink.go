// Package sink provides implementations for forwarding parsed log entries
// to configurable output destinations.
package sink

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"logpipe/internal/parser"
)

// Sink defines the interface for log entry destinations.
type Sink interface {
	Write(entry parser.Entry) error
	Close() error
}

// HTTPSink forwards log entries as JSON POST requests to a remote endpoint.
type HTTPSink struct {
	URL     string
	Headers map[string]string
	client  *http.Client
}

// NewHTTPSink creates a new HTTPSink with the given URL and optional headers.
func NewHTTPSink(url string, headers map[string]string) *HTTPSink {
	return &HTTPSink{
		URL:     url,
		Headers: headers,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Write serializes the log entry and sends it to the configured HTTP endpoint.
func (s *HTTPSink) Write(entry parser.Entry) error {
	payload, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("sink: marshal entry: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.URL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("sink: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range s.Headers {
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("sink: http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("sink: unexpected status %d from %s", resp.StatusCode, s.URL)
	}
	return nil
}

// Close is a no-op for HTTPSink but satisfies the Sink interface.
func (s *HTTPSink) Close() error { return nil }
