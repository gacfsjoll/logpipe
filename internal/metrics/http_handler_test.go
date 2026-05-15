package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_ReturnsJSON(t *testing.T) {
	m := New()
	m.IncLinesRead()
	m.IncLinesRead()
	m.IncLinesParsed()
	m.IncLinesDropped()
	m.IncSinkErrors()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	m.Handler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("unexpected Content-Type: %s", contentType)
	}

	var resp SnapshotResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.LinesRead != 2 {
		t.Errorf("expected LinesRead=2, got %d", resp.LinesRead)
	}
	if resp.LinesParsed != 1 {
		t.Errorf("expected LinesParsed=1, got %d", resp.LinesParsed)
	}
	if resp.LinesDropped != 1 {
		t.Errorf("expected LinesDropped=1, got %d", resp.LinesDropped)
	}
	if resp.SinkErrors != 1 {
		t.Errorf("expected SinkErrors=1, got %d", resp.SinkErrors)
	}
	if resp.UptimeSeconds < 0 {
		t.Errorf("expected non-negative uptime, got %f", resp.UptimeSeconds)
	}
}

func TestHandler_ZeroCounters(t *testing.T) {
	m := New()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	m.Handler()(rec, req)

	var resp SnapshotResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.LinesRead != 0 || resp.LinesParsed != 0 || resp.LinesDropped != 0 || resp.SinkErrors != 0 {
		t.Errorf("expected all zero counters, got %+v", resp)
	}
}
