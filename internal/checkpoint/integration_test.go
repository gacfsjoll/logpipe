package checkpoint_test

import (
	"path/filepath"
	"sync"
	"testing"

	"logpipe/internal/checkpoint"
)

// TestConcurrentSetAndFlush verifies that concurrent writers do not corrupt
// the in-memory state or the flushed file.
func TestConcurrentSetAndFlush(t *testing.T) {
	path := filepath.Join(t.TempDir(), "offsets.json")
	s, err := checkpoint.New(path)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(n int) {
			defer wg.Done()
			key := filepath.Join("/logs", string(rune('a'+n))+".log")
			for j := 0; j < 50; j++ {
				s.Set(key, int64(j))
				_ = s.Get(key)
			}
		}(i)
	}
	wg.Wait()

	if err := s.Flush(); err != nil {
		t.Fatalf("flush after concurrent writes: %v", err)
	}

	s2, err := checkpoint.New(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	snap := s2.Snapshot()
	if len(snap) != workers {
		t.Errorf("expected %d entries, got %d", workers, len(snap))
	}
}

// TestMultipleFlushesAreIdempotent ensures repeated flushes do not corrupt data.
func TestMultipleFlushesAreIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "offsets.json")
	s, _ := checkpoint.New(path)
	s.Set("/x.log", 42)

	for i := 0; i < 5; i++ {
		if err := s.Flush(); err != nil {
			t.Fatalf("flush %d: %v", i, err)
		}
	}

	s2, _ := checkpoint.New(path)
	if got := s2.Get("/x.log"); got != 42 {
		t.Errorf("expected 42 after repeated flushes, got %d", got)
	}
}
