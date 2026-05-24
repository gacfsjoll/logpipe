package sink_test

import (
	"errors"
	"testing"

	"logpipe/internal/parser"
	"logpipe/internal/router"
	"logpipe/internal/sink"
)

type routerCaptureSink struct {
	entries []parser.Entry
	fail    bool
}

func (r *routerCaptureSink) Write(e parser.Entry) error {
	if r.fail {
		return errors.New("sink error")
	}
	r.entries = append(r.entries, e)
	return nil
}

func routerEntry(level string) parser.Entry {
	return parser.Entry{
		"level":   level,
		"message": "test",
	}
}

func TestNewRouterSink_NoRoutesReturnsError(t *testing.T) {
	_, err := sink.NewRouterSink(nil)
	if err == nil {
		t.Fatal("expected error for empty routes")
	}
}

func TestNewRouterSink_InvalidRouteReturnsError(t *testing.T) {
	routes := []router.Route{
		{Field: "", Values: []string{"error"}, Sink: &routerCaptureSink{}},
	}
	_, err := sink.NewRouterSink(routes)
	if err == nil {
		t.Fatal("expected error for empty field name")
	}
}

func TestRouterSink_Write_MatchingRouteForwarded(t *testing.T) {
	errSink := &routerCaptureSink{}
	routes := []router.Route{
		{Field: "level", Values: []string{"error"}, Sink: errSink},
	}
	s, err := sink.NewRouterSink(routes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := s.Write(routerEntry("error")); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	if len(errSink.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(errSink.entries))
	}
}

func TestRouterSink_Write_NonMatchingEntryDropped(t *testing.T) {
	errSink := &routerCaptureSink{}
	routes := []router.Route{
		{Field: "level", Values: []string{"error"}, Sink: errSink},
	}
	s, _ := sink.NewRouterSink(routes)

	// "info" does not match "error" route
	if err := s.Write(routerEntry("info")); err != nil {
		t.Fatalf("unexpected error for non-matching entry: %v", err)
	}
	if len(errSink.entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(errSink.entries))
	}
}

func TestRouterSink_Write_PropagatesInnerError(t *testing.T) {
	failSink := &routerCaptureSink{fail: true}
	routes := []router.Route{
		{Field: "level", Values: []string{"warn"}, Sink: failSink},
	}
	s, _ := sink.NewRouterSink(routes)

	err := s.Write(routerEntry("warn"))
	if err == nil {
		t.Fatal("expected error from failing inner sink")
	}
}
