package sink

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"logpipe/internal/parser"
)

// FileSink writes parsed log entries as JSON lines to a file on disk.
type FileSink struct {
	mu   sync.Mutex
	file *os.File
	path string
}

// NewFileSink opens (or creates) the file at the given path for appending
// and returns a FileSink ready for use.
func NewFileSink(path string) (*FileSink, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("file sink: open %q: %w", path, err)
	}
	return &FileSink{file: f, path: path}, nil
}

// Write serialises entry to JSON and appends it as a single line to the file.
func (s *FileSink) Write(entry parser.Entry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("file sink: marshal: %w", err)
	}
	data = append(data, '\n')

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := s.file.Write(data); err != nil {
		return fmt.Errorf("file sink: write: %w", err)
	}
	return nil
}

// Close flushes and closes the underlying file.
func (s *FileSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.file.Close()
}

// Path returns the file path this sink writes to.
func (s *FileSink) Path() string { return s.path }
