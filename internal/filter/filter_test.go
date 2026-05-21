package filter_test

import (
	"testing"
	"time"

	"logpipe/internal/filter"
	"logpipe/internal/parser"
)

func entry(fields map[string]any) parser.Entry {
	return parser.Entry{Timestamp: time.Now(), Fields: fields}
}

func TestNew_EmptyFieldReturnsError(t *testing.T) {
	_, err := filter.New([]filter.Rule{{Field: "", Values: []string{"info"}}})
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestNew_NoValuesReturnsError(t *testing.T) {
	_, err := filter.New([]filter.Rule{{Field: "level", Values: nil}})
	if err == nil {
		t.Fatal("expected error for empty values")
	}
}

func TestKeep_NoRules_KeepsAll(t *testing.T) {
	f, _ := filter.New(nil)
	if !f.Keep(entry(map[string]any{"level": "debug"})) {
		t.Fatal("expected entry to be kept when no rules configured")
	}
}

func TestKeep_MatchingRule_KeepsEntry(t *testing.T) {
	f, _ := filter.New([]filter.Rule{
		{Field: "level", Values: []string{"info", "warn"}},
	})
	if !f.Keep(entry(map[string]any{"level": "info"})) {
		t.Fatal("expected info entry to be kept")
	}
}

func TestKeep_NonMatchingValue_DropsEntry(t *testing.T) {
	f, _ := filter.New([]filter.Rule{
		{Field: "level", Values: []string{"error"}},
	})
	if f.Keep(entry(map[string]any{"level": "debug"})) {
		t.Fatal("expected debug entry to be dropped")
	}
}

func TestKeep_MissingField_DropsEntry(t *testing.T) {
	f, _ := filter.New([]filter.Rule{
		{Field: "service", Values: []string{"api"}},
	})
	if f.Keep(entry(map[string]any{"level": "info"})) {
		t.Fatal("expected entry without 'service' field to be dropped")
	}
}

func TestKeep_CaseInsensitiveMatch(t *testing.T) {
	f, _ := filter.New([]filter.Rule{
		{Field: "level", Values: []string{"INFO"}},
	})
	if !f.Keep(entry(map[string]any{"level": "info"})) {
		t.Fatal("expected case-insensitive match to keep entry")
	}
}

func TestKeep_MultipleRules_AllMustMatch(t *testing.T) {
	f, _ := filter.New([]filter.Rule{
		{Field: "level", Values: []string{"error"}},
		{Field: "service", Values: []string{"payments"}},
	})
	// only level matches
	if f.Keep(entry(map[string]any{"level": "error", "service": "auth"})) {
		t.Fatal("expected entry to be dropped when second rule fails")
	}
	// both match
	if !f.Keep(entry(map[string]any{"level": "error", "service": "payments"})) {
		t.Fatal("expected entry to be kept when all rules pass")
	}
}

func TestStats_ReturnsRuleCount(t *testing.T) {
	f, _ := filter.New([]filter.Rule{
		{Field: "level", Values: []string{"info"}},
		{Field: "env", Values: []string{"prod"}},
	})
	if f.Stats() != 2 {
		t.Fatalf("expected 2 rules, got %d", f.Stats())
	}
}
