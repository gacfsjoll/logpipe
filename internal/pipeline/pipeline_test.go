package pipeline_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"logpipe/internal/config"
	"logpipe/internal/pipeline"
)

func TestPipeline_ForwardsLogLine(t *testing.T) {
	// Temporary log file.
	f, err := os.CreateTemp(t.TempDir(), "test-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	received := make(chan map[string]interface{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		received <- body
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.Config{
		Sources: []config.Source{
			{Path: f.Name()},
		},
		Sinks: []config.SinkConfig{
			{Type: "http", URL: server.URL},
		},
	}

	p, err := pipeline.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go p.Run(ctx)

	// Give the tailer a moment to start.
	time.Sleep(100 * time.Millisecond)

	_, _ = f.WriteString(`{"timestamp":"2024-01-01T00:00:00Z","level":"info","message":"hello pipeline"}` + "\n")

	select {
	case body := <-received:
		if body["message"] != "hello pipeline" {
			t.Errorf("unexpected message: %v", body["message"])
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for log entry to be forwarded")
	}
}

func TestPipeline_InvalidSinkURL(t *testing.T) {
	cfg := &config.Config{
		Sources: []config.Source{{Path: "/tmp/x.log"}},
		Sinks:   []config.SinkConfig{{Type: "http", URL: "://bad url"}},
	}
	_, err := pipeline.New(cfg)
	if err == nil {
		t.Fatal("expected error for invalid sink URL, got nil")
	}
}
