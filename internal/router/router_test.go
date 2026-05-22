package router_test

import (
	"testing"
	"time"

	"logpipe/internal/parser"
	"logpipe/internal/router"
	"logpipe/internal/sink"
)

func makeEntry(fields map[string]any) parser.Entry {
	return parser.Entry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "test",
		Fields:    fields,
	}
}

func TestNew_NoRoutesReturnsError(t *testing.T) {
	_, err := router.New(nil, nil)
	if err == nil {
		t.Fatal("expected error for empty routes")
	}
}

func TestNew_EmptyFieldReturnsError(t *testing.T) {
	s := sink.NewStdoutSinkWithWriter(nil)
	_, err := router.New([]router.Route{{Field: "", Values: []string{"info"}, Sink: s}}, nil)
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestNew_NoValuesReturnsError(t *testing.T) {
	s := sink.NewStdoutSinkWithWriter(nil)
	_, err := router.New([]router.Route{{Field: "level", Values: nil, Sink: s}}, nil)
	if err == nil {
		t.Fatal("expected error for nil values")
	}
}

func TestNew_NilSinkReturnsError(t *testing.T) {
	_, err := router.New([]router.Route{{Field: "level", Values: []string{"info"}, Sink: nil}}, nil)
	if err == nil {
		t.Fatal("expected error for nil sink")
	}
}

func TestRoute_MatchingSinkReceivesEntry(t *testing.T) {
	var buf bytes.Buffer
	s := sink.NewStdoutSinkWithWriter(&buf)
	r, err := router.New([]router.Route{
		{Field: "level", Values: []string{"error"}, Sink: s},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	entry := makeEntry(map[string]any{"level": "error"})
	if err := r.Route(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected sink to receive entry")
	}
}

func TestRoute_FallbackReceivesUnmatchedEntry(t *testing.T) {
	var matched, fallback bytes.Buffer
	sMatched := sink.NewStdoutSinkWithWriter(&matched)
	sFallback := sink.NewStdoutSinkWithWriter(&fallback)
	r, _ := router.New([]router.Route{
		{Field: "level", Values: []string{"error"}, Sink: sMatched},
	}, sFallback)
	entry := makeEntry(map[string]any{"level": "info"})
	r.Route(entry) //nolint:errcheck
	if matched.Len() != 0 {
		t.Error("matched sink should not receive entry")
	}
	if fallback.Len() == 0 {
		t.Error("fallback sink should receive entry")
	}
}

func TestRoute_NoFallbackDropsEntry(t *testing.T) {
	var buf bytes.Buffer
	s := sink.NewStdoutSinkWithWriter(&buf)
	r, _ := router.New([]router.Route{
		{Field: "level", Values: []string{"error"}, Sink: s},
	}, nil)
	entry := makeEntry(map[string]any{"level": "debug"})
	if err := r.Route(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Error("no sink should receive unmatched entry without fallback")
	}
}
