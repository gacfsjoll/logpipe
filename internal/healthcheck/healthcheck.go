// Package healthcheck provides a simple HTTP health endpoint that reports
// the overall status of logpipe, including source watcher states and sink
// reachability.
package healthcheck

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Status represents the health of a single named component.
type Status struct {
	Name    string `json:"name"`
	Healthy bool   `json:"healthy"`
	Detail  string `json:"detail,omitempty"`
}

// Report is the top-level health response returned by the HTTP handler.
type Report struct {
	Healthy    bool      `json:"healthy"`
	CheckedAt  time.Time `json:"checked_at"`
	Components []Status  `json:"components"`
}

// Checker holds a registry of named check functions.
type Checker struct {
	mu     sync.RWMutex
	checks map[string]func() (bool, string)
}

// New returns an initialised Checker.
func New() *Checker {
	return &Checker{
		checks: make(map[string]func() (bool, string)),
	}
}

// Register adds or replaces a named check. The function should return
// (true, "") when healthy, or (false, reason) when unhealthy.
func (c *Checker) Register(name string, fn func() (bool, string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks[name] = fn
}

// Deregister removes a named check. It is a no-op if the name is not found.
func (c *Checker) Deregister(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.checks, name)
}

// Run evaluates all registered checks and returns a Report.
func (c *Checker) Run() Report {
	c.mu.RLock()
	defer c.mu.RUnlock()

	report := Report{
		Healthy:   true,
		CheckedAt: time.Now().UTC(),
	}

	for name, fn := range c.checks {
		ok, detail := fn()
		if !ok {
			report.Healthy = false
		}
		report.Components = append(report.Components, Status{
			Name:    name,
			Healthy: ok,
			Detail:  detail,
		})
	}

	return report
}

// Handler returns an http.HandlerFunc that serves the health report as JSON.
// It responds with 200 when all components are healthy, 503 otherwise.
func (c *Checker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		report := c.Run()
		w.Header().Set("Content-Type", "application/json")
		if !report.Healthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_ = json.NewEncoder(w).Encode(report)
	}
}
