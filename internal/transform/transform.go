// Package transform provides a simple field-manipulation pipeline that
// runs a sequence of transformations against a parsed log entry before
// it is forwarded to any sink.
package transform

import (
	"strings"

	"github.com/logpipe/logpipe/internal/parser"
)

// Op is a single transformation that mutates a log entry in place.
type Op func(e *parser.Entry)

// Transformer holds an ordered list of Ops and applies them in sequence.
type Transformer struct {
	ops []Op
}

// New returns a Transformer that will apply the supplied ops in order.
func New(ops ...Op) *Transformer {
	return &Transformer{ops: ops}
}

// Apply runs every Op against e and returns the (mutated) entry.
func (t *Transformer) Apply(e *parser.Entry) *parser.Entry {
	for _, op := range t.ops {
		op(e)
	}
	return e
}

// AddField returns an Op that inserts key=value into Fields, overwriting any
// existing value for that key.
func AddField(key, value string) Op {
	return func(e *parser.Entry) {
		if e.Fields == nil {
			e.Fields = make(map[string]any)
		}
		e.Fields[key] = value
	}
}

// RemoveField returns an Op that deletes key from Fields (no-op if absent).
func RemoveField(key string) Op {
	return func(e *parser.Entry) {
		delete(e.Fields, key)
	}
}

// NormaliseLevel returns an Op that upper-cases the Level field so that
// downstream consumers receive a consistent value ("info", "INFO" → "INFO").
func NormaliseLevel() Op {
	return func(e *parser.Entry) {
		e.Level = strings.ToUpper(e.Level)
	}
}

// RedactField returns an Op that replaces the value of key in Fields with the
// supplied mask string (e.g. "***REDACTED***").
func RedactField(key, mask string) Op {
	return func(e *parser.Entry) {
		if _, ok := e.Fields[key]; ok {
			e.Fields[key] = mask
		}
	}
}
