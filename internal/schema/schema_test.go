package schema_test

import (
	"testing"
	"time"

	"github.com/logpipe/logpipe/internal/parser"
	"github.com/logpipe/logpipe/internal/schema"
)

func makeEntry(fields map[string]any) parser.Entry {
	return parser.Entry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "test",
		Fields:    fields,
	}
}

func TestNew_EmptyRulesReturnsError(t *testing.T) {
	_, err := schema.New(nil)
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestNew_EmptyFieldNameReturnsError(t *testing.T) {
	_, err := schema.New([]schema.FieldRule{{Name: ""}})
	if err == nil {
		t.Fatal("expected error for empty field name")
	}
}

func TestValidate_RequiredFieldPresent(t *testing.T) {
	v, _ := schema.New([]schema.FieldRule{{Name: "service", Required: true}})
	err := v.Validate(makeEntry(map[string]any{"service": "auth"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_RequiredFieldMissing(t *testing.T) {
	v, _ := schema.New([]schema.FieldRule{{Name: "service", Required: true}})
	err := v.Validate(makeEntry(map[string]any{}))
	if err == nil {
		t.Fatal("expected error for missing required field")
	}
}

func TestValidate_OptionalFieldAbsent_NoError(t *testing.T) {
	v, _ := schema.New([]schema.FieldRule{{Name: "trace_id", Required: false}})
	err := v.Validate(makeEntry(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_CorrectStringType(t *testing.T) {
	v, _ := schema.New([]schema.FieldRule{{Name: "env", Type: schema.FieldTypeString}})
	err := v.Validate(makeEntry(map[string]any{"env": "production"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_WrongTypeReturnsError(t *testing.T) {
	v, _ := schema.New([]schema.FieldRule{{Name: "latency", Type: schema.FieldTypeNumber}})
	err := v.Validate(makeEntry(map[string]any{"latency": "fast"}))
	if err == nil {
		t.Fatal("expected type mismatch error")
	}
}

func TestValidate_BooleanType(t *testing.T) {
	v, _ := schema.New([]schema.FieldRule{{Name: "ok", Type: schema.FieldTypeBoolean}})
	if err := v.Validate(makeEntry(map[string]any{"ok": true})); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := v.Validate(makeEntry(map[string]any{"ok": "yes"})); err == nil {
		t.Fatal("expected error for string in boolean field")
	}
}

func TestValidate_MultipleRules_FirstViolationReturned(t *testing.T) {
	v, _ := schema.New([]schema.FieldRule{
		{Name: "service", Required: true},
		{Name: "latency", Type: schema.FieldTypeNumber},
	})
	err := v.Validate(makeEntry(map[string]any{}))
	if err == nil {
		t.Fatal("expected error")
	}
}
