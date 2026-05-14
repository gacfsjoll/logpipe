package sink_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"logpipe/internal/parser"
	"logpipe/internal/sink"
)

func sampleEntry() parser.Entry {
	return parser.Entry{
		Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		Level:     "info",
		Message:   "test message",
		Fields:    map[string]any{"service": "api"},
	}
}

func TestHTTPSink_Write_Success(t *testing.T) {
	var received parser.Entry
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("unexpected Content-Type: %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := sink.NewHTTPSink(server.URL, nil)
	entry := sampleEntry()
	if err := s.Write(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Message != entry.Message {
		t.Errorf("expected message %q, got %q", entry.Message, received.Message)
	}
}

func TestHTTPSink_Write_CustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v := r.Header.Get("X-Api-Key"); v != "secret" {
			t.Errorf("expected header X-Api-Key=secret, got %q", v)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	s := sink.NewHTTPSink(server.URL, map[string]string{"X-Api-Key": "secret"})
	if err := s.Write(sampleEntry()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPSink_Write_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	s := sink.NewHTTPSink(server.URL, nil)
	if err := s.Write(sampleEntry()); err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestHTTPSink_Write_InvalidURL(t *testing.T) {
	s := sink.NewHTTPSink("http://127.0.0.1:0", nil)
	if err := s.Write(sampleEntry()); err == nil {
		t.Fatal("expected error for unreachable URL, got nil")
	}
}

func TestHTTPSink_Close(t *testing.T) {
	s := sink.NewHTTPSink("http://example.com", nil)
	if err := s.Close(); err != nil {
		t.Errorf("Close() returned unexpected error: %v", err)
	}
}
