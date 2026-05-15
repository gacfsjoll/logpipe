package watcher_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"logpipe/internal/watcher"
)

// TestWatcher_MultiplePathsIndependent ensures that events for distinct paths
// do not interfere with one another.
func TestWatcher_MultiplePathsIndependent(t *testing.T) {
	tmp := t.TempDir()
	pathA := filepath.Join(tmp, "a.log")
	pathB := filepath.Join(tmp, "b.log")

	// Only create pathA.
	f, err := os.Create(pathA)
	if err != nil {
		t.Fatalf("create pathA: %v", err)
	}
	f.Close()

	w := watcher.New([]string{pathA, pathB}, pollInterval)
	w.Start()
	defer w.Stop()

	seen := map[string]watcher.State{}
	deadline := time.After(300 * time.Millisecond)
	for len(seen) < 2 {
		select {
		case ev := <-w.Events():
			seen[ev.Path] = ev.State
		case <-deadline:
			t.Fatalf("timed out; only received events for: %v", seen)
		}
	}

	if seen[pathA] != watcher.StateAvailable {
		t.Errorf("pathA: expected StateAvailable, got %v", seen[pathA])
	}
	if seen[pathB] != watcher.StateMissing {
		t.Errorf("pathB: expected StateMissing, got %v", seen[pathB])
	}
}

// TestWatcher_StopIsIdempotent verifies that calling Stop does not panic.
func TestWatcher_StopIsIdempotent(t *testing.T) {
	w := watcher.New([]string{"/nonexistent"}, pollInterval)
	w.Start()
	w.Stop() // should not panic or block
}
