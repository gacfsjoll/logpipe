package sink

import (
	"errors"
	"fmt"

	"logpipe/internal/parser"
)

// RenameSink rewrites field keys in a log entry before forwarding to an inner
// sink. Each rule maps an existing key to a new key name. If the source key is
// absent the rule is silently skipped.
type RenameSink struct {
	inner Sink
	rules map[string]string // oldKey -> newKey
}

// RenameRule describes a single field rename operation.
type RenameRule struct {
	From string
	To   string
}

// NewRenameSink constructs a RenameSink. At least one rule must be provided and
// every rule must have non-empty From and To values.
func NewRenameSink(inner Sink, rules []RenameRule) (*RenameSink, error) {
	if inner == nil {
		return nil, errors.New("rename_sink: inner sink must not be nil")
	}
	if len(rules) == 0 {
		return nil, errors.New("rename_sink: at least one rule is required")
	}
	m := make(map[string]string, len(rules))
	for i, r := range rules {
		if r.From == "" {
			return nil, fmt.Errorf("rename_sink: rule[%d]: From must not be empty", i)
		}
		if r.To == "" {
			return nil, fmt.Errorf("rename_sink: rule[%d]: To must not be empty", i)
		}
		m[r.From] = r.To
	}
	return &RenameSink{inner: inner, rules: m}, nil
}

// Write renames configured fields in a shallow copy of the entry, then
// forwards the modified entry to the inner sink.
func (s *RenameSink) Write(entry parser.Entry) error {
	out := make(parser.Entry, len(entry))
	for k, v := range entry {
		if newKey, ok := s.rules[k]; ok {
			out[newKey] = v
		} else {
			out[k] = v
		}
	}
	return s.inner.Write(out)
}
