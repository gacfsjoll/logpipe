package sink

import (
	"errors"
	"fmt"
	"time"

	"logpipe/internal/parser"
)

// TimestampSink rewrites the timestamp field of every log entry to a
// normalised RFC3339Nano string before forwarding to the inner sink.
// Optionally the field name written can be overridden via FieldName.
type TimestampSink struct {
	inner     Sink
	fieldName string
	location  *time.Location
}

// TimestampConfig holds options for NewTimestampSink.
type TimestampConfig struct {
	// FieldName is the key written into the entry. Defaults to "timestamp".
	FieldName string
	// Location is the timezone used when formatting. Defaults to UTC.
	Location *time.Location
}

// NewTimestampSink returns a sink that normalises the timestamp field of
// each entry to RFC3339Nano before delegating to inner.
func NewTimestampSink(inner Sink, cfg TimestampConfig) (*TimestampSink, error) {
	if inner == nil {
		return nil, errors.New("timestamp sink: inner sink must not be nil")
	}
	fieldName := cfg.FieldName
	if fieldName == "" {
		fieldName = "timestamp"
	}
	loc := cfg.Location
	if loc == nil {
		loc = time.UTC
	}
	return &TimestampSink{
		inner:     inner,
		fieldName: fieldName,
		location:  loc,
	}, nil
}

// Write normalises the timestamp on a copy of entry and forwards it.
func (s *TimestampSink) Write(entry parser.Entry) error {
	out := entry.Clone()
	out.Fields[s.fieldName] = entry.Timestamp.In(s.location).Format(time.RFC3339Nano)
	if err := s.inner.Write(out); err != nil {
		return fmt.Errorf("timestamp sink: %w", err)
	}
	return nil
}
