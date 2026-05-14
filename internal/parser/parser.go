package parser

import (
	"encoding/json"
	"fmt"
	"time"
)

// Entry represents a parsed log entry.
type Entry struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Source    string            `json:"source"`
	Fields    map[string]any    `json:"fields,omitempty"`
	Raw       string            `json:"-"`
}

// Parser parses raw log lines into structured Entry values.
type Parser struct {
	source string
}

// New creates a new Parser associated with the given source label.
func New(source string) *Parser {
	return &Parser{source: source}
}

// Parse attempts to decode a raw JSON log line into an Entry.
// Unknown top-level keys are collected into Fields.
func (p *Parser) Parse(line string) (*Entry, error) {
	if line == "" {
		return nil, fmt.Errorf("empty line")
	}

	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	entry := &Entry{
		Source: p.source,
		Raw:    line,
		Fields: make(map[string]any),
	}

	for k, v := range raw {
		switch k {
		case "timestamp", "time", "ts":
			if s, ok := v.(string); ok {
				if t, err := time.Parse(time.RFC3339, s); err == nil {
					entry.Timestamp = t
				}
			}
		case "level", "severity":
			if s, ok := v.(string); ok {
				entry.Level = s
			}
		case "message", "msg":
			if s, ok := v.(string); ok {
				entry.Message = s
			}
		default:
			entry.Fields[k] = v
		}
	}

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	return entry, nil
}
