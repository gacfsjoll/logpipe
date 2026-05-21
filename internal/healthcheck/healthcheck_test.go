package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/logpipe/logpipe/internal/healthcheck"
)

func TestRun_AllHealthy(t *testing.T) {
	c := healthcheck.New()
	c.Register("source:app.log", func() (bool, string) { return true, "" })
	c.Register("sink:http", func() (bool, string) { return true, "" })

	report := c.Run()

	if !report.Healthy {
		t.Fatal("expected report to be healthy")
	}
	if len(report.Components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(report.Components))
	}
}

func TestRun_OneUnhealthy(t *testing.T) {
	c := healthcheck.New()
	c.Register("source:app.log", func() (bool, string) { return true, "" })
	c.Register("sink:http", func() (bool, string) { return false, "connection refused" })

	report := c.Run()

	if report.Healthy {
		t.Fatal("expected report to be unhealthy")
	}

	var found bool
	for _, comp := range report.Components {
		if comp.Name == "sink:http" {
			found = true
			if comp.Healthy {
				t.Error("expected sink:http component to be unhealthy")
			}
			if comp.Detail != "connection refused" {
				t.Errorf("unexpected detail: %s", comp.Detail)
			}
		}
	}
	if !found {
		t.Error("sink:http component not found in report")
	}
}

func TestHandler_Returns200WhenHealthy(t *testing.T) {
	c := healthcheck.New()
	c.Register("db", func() (bool, string) { return true, "" })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	c.Handler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var report healthcheck.Report
	if err := json.NewDecoder(rec.Body).Decode(&report); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !report.Healthy {
		t.Error("expected healthy report")
	}
}

func TestHandler_Returns503WhenUnhealthy(t *testing.T) {
	c := healthcheck.New()
	c.Register("sink:file", func() (bool, string) { return false, "disk full" })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	c.Handler()(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestRun_NoChecks(t *testing.T) {
	c := healthcheck.New()
	report := c.Run()

	if !report.Healthy {
		t.Error("empty checker should report healthy")
	}
	if report.CheckedAt.IsZero() {
		t.Error("checked_at should not be zero")
	}
}
