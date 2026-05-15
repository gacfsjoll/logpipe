package watcher_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"logpipe/internal/watcher"
)

const pollInterval = 20 * time.Millisecond

func TestWatcher_DetectsAvailableFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "app.log")

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}
	f.Close()

	w := watcher.New([]string{path}, pollInterval)
	w.Start()
	defer w.Stop()

	select {
	case ev := <-w.Events():
		if ev.Path != path {
			t.Errorf("expected path %q, got %q", path, ev.Path)
		}
		if ev.State != watcher.StateAvailable {
			t.Errorf("expected StateAvailable, got %v", ev.State)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for available event")
	}
}

func TestWatcher_DetectsMissingFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "missing.log")

	w := watcher.New([]string{path}, pollInterval)
	w.Start()
	defer w.Stop()

	select {
	case ev := <-w.Events():
		if ev.State != watcher.StateMissing {
			t.Errorf("expected StateMissing, got %v", ev.State)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for missing event")
	}
}

func TestWatcher_EmitsTransitionEvents(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "rotate.log")

	w := watcher.New([]string{path}, pollInterval)
	w.Start()
	defer w.Stop()

	// First event: missing
	drainOne(t, w.Events(), watcher.StateMissing, 200*time.Millisecond)

	// Create the file → available
	f, _ := os.Create(path)
	f.Close()
	drainOne(t, w.Events(), watcher.StateAvailable, 200*time.Millisecond)

	// Remove the file → missing again
	os.Remove(path)
	drainOne(t, w.Events(), watcher.StateMissing, 200*time.Millisecond)
}

func drainOne(t *testing.T, ch <-chan watcher.Event, want watcher.State, timeout time.Duration) {
	t.Helper()
	select {
	case ev := <-ch:
		if ev.State != want {
			t.Errorf("expected state %v, got %v", want, ev.State)
		}
	case <-time.After(timeout):
		t.Fatalf("timed out waiting for state %v", want)
	}
}
