package sink_test

import (
	"errors"
	"testing"
	"time"

	"logpipe/internal/parser"
	"logpipe/internal/schema"
	"logpipe/internal/sink"
)

func schemaValidateEntry() parser.Entry {
	return parser.Entry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "hello",
		Fields:    map[string]any{"service": "api", "code": float64(200)},
	}
}

var validSchemaValidateRules = []schema.Rule{
	{Field: "service", Required: true, Type: "string"},
	{Field: "code", Required: true, Type: "number"},
}

func TestNewSchemaSink_NilInnerReturnsError(t *testing.T) {
	_, err := sink.NewSchemaSink(nil, validSchemaValidateRules)
	if err == nil {
		t.Fatal("expected error for nil inner sink")
	}
}

func TestNewSchemaSink_EmptyRulesReturnsError(t *testing.T) {
	inner := sink.NewStdoutSink()
	_, err := sink.NewSchemaSink(inner, nil)
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestSchemaSink_Write_ValidEntryForwarded(t *testing.T) {
	var received []parser.Entry
	inner := &captureSink{fn: func(e parser.Entry) error {
		received = append(received, e)
		return nil
	}}
	s, err := sink.NewSchemaSink(inner, validSchemaValidateRules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entry := schemaValidateEntry()
	if err := s.Write(entry); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected 1 entry forwarded, got %d", len(received))
	}
}

func TestSchemaSink_Write_InvalidEntryDropped(t *testing.T) {
	var received []parser.Entry
	inner := &captureSink{fn: func(e parser.Entry) error {
		received = append(received, e)
		return nil
	}}
	s, err := sink.NewSchemaSink(inner, validSchemaValidateRules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bad := schemaValidateEntry()
	delete(bad.Fields, "service") // violates required rule
	if err := s.Write(bad); err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if len(received) != 0 {
		t.Fatalf("expected 0 forwarded entries, got %d", len(received))
	}
}

func TestSchemaSink_Write_InnerErrorPropagates(t *testing.T) {
	sentinel := errors.New("inner failure")
	inner := &captureSink{fn: func(e parser.Entry) error { return sentinel }}
	s, err := sink.NewSchemaSink(inner, validSchemaValidateRules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Write(schemaValidateEntry()); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
