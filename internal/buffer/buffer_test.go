package buffer_test

import (
	"sync"
	"testing"
	"time"

	"logpipe/internal/buffer"
	"logpipe/internal/parser"
)

func entry(msg string) *parser.Entry {
	return &parser.Entry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   msg,
		Fields:    map[string]interface{}{},
	}
}

func TestNew_PanicOnZeroCap(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero capacity")
		}
	}()
	buffer.New(0)
}

func TestPushPop_FIFO(t *testing.T) {
	b := buffer.New(4)
	msgs := []string{"first", "second", "third"}
	for _, m := range msgs {
		if err := b.Push(entry(m)); err != nil {
			t.Fatalf("unexpected push error: %v", err)
		}
	}
	for _, want := range msgs {
		got := b.Pop()
		if got == nil {
			t.Fatal("expected entry, got nil")
		}
		if got.Message != want {
			t.Errorf("got %q, want %q", got.Message, want)
		}
	}
}

func TestPush_ReturnErrFullWhenAtCapacity(t *testing.T) {
	b := buffer.New(2)
	_ = b.Push(entry("a"))
	_ = b.Push(entry("b"))

	if err := b.Push(entry("c")); err != buffer.ErrFull {
		t.Errorf("expected ErrFull, got %v", err)
	}
}

func TestPop_ReturnsNilWhenEmpty(t *testing.T) {
	b := buffer.New(4)
	if got := b.Pop(); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestLen_TracksCount(t *testing.T) {
	b := buffer.New(8)
	for i := 0; i < 5; i++ {
		_ = b.Push(entry("x"))
	}
	if b.Len() != 5 {
		t.Errorf("expected len 5, got %d", b.Len())
	}
	b.Pop()
	if b.Len() != 4 {
		t.Errorf("expected len 4, got %d", b.Len())
	}
}

func TestConcurrentPushPop(t *testing.T) {
	b := buffer.New(512)
	var wg sync.WaitGroup
	const n = 200
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			for b.Push(entry("concurrent")) != nil {
				time.Sleep(time.Microsecond)
			}
		}
	}()
	go func() {
		defer wg.Done()
		received := 0
		for received < n {
			if b.Pop() != nil {
				received++
			} else {
				time.Sleep(time.Microsecond)
			}
		}
	}()
	wg.Wait()
	if b.Len() != 0 {
		t.Errorf("expected empty buffer after concurrent test, got len %d", b.Len())
	}
}
