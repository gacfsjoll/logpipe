// Package checkpoint persists the last-read byte offset for each tailed file
// so that logpipe can resume from where it left off after a restart.
package checkpoint

import (
	"encoding/json"
	"os"
	"sync"
)

// Offsets maps a file path to its last committed read offset.
type Offsets map[string]int64

// Store is a thread-safe, file-backed offset registry.
type Store struct {
	mu   sync.Mutex
	path string
	data Offsets
}

// New loads an existing checkpoint file at path, or returns an empty Store if
// the file does not yet exist. Any other I/O error is returned to the caller.
func New(path string) (*Store, error) {
	s := &Store{path: path, data: make(Offsets)}

	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &s.data); err != nil {
		return nil, err
	}
	return s, nil
}

// Get returns the stored offset for the given file path, or 0 if none exists.
func (s *Store) Get(filePath string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data[filePath]
}

// Set updates the in-memory offset for the given file path.
func (s *Store) Set(filePath string, offset int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[filePath] = offset
}

// Flush atomically writes all in-memory offsets to disk.
func (s *Store) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// Snapshot returns a copy of the current offsets map.
func (s *Store) Snapshot() Offsets {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(Offsets, len(s.data))
	for k, v := range s.data {
		out[k] = v
	}
	return out
}
