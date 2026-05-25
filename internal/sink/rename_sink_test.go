package sink_test

import (
	"errors"
	"testing"

	"logpipe/internal/parser"
	"logpipe/internal/sink"
)

func renameEntry() parser.Entry {
	return parser.Entry{
		"level":   "info",
		"message": "hello",
		"host":    "srv-1",
	}
}

func TestNewRenameSink_NilInnerReturnsError(t *testing.T) {
	_, err := sink.NewRenameSink(nil, []sink.RenameRule{{From: "a", To: "b"}})
	if err == nil {
		t.Fatal("expected error for nil inner sink")
	}
}

func TestNewRenameSink_EmptyRulesReturnsError(t *testing.T) {
	_, err := sink.NewRenameSink(&captureSink{}, []sink.RenameRule{})
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestNewRenameSink_EmptyFromReturnsError(t *testing.T) {
	_, err := sink.NewRenameSink(&captureSink{}, []sink.RenameRule{{From: "", To: "b"}})
	if err == nil {
		t.Fatal("expected error when From is empty")
	}
}

func TestNewRenameSink_EmptyToReturnsError(t *testing.T) {
	_, err := sink.NewRenameSink(&captureSink{}, []sink.RenameRule{{From: "a", To: ""}})
	if err == nil {
		t.Fatal("expected error when To is empty")
	}
}

func TestRenameSink_Write_RenamesField(t *testing.T) {
	cap := &captureSink{}
	s, err := sink.NewRenameSink(cap, []sink.RenameRule{{From: "message", To: "msg"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Write(renameEntry()); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if _, ok := cap.last["message"]; ok {
		t.Error("original key 'message' should have been removed")
	}
	if cap.last["msg"] != "hello" {
		t.Errorf("expected msg=hello, got %v", cap.last["msg"])
	}
}

func TestRenameSink_Write_UnaffectedFieldsPreserved(t *testing.T) {
	cap := &captureSink{}
	s, _ := sink.NewRenameSink(cap, []sink.RenameRule{{From: "host", To: "hostname"}})
	s.Write(renameEntry()) //nolint:errcheck
	if cap.last["level"] != "info" {
		t.Errorf("expected level=info, got %v", cap.last["level"])
	}
	if cap.last["hostname"] != "srv-1" {
		t.Errorf("expected hostname=srv-1, got %v", cap.last["hostname"])
	}
}

func TestRenameSink_Write_MissingSourceKeySkipped(t *testing.T) {
	cap := &captureSink{}
	s, _ := sink.NewRenameSink(cap, []sink.RenameRule{{From: "nonexistent", To: "x"}})
	s.Write(renameEntry()) //nolint:errcheck
	if _, ok := cap.last["x"]; ok {
		t.Error("key 'x' should not exist when source key is absent")
	}
}

func TestRenameSink_Write_DoesNotMutateOriginal(t *testing.T) {
	cap := &captureSink{}
	s, _ := sink.NewRenameSink(cap, []sink.RenameRule{{From: "level", To: "severity"}})
	orig := renameEntry()
	s.Write(orig) //nolint:errcheck
	if _, ok := orig["level"]; !ok {
		t.Error("original entry should not be mutated")
	}
}

func TestRenameSink_Write_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("inner error")
	s, _ := sink.NewRenameSink(&errorSink{err: sentinel}, []sink.RenameRule{{From: "level", To: "severity"}})
	if err := s.Write(renameEntry()); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}
