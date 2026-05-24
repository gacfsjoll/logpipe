package sink_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"logpipe/internal/parser"
	"logpipe/internal/sink"
)

func batchEntry(msg string) parser.Entry {
	return parser.Entry{
		"message": msg,
		"level":   "info",
	}
}

func TestNewBatchedSink_NilInnerReturnsError(t *testing.T) {
	_, err := sink.NewBatchedSink(nil, sink.BatchedSinkConfig{MaxSize: 10})
	if err == nil {
		t.Fatal("expected error for nil inner sink")
	}
}

func TestNewBatchedSink_InvalidConfigReturnsError(t *testing.T) {
	_, err := sink.NewBatchedSink(&sink.StdoutSink{}, sink.BatchedSinkConfig{MaxSize: 0})
	if err == nil {
		t.Fatal("expected error for zero MaxSize")
	}
}

func TestBatchedSink_FlushesOnMaxSize(t *testing.T) {
	var mu sync.Mutex
	var received []parser.Entry

	collector := &collectorSink{fn: func(e parser.Entry) error {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, e)
		return nil
	}}

	bs, err := sink.NewBatchedSink(collector, sink.BatchedSinkConfig{
		MaxSize:       3,
		FlushInterval: 10 * time.Second, // long interval; flush via size
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go bs.Start(ctx)

	for i := 0; i < 3; i++ {
		if err := bs.Write(batchEntry("msg")); err != nil {
			t.Fatalf("Write error: %v", err)
		}
	}

	// Flush should have occurred synchronously when the third entry was added.
	mu.Lock()
	got := len(received)
	mu.Unlock()

	if got != 3 {
		t.Fatalf("expected 3 entries flushed, got %d", got)
	}
}

func TestBatchedSink_FlushesOnInterval(t *testing.T) {
	var mu sync.Mutex
	var received []parser.Entry

	collector := &collectorSink{fn: func(e parser.Entry) error {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, e)
		return nil
	}}

	bs, err := sink.NewBatchedSink(collector, sink.BatchedSinkConfig{
		MaxSize:       100,
		FlushInterval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go bs.Start(ctx)

	_ = bs.Write(batchEntry("hello"))

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	got := len(received)
	mu.Unlock()

	if got != 1 {
		t.Fatalf("expected 1 entry after interval flush, got %d", got)
	}
}

// collectorSink is a minimal Sink that records written entries via a callback.
type collectorSink struct {
	fn func(parser.Entry) error
}

func (c *collectorSink) Write(e parser.Entry) error { return c.fn(e) }
