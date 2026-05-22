package redact_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/yourorg/logpipe/internal/parser"
	"github.com/yourorg/logpipe/internal/redact"
)

func baseEntry() parser.Entry {
	return parser.Entry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "test",
		Extra: map[string]any{
			"password": "s3cr3t",
			"email":    "user@example.com",
			"host":     "localhost",
		},
	}
}

func TestNew_EmptyRulesReturnsError(t *testing.T) {
	_, err := redact.New(nil, "")
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestNew_EmptyFieldReturnsError(t *testing.T) {
	_, err := redact.New([]redact.Rule{{Field: ""}}, "")
	if err == nil {
		t.Fatal("expected error for empty field name")
	}
}

func TestApply_RedactsFieldUnconditionally(t *testing.T) {
	r, _ := redact.New([]redact.Rule{{Field: "password"}}, "")
	out := r.Apply(baseEntry())
	if out.Extra["password"] != "[REDACTED]" {
		t.Fatalf("expected [REDACTED], got %v", out.Extra["password"])
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	r, _ := redact.New([]redact.Rule{{Field: "password"}}, "")
	orig := baseEntry()
	r.Apply(orig)
	if orig.Extra["password"] != "s3cr3t" {
		t.Fatal("original entry was mutated")
	}
}

func TestApply_PatternMatchRedacts(t *testing.T) {
	pat := regexp.MustCompile(`@`)
	r, _ := redact.New([]redact.Rule{{Field: "email", Pattern: pat}}, "***")
	out := r.Apply(baseEntry())
	if out.Extra["email"] != "***" {
		t.Fatalf("expected ***, got %v", out.Extra["email"])
	}
}

func TestApply_PatternNoMatchKeepsValue(t *testing.T) {
	pat := regexp.MustCompile(`@`)
	r, _ := redact.New([]redact.Rule{{Field: "host", Pattern: pat}}, "")
	out := r.Apply(baseEntry())
	if out.Extra["host"] != "localhost" {
		t.Fatalf("expected localhost, got %v", out.Extra["host"])
	}
}

func TestApply_MissingFieldIsNoOp(t *testing.T) {
	r, _ := redact.New([]redact.Rule{{Field: "token"}}, "")
	out := r.Apply(baseEntry())
	if _, exists := out.Extra["token"]; exists {
		t.Fatal("did not expect token key to appear")
	}
}

func TestApply_CustomMask(t *testing.T) {
	r, _ := redact.New([]redact.Rule{{Field: "password"}}, "<hidden>")
	out := r.Apply(baseEntry())
	if out.Extra["password"] != "<hidden>" {
		t.Fatalf("expected <hidden>, got %v", out.Extra["password"])
	}
}
