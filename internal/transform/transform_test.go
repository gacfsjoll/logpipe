package transform_test

import (
	"testing"
	"time"

	"github.com/logpipe/logpipe/internal/parser"
	"github.com/logpipe/logpipe/internal/transform"
)

func baseEntry() *parser.Entry {
	return &parser.Entry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "hello",
		Fields:    map[string]any{"user": "alice", "token": "secret"},
	}
}

func TestApply_NoOps_ReturnsEntryUnchanged(t *testing.T) {
	e := baseEntry()
	out := transform.New().Apply(e)
	if out.Message != "hello" {
		t.Fatalf("expected message unchanged, got %q", out.Message)
	}
}

func TestAddField_InsertsValue(t *testing.T) {
	e := baseEntry()
	transform.New(transform.AddField("env", "prod")).Apply(e)
	if e.Fields["env"] != "prod" {
		t.Fatalf("expected env=prod, got %v", e.Fields["env"])
	}
}

func TestAddField_OverwritesExisting(t *testing.T) {
	e := baseEntry()
	transform.New(transform.AddField("user", "bob")).Apply(e)
	if e.Fields["user"] != "bob" {
		t.Fatalf("expected user=bob, got %v", e.Fields["user"])
	}
}

func TestRemoveField_DeletesKey(t *testing.T) {
	e := baseEntry()
	transform.New(transform.RemoveField("token")).Apply(e)
	if _, ok := e.Fields["token"]; ok {
		t.Fatal("expected token to be removed")
	}
}

func TestRemoveField_NoOpWhenAbsent(t *testing.T) {
	e := baseEntry()
	// should not panic
	transform.New(transform.RemoveField("nonexistent")).Apply(e)
}

func TestNormaliseLevel_UpperCases(t *testing.T) {
	e := baseEntry()
	transform.New(transform.NormaliseLevel()).Apply(e)
	if e.Level != "INFO" {
		t.Fatalf("expected INFO, got %q", e.Level)
	}
}

func TestRedactField_MasksValue(t *testing.T) {
	e := baseEntry()
	transform.New(transform.RedactField("token", "***")).Apply(e)
	if e.Fields["token"] != "***" {
		t.Fatalf("expected ***, got %v", e.Fields["token"])
	}
}

func TestRedactField_NoOpWhenKeyAbsent(t *testing.T) {
	e := baseEntry()
	transform.New(transform.RedactField("password", "***")).Apply(e)
	if _, ok := e.Fields["password"]; ok {
		t.Fatal("expected no password key to be inserted")
	}
}

func TestApply_ChainedOps(t *testing.T) {
	e := baseEntry()
	transform.New(
		transform.NormaliseLevel(),
		transform.AddField("env", "staging"),
		transform.RedactField("token", "[REDACTED]"),
	).Apply(e)

	if e.Level != "INFO" {
		t.Errorf("level: want INFO, got %q", e.Level)
	}
	if e.Fields["env"] != "staging" {
		t.Errorf("env: want staging, got %v", e.Fields["env"])
	}
	if e.Fields["token"] != "[REDACTED]" {
		t.Errorf("token: want [REDACTED], got %v", e.Fields["token"])
	}
}
