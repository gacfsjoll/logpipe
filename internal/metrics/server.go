package metrics

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

const shutdownTimeout = 5 * time.Second

// Server wraps an HTTP server that exposes the metrics endpoint.
type Server struct {
	http *http.Server
}

// NewServer creates a metrics HTTP server listening on the given address
// (e.g. ":9091") and registers the /metrics route against m.
func NewServer(addr string, m *Metrics) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", m.Handler())

	return &Server{
		http: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
	}
}

// Start begins listening in a background goroutine. It returns an error if the
// server fails to bind immediately (detected via a short-lived ListenAndServe).
func (s *Server) Start() {
	go func() {
		log.Printf("metrics: listening on %s", s.http.Addr)
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("metrics: server error: %v", err)
		}
	}()
}

// Shutdown gracefully stops the HTTP server within the shutdown timeout.
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := s.http.Shutdown(ctx); err != nil {
		return fmt.Errorf("metrics: shutdown error: %w", err)
	}
	return nil
}
