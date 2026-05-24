package sink_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"logpipe/internal/parser"
	"logpipe/internal/sink"
)

func bpEntry() parser.Entry {
	return parser.Entry{"level": "info", "msg": "hello"}
}

func TestNewBackpressureSink_NilInnerReturnsError(t *testing.T) {
	_, err := sink.NewBackpressureSink(nil, sink.BackpressureConfig{Limit: 2})
	if err == nil {
		t.Fatal("expected error for nil inner sink")
	}
}

func TestNewBackpressureSink_ZeroLimitReturnsError(t *testing.T) {
	inner := sink.NewStdoutSinkWithWriter(nil)
	_, err := sink.NewBackpressureSink(inner, sink.BackpressureConfig{Limit: 0})
	if err == nil {
		t.Fatal("expected error for limit=0")
	}
}

func TestBackpressureSink_Write_ForwardsToInner(t *testing.T) {
	var captured []parser.Entry
	var mu sync.Mutex
	collector := &callbackSink{fn: func(e parser.Entry) error {
		mu.Lock()
		captured = append(captured, e)
		mu.Unlock()
		return nil
	}}

	s, err := sink.NewBackpressureSink(collector, sink.BackpressureConfig{Limit: 4})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := s.Write(context.Background(), bpEntry()); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(captured) != 1 {
		t.Fatalf("expected 1 entry forwarded, got %d", len(captured))
	}
}

func TestBackpressureSink_Write_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("inner failure")
	collector := &callbackSink{fn: func(_ parser.Entry) error { return sentinel }}

	s, _ := sink.NewBackpressureSink(collector, sink.BackpressureConfig{Limit: 1})
	err := s.Write(context.Background(), bpEntry())
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestBackpressureSink_Write_CancelledContextReturnsError(t *testing.T) {
	blocking := &callbackSink{fn: func(_ parser.Entry) error {
		time.Sleep(10 * time.Second)
		return nil
	}}

	s, _ := sink.NewBackpressureSink(blocking, sink.BackpressureConfig{Limit: 1})

	// Fill the one available slot in a goroutine.
	started := make(chan struct{})
	go func() {
		close(started)
		_ = s.Write(context.Background(), bpEntry()) //nolint:errcheck
	}()
	<-started
	time.Sleep(5 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := s.Write(ctx, bpEntry())
	if err == nil {
		t.Fatal("expected error when context cancelled while waiting for token")
	}
}

func TestBackpressureSink_ConcurrentWritesRespectLimit(t *testing.T) {
	const limit = 3
	var inFlight int64
	var overLimit int64

	collector := &callbackSink{fn: func(_ parser.Entry) error {
		current := atomic.AddInt64(&inFlight, 1)
		if current > limit {
			atomic.AddInt64(&overLimit, 1)
		}
		time.Sleep(5 * time.Millisecond)
		atomic.AddInt64(&inFlight, -1)
		return nil
	}}

	s, _ := sink.NewBackpressureSink(collector, sink.BackpressureConfig{Limit: limit})

	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.Write(context.Background(), bpEntry()) //nolint:errcheck
		}()
	}
	wg.Wait()

	if overLimit > 0 {
		t.Fatalf("concurrency limit exceeded: %d writes ran concurrently above limit %d", overLimit, limit)
	}
}
