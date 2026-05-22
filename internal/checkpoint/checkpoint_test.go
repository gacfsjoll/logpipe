package checkpoint_test

import (
	"os"
	"path/filepath"
	"testing"

	"logpipe/internal/checkpoint"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "offsets.json")
}

func TestNew_EmptyWhenFileAbsent(t *testing.T) {
	s, err := checkpoint.New(tempPath(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := s.Get("/var/log/app.log"); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestSetAndGet(t *testing.T) {
	s, _ := checkpoint.New(tempPath(t))
	s.Set("/var/log/app.log", 1024)
	if got := s.Get("/var/log/app.log"); got != 1024 {
		t.Fatalf("expected 1024, got %d", got)
	}
}

func TestFlushAndReload(t *testing.T) {
	path := tempPath(t)
	s, _ := checkpoint.New(path)
	s.Set("/a.log", 512)
	s.Set("/b.log", 2048)

	if err := s.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	s2, err := checkpoint.New(path)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}
	if got := s2.Get("/a.log"); got != 512 {
		t.Errorf("/a.log: expected 512, got %d", got)
	}
	if got := s2.Get("/b.log"); got != 2048 {
		t.Errorf("/b.log: expected 2048, got %d", got)
	}
}

func TestNew_CorruptFileReturnsError(t *testing.T) {
	path := tempPath(t)
	_ = os.WriteFile(path, []byte("not json{"), 0o644)
	_, err := checkpoint.New(path)
	if err == nil {
		t.Fatal("expected error for corrupt file, got nil")
	}
}

func TestSnapshot_IsIndependentCopy(t *testing.T) {
	s, _ := checkpoint.New(tempPath(t))
	s.Set("/c.log", 99)

	snap := s.Snapshot()
	s.Set("/c.log", 200)

	if snap["/c.log"] != 99 {
		t.Errorf("snapshot was mutated: got %d", snap["/c.log"])
	}
}

func TestFlush_AtomicRename(t *testing.T) {
	path := tempPath(t)
	s, _ := checkpoint.New(path)
	s.Set("/d.log", 777)
	if err := s.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	// tmp file must not linger
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Error("tmp file still present after flush")
	}
}
