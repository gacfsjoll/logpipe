// Package tail provides functionality for tailing log files and emitting
// parsed JSON log entries to a channel for downstream processing.
package tail

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"go.uber.org/zap"
)

// Entry represents a single parsed log record read from a tailed file.
type Entry struct {
	// Source is the file path from which this entry was read.
	Source string
	// Fields contains the parsed JSON key-value pairs.
	Fields map[string]any
}

// Tailer tails a single file and sends parsed JSON entries to Out.
type Tailer struct {
	// Out receives successfully parsed log entries.
	Out chan<- Entry

	filePath string
	pollInterval time.Duration
	logger       *zap.Logger
}

// New creates a new Tailer for the given file path.
// pollInterval controls how often the file is checked for new data when EOF is reached.
func New(filePath string, out chan<- Entry, pollInterval time.Duration, logger *zap.Logger) *Tailer {
	if pollInterval <= 0 {
		pollInterval = 250 * time.Millisecond
	}
	return &Tailer{
		filePath:     filePath,
		Out:          out,
		pollInterval: pollInterval,
		logger:       logger,
	}
}

// Run opens the file, seeks to the end, and continuously reads new lines until
// ctx is cancelled. Malformed JSON lines are logged and skipped.
func (t *Tailer) Run(ctx context.Context) error {
	f, err := os.Open(t.filePath)
	if err != nil {
		return fmt.Errorf("tail: open %q: %w", t.filePath, err)
	}
	defer f.Close()

	// Seek to end so we only process new entries written after startup.
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("tail: seek %q: %w", t.filePath, err)
	}

	t.logger.Info("tailing file", zap.String("path", t.filePath))

	reader := bufio.NewReader(f)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// No new data yet; wait before retrying.
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(t.pollInterval):
				}
				continue
			}
			return fmt.Errorf("tail: read %q: %w", t.filePath, err)
		}

		fields := make(map[string]any)
		if jsonErr := json.Unmarshal([]byte(line), &fields); jsonErr != nil {
			t.logger.Warn("skipping non-JSON line",
				zap.String("path", t.filePath),
				zap.String("line", line),
				zap.Error(jsonErr),
			)
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case t.Out <- Entry{Source: t.filePath, Fields: fields}:
		}
	}
}
