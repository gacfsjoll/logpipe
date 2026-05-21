package healthcheck

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Server wraps an HTTP server that exposes the health endpoint.
type Server struct {
	addr    string
	checker *Checker
	httpSrv *http.Server
}

// NewServer creates a Server that will listen on addr (e.g. ":9091") and
// serve the health report at /healthz.
func NewServer(addr string, checker *Checker) *Server {
	mux := http.NewServeMux()
	s := &Server{
		addr:    addr,
		checker: checker,
	}
	mux.HandleFunc("/healthz", checker.Handler())
	s.httpSrv = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	return s
}

// Start begins listening in a background goroutine. It returns an error if
// the listener cannot be bound.
func (s *Server) Start() error {
	log.Printf("healthcheck: listening on %s", s.addr)
	errCh := make(chan error, 1)
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("healthcheck server: %w", err)
		}
	}()

	// Give the server a moment to fail fast on bad config (e.g. port in use).
	select {
	case err := <-errCh:
		return err
	case <-time.After(30 * time.Millisecond):
		return nil
	}
}

// Stop gracefully shuts down the server.
func (s *Server) Stop(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}
