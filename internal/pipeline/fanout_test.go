package pipeline

import (
	"errors"
	"testing"

	"logpipe/internal/parser"
)

// mockSink records entries written to it and optionally returns an error.
type mockSink struct {
	written []parser.Entry
	errOnWrite error
}

func (m *mockSink) Write(entry parser.Entry) error {
	m.written = append(m.written, entry)
	return m.errOnWrite
}

func sampleEntry() parser.Entry {
	return parser.Entry{
		Timestamp: "2024-01-01T00:00:00Z",
		Level:     "info",
		Message:   "hello fanout",
		Raw:       map[string]any{"service": "test"},
	}
}

func TestFanOut_AllSinksReceiveEntry(t *testing.T) {
	a := &mockSink{}
	b := &mockSink{}
	entry := sampleEntry()

	if err := fanOut(entry, []Sink{a, b}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(a.written) != 1 || a.written[0].Message != entry.Message {
		t.Errorf("sink A did not receive entry correctly")
	}
	if len(b.written) != 1 || b.written[0].Message != entry.Message {
		t.Errorf("sink B did not receive entry correctly")
	}
}

func TestFanOut_SinkErrorDoesNotBlockOthers(t *testing.T) {
	failing := &mockSink{errOnWrite: errors.New("write failed")}
	good := &mockSink{}
	entry := sampleEntry()

	err := fanOut(entry, []Sink{failing, good})
	if err == nil {
		t.Fatal("expected error from failing sink, got nil")
	}

	// The good sink should still have received the entry.
	if len(good.written) != 1 {
		t.Errorf("good sink should have received the entry despite other sink failing")
	}
}

func TestFanOut_NoSinks(t *testing.T) {
	entry := sampleEntry()
	if err := fanOut(entry, nil); err != nil {
		t.Fatalf("expected no error with empty sink list, got: %v", err)
	}
}

func TestFanOut_MultipleSinkErrors(t *testing.T) {
	a := &mockSink{errOnWrite: errors.New("err-a")}
	b := &mockSink{errOnWrite: errors.New("err-b")}
	entry := sampleEntry()

	err := fanOut(entry, []Sink{a, b})
	if err == nil {
		t.Fatal("expected combined error, got nil")
	}

	msg := err.Error()
	if !contains(msg, "err-a") || !contains(msg, "err-b") {
		t.Errorf("expected both errors in message, got: %s", msg)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		})())
}
