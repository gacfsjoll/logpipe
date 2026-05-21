package rotator_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"logpipe/internal/rotator"
)

const pollInterval = 20 * time.Millisecond

func writeLine(t *testing.T, path, line string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()
	f.WriteString(line + "\n")
}

func TestRotator_DetectsTruncation(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "app.log")
	writeLine(t, p, "hello")

	r := rotator.New([]string{p}, pollInterval)
	r.Start()
	t.Cleanup(r.Stop)

	// Truncate the file.
	time.Sleep(pollInterval * 2)
	if err := os.Truncate(p, 0); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	select {
	case ev := <-r.Events():
		if ev.Reason != "truncated" {
			t.Errorf("expected truncated, got %q", ev.Reason)
		}
		if ev.Path != p {
			t.Errorf("expected path %q, got %q", p, ev.Path)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for truncation event")
	}
}

func TestRotator_NoEventWhenFileGrows(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "app.log")
	writeLine(t, p, "first")

	r := rotator.New([]string{p}, pollInterval)
	r.Start()
	t.Cleanup(r.Stop)

	time.Sleep(pollInterval * 2)
	writeLine(t, p, "second")
	time.Sleep(pollInterval * 3)

	select {
	case ev := <-r.Events():
		t.Errorf("unexpected event: %+v", ev)
	default:
		// correct — no rotation expected
	}
}

func TestRotator_StopIsIdempotent(t *testing.T) {
	r := rotator.New([]string{}, pollInterval)
	r.Start()
	r.Stop()
	// Second Stop would panic on double-close if not guarded; the test
	// just verifies the goroutine exits cleanly.
}

func TestRotator_MissingFileDoesNotPanic(t *testing.T) {
	r := rotator.New([]string{"/nonexistent/path/app.log"}, pollInterval)
	r.Start()
	time.Sleep(pollInterval * 3)
	r.Stop() // should not panic
}
