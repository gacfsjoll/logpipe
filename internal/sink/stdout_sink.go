package sink

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"logpipe/internal/parser"
)

// StdoutSink writes parsed log entries as JSON lines to an io.Writer (defaults to os.Stdout).
type StdoutSink struct {
	out io.Writer
}

// NewStdoutSink creates a StdoutSink that writes to os.Stdout.
func NewStdoutSink() *StdoutSink {
	return &StdoutSink{out: os.Stdout}
}

// NewStdoutSinkWithWriter creates a StdoutSink that writes to the provided writer.
// Useful for testing.
func NewStdoutSinkWithWriter(w io.Writer) *StdoutSink {
	return &StdoutSink{out: w}
}

// Write serialises the log entry as a single JSON line and writes it to the sink's writer.
func (s *StdoutSink) Write(entry parser.Entry) error {
	b, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("stdout sink: marshal entry: %w", err)
	}
	_, err = fmt.Fprintln(s.out, string(b))
	if err != nil {
		return fmt.Errorf("stdout sink: write: %w", err)
	}
	return nil
}
