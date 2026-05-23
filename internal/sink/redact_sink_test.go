package sink_test

import (
	"errors"
	"testing"

	"logpipe/internal/parser"
	"logpipe/internal/sink"
)

func redactEntry() parser.Entry {
	return parser.Entry{
		"level":    "info",
		"message":  "user login",
		"password": "s3cr3t",
		"token":    "abc123",
	}
}

func TestNewRedactSink_NilInnerReturnsError(t *testing.T) {
	_, err := sink.NewRedactSink(nil, map[string]string{"password": "[REDACTED]"})
	if err == nil {
		t.Fatal("expected error for nil inner sink")
	}
}

func TestNewRedactSink_EmptyRulesReturnsError(t *testing.T) {
	cap := &captureSink{}
	_, err := sink.NewRedactSink(cap, map[string]string{})
	if err == nil {
		t.Fatal("expected error for empty rules map")
	}
}

func TestRedactSink_Write_RedactsConfiguredFields(t *testing.T) {
	cap := &captureSink{}
	s, err := sink.NewRedactSink(cap, map[string]string{
		"password": "[REDACTED]",
		"token":    "***",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := s.Write(redactEntry()); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	if len(cap.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(cap.entries))
	}
	got := cap.entries[0]
	if got["password"] != "[REDACTED]" {
		t.Errorf("password field: got %q, want %q", got["password"], "[REDACTED]")
	}
	if got["token"] != "***" {
		t.Errorf("token field: got %q, want %q", got["token"], "***")
	}
}

func TestRedactSink_Write_DoesNotMutateOriginal(t *testing.T) {
	cap := &captureSink{}
	s, _ := sink.NewRedactSink(cap, map[string]string{"password": "[REDACTED]"})

	orig := redactEntry()
	_ = s.Write(orig)

	if orig["password"] != "s3cr3t" {
		t.Error("original entry was mutated by Write")
	}
}

func TestRedactSink_Write_PreservesUnredactedFields(t *testing.T) {
	cap := &captureSink{}
	s, _ := sink.NewRedactSink(cap, map[string]string{"password": "[REDACTED]"})
	_ = s.Write(redactEntry())

	got := cap.entries[0]
	if got["level"] != "info" {
		t.Errorf("level field altered: got %q", got["level"])
	}
	if got["message"] != "user login" {
		t.Errorf("message field altered: got %q", got["message"])
	}
}

func TestRedactSink_Write_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("downstream failure")
	errSink := &errorSink{err: sentinel}
	s, _ := sink.NewRedactSink(errSink, map[string]string{"password": "[REDACTED]"})

	if err := s.Write(redactEntry()); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}
