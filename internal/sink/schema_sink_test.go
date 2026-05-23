package sink_test

import (
	"errors"
	"testing"
	"time"

	"logpipe/internal/parser"
	"logpipe/internal/schema"
	"logpipe/internal/sink"
)

func schemaEntry(fields map[string]any) parser.LogEntry {
	e := parser.LogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "schema test",
		Fields:    make(map[string]any),
	}
	for k, v := range fields {
		e.Fields[k] = v
	}
	return e
}

func validSchemaRules() []schema.Rule {
	return []schema.Rule{
		{Field: "request_id", Required: true, Type: "string"},
	}
}

func TestNewSchemaSink_NilInnerReturnsError(t *testing.T) {
	_, err := sink.NewSchemaSink(nil, validSchemaRules())
	if err == nil {
		t.Fatal("expected error for nil inner sink")
	}
}

func TestNewSchemaSink_EmptyRulesReturnsError(t *testing.T) {
	_, err := sink.NewSchemaSink(&captureSink{}, []schema.Rule{})
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestSchemaSink_Write_ValidEntryForwarded(t *testing.T) {
	cap := &captureSink{}
	s, err := sink.NewSchemaSink(cap, validSchemaRules())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry := schemaEntry(map[string]any{"request_id": "abc-123"})
	if err := s.Write(entry); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(cap.entries) != 1 {
		t.Fatalf("expected 1 forwarded entry, got %d", len(cap.entries))
	}
}

func TestSchemaSink_Write_InvalidEntryDropped(t *testing.T) {
	cap := &captureSink{}
	s, err := sink.NewSchemaSink(cap, validSchemaRules())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// entry missing required field "request_id"
	entry := schemaEntry(nil)
	if err := s.Write(entry); err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if len(cap.entries) != 0 {
		t.Fatalf("expected 0 forwarded entries, got %d", len(cap.entries))
	}
}

func TestSchemaSink_Write_InnerErrorPropagated(t *testing.T) {
	sentinel := errors.New("inner failure")
	errSink := &errorSink{err: sentinel}
	s, err := sink.NewSchemaSink(errSink, validSchemaRules())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry := schemaEntry(map[string]any{"request_id": "xyz"})
	if got := s.Write(entry); !errors.Is(got, sentinel) {
		t.Fatalf("expected sentinel error, got: %v", got)
	}
}
