package sink

import (
	"errors"
	"fmt"
	"strings"

	"logpipe/internal/parser"
)

// MultilineSink coalesces consecutive log entries whose message field contains
// a continuation marker (e.g. a stack-trace line starting with "\t" or "at ")
// into a single entry before forwarding to the inner sink.
//
// Flushing is triggered when a non-continuation line is received or when
// Flush is called explicitly.
type MultilineSink struct {
	inner      Sink
	field      string
	prefixes   []string
	pending    *parser.Entry
	accumLines []string
}

// NewMultilineSink creates a MultilineSink that merges continuation lines
// (identified by any of the given prefixes on the entry's field value) into
// the preceding entry.
func NewMultilineSink(inner Sink, field string, continuationPrefixes []string) (*MultilineSink, error) {
	if inner == nil {
		return nil, errors.New("multiline: inner sink must not be nil")
	}
	if field == "" {
		return nil, errors.New("multiline: field must not be empty")
	}
	if len(continuationPrefixes) == 0 {
		return nil, errors.New("multiline: at least one continuation prefix is required")
	}
	return &MultilineSink{
		inner:    inner,
		field:    field,
		prefixes: continuationPrefixes,
	}, nil
}

// Write accepts an entry. If its field value starts with a known continuation
// prefix, the value is appended to the pending entry. Otherwise any pending
// entry is flushed first and the new entry becomes pending.
func (m *MultilineSink) Write(entry parser.Entry) error {
	val, _ := entry.Fields[m.field].(string)
	if m.isContinuation(val) {
		if m.pending == nil {
			// No prior entry – treat as a standalone line.
			m.pending = cloneEntry(entry)
			m.accumLines = []string{val}
		} else {
			m.accumLines = append(m.accumLines, val)
		}
		return nil
	}

	if err := m.flush(); err != nil {
		return fmt.Errorf("multiline: flush: %w", err)
	}
	m.pending = cloneEntry(entry)
	m.accumLines = []string{val}
	return nil
}

// Flush forwards any buffered entry to the inner sink.
func (m *MultilineSink) Flush() error {
	return m.flush()
}

func (m *MultilineSink) flush() error {
	if m.pending == nil {
		return nil
	}
	if len(m.accumLines) > 0 {
		m.pending.Fields[m.field] = strings.Join(m.accumLines, "\n")
	}
	err := m.inner.Write(*m.pending)
	m.pending = nil
	m.accumLines = nil
	return err
}

func (m *MultilineSink) isContinuation(val string) bool {
	for _, p := range m.prefixes {
		if strings.HasPrefix(val, p) {
			return true
		}
	}
	return false
}

func cloneEntry(e parser.Entry) *parser.Entry {
	fields := make(map[string]interface{}, len(e.Fields))
	for k, v := range e.Fields {
		fields[k] = v
	}
	return &parser.Entry{Timestamp: e.Timestamp, Level: e.Level, Message: e.Message, Fields: fields}
}
