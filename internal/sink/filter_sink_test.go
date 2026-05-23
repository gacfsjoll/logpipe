package sink_test

import (
	"errors"
	"testing"

	"logpipe/internal/filter"
	"logpipe/internal/parser"
	"logpipe/internal/sink"
)

func filterEntry(level, msg string) parser.Entry {
	return parser.Entry{
		"level":   level,
		"message": msg,
	}
}

func TestNewFilterSink_NilInnerReturnsError(t *testing.T) {
	rules := []filter.Rule{{Field: "level", Values: []string{"error"}}}
	_, err := sink.NewFilterSink(nil, rules)
	if err == nil {
		t.Fatal("expected error for nil inner sink")
	}
}

func TestNewFilterSink_InvalidRuleReturnsError(t *testing.T) {
	rules := []filter.Rule{{Field: "", Values: []string{"error"}}}
	_, err := sink.NewFilterSink(&captureSink{}, rules)
	if err == nil {
		t.Fatal("expected error for empty field name")
	}
}

func TestFilterSink_Write_PassingEntryForwarded(t *testing.T) {
	cap := &captureSink{}
	rules := []filter.Rule{{Field: "level", Values: []string{"error"}}}
	s, err := sink.NewFilterSink(cap, rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry := filterEntry("error", "something broke")
	if err := s.Write(entry); err != nil {
		t.Fatalf("Write returned unexpected error: %v", err)
	}
	if len(cap.entries) != 1 {
		t.Fatalf("expected 1 forwarded entry, got %d", len(cap.entries))
	}
}

func TestFilterSink_Write_NonMatchingEntryDropped(t *testing.T) {
	cap := &captureSink{}
	rules := []filter.Rule{{Field: "level", Values: []string{"error"}}}
	s, _ := sink.NewFilterSink(cap, rules)

	if err := s.Write(filterEntry("info", "all good")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cap.entries) != 0 {
		t.Fatalf("expected 0 forwarded entries, got %d", len(cap.entries))
	}
}

func TestFilterSink_Write_InnerErrorPropagated(t *testing.T) {
	sentinel := errors.New("inner failure")
	inner := &errorSink{err: sentinel}
	rules := []filter.Rule{{Field: "level", Values: []string{"error"}}}
	s, _ := sink.NewFilterSink(inner, rules)

	err := s.Write(filterEntry("error", "boom"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestFilterSink_Write_MultipleRulesAllMustMatch(t *testing.T) {
	cap := &captureSink{}
	rules := []filter.Rule{
		{Field: "level", Values: []string{"error"}},
		{Field: "service", Values: []string{"auth"}},
	}
	s, _ := sink.NewFilterSink(cap, rules)

	// matches only first rule
	s.Write(parser.Entry{"level": "error", "service": "billing"})
	if len(cap.entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(cap.entries))
	}

	// matches both rules
	s.Write(parser.Entry{"level": "error", "service": "auth"})
	if len(cap.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(cap.entries))
	}
}
