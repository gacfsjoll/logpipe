package batch_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"logpipe/internal/batch"
	"logpipe/internal/parser"
)

func makeEntry(msg string) parser.Entry {
	return parser.Entry{
		"message": msg,
		"level":   "info",
	}
}

func TestBatcher_FlushesOnMaxSize(t *testing.T) {
	var mu sync.Mutex
	var got [][]parser.Entry

	b := batch.New(3, time.Hour, func(entries []parser.Entry) {
		mu.Lock()
		got = append(got, entries)
		mu.Unlock()
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Run(ctx)

	b.Add(makeEntry("a"))
	b.Add(makeEntry("b"))
	b.Add(makeEntry("c")) // triggers size flush

	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 {
		t.Fatalf("expected 1 flush, got %d", len(got))
	}
	if len(got[0]) != 3 {
		t.Fatalf("expected 3 entries in batch, got %d", len(got[0]))
	}
}

func TestBatcher_FlushesOnInterval(t *testing.T) {
	var mu sync.Mutex
	var got [][]parser.Entry

	b := batch.New(100, 50*time.Millisecond, func(entries []parser.Entry) {
		mu.Lock()
		got = append(got, entries)
		mu.Unlock()
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Run(ctx)

	b.Add(makeEntry("x"))
	b.Add(makeEntry("y"))

	time.Sleep(120 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(got) == 0 {
		t.Fatal("expected at least one interval flush")
	}
	if len(got[0]) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got[0]))
	}
}

func TestBatcher_FlushesOnContextCancel(t *testing.T) {
	var mu sync.Mutex
	var got [][]parser.Entry

	b := batch.New(100, time.Hour, func(entries []parser.Entry) {
		mu.Lock()
		got = append(got, entries)
		mu.Unlock()
	})

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); b.Run(ctx) }()

	b.Add(makeEntry("final"))
	cancel()
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if len(got) == 0 {
		t.Fatal("expected flush on context cancel")
	}
}

func TestBatcher_EmptyBatchNotFlushed(t *testing.T) {
	flushCount := 0
	b := batch.New(5, 30*time.Millisecond, func(_ []parser.Entry) {
		flushCount++
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Run(ctx)

	time.Sleep(80 * time.Millisecond)

	if flushCount != 0 {
		t.Fatalf("expected no flushes for empty batcher, got %d", flushCount)
	}
}

func TestNew_PanicsOnInvalidArgs(t *testing.T) {
	noop := func(_ []parser.Entry) {}

	mustPanic := func(name string, fn func()) {
		t.Helper()
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("%s: expected panic", name)
			}
		}()
		fn()
	}

	mustPanic("zero maxSize", func() { batch.New(0, time.Second, noop) })
	mustPanic("zero interval", func() { batch.New(1, 0, noop) })
	mustPanic("nil flush", func() { batch.New(1, time.Second, nil) })
}
