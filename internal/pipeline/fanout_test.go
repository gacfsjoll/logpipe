package pipeline

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/user/logpipe/internal/parser"
)

// mockSink records every entry written to it and can be configured to error.
type mockSink struct {
	mu      sync.Mutex
	entries []parser.Entry
	errOn   int // return error on the n-th call (1-based); 0 = never
	calls   int
}

func (m *mockSink) Write(_ context.Context, e parser.Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	if m.errOn > 0 && m.calls == m.errOn {
		return errors.New("mock sink error")
	}
	m.entries = append(m.entries, e)
	return nil
}

func (m *mockSink) snapshot() []parser.Entry {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]parser.Entry, len(m.entries))
	copy(out, m.entries)
	return out
}

// fanOut sends entry to every sink and returns the first non-nil error.
func TestFanOut_AllSinksReceiveEntry(t *testing.T) {
	a := &mockSink{}
	b := &mockSink{}
	sinks := []interface {
		Write(context.Context, parser.Entry) error
	}{a, b}

	entry := parser.Entry{Message: "hello", Level: "info"}
	ctx := context.Background()

	var firstErr error
	for _, s := range sinks {
		if err := s.Write(ctx, entry); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if firstErr != nil {
		t.Fatalf("unexpected error: %v", firstErr)
	}
	if got := a.snapshot(); len(got) != 1 || got[0].Message != "hello" {
		t.Errorf("sink a: expected 1 entry with message 'hello', got %+v", got)
	}
	if got := b.snapshot(); len(got) != 1 || got[0].Message != "hello" {
		t.Errorf("sink b: expected 1 entry with message 'hello', got %+v", got)
	}
}

func TestFanOut_SinkErrorDoesNotBlockOthers(t *testing.T) {
	bad := &mockSink{errOn: 1}
	good := &mockSink{}

	entry := parser.Entry{Message: "world", Level: "warn"}
	ctx := context.Background()

	sinks := []interface {
		Write(context.Context, parser.Entry) error
	}{bad, good}

	var errCount int
	for _, s := range sinks {
		if err := s.Write(ctx, entry); err != nil {
			errCount++
		}
	}

	if errCount != 1 {
		t.Errorf("expected 1 error, got %d", errCount)
	}
	if got := good.snapshot(); len(got) != 1 {
		t.Errorf("good sink should have received the entry, got %+v", got)
	}
}
