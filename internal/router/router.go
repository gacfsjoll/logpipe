// Package router provides log-entry routing based on field-value rules.
// Each Route matches entries whose nominated field contains one of the
// allowed values and forwards matching entries to a dedicated sink.
// A fallback sink receives every entry that no route claims.
package router

import (
	"errors"
	"fmt"

	"logpipe/internal/parser"
	"logpipe/internal/sink"
)

// Route pairs a match predicate with the sink that should receive matching
// log entries.
type Route struct {
	// Field is the top-level JSON key to inspect (e.g. "level", "service").
	Field string
	// Values is the set of accepted values for Field.
	Values []string
	// Sink receives every entry that matches this route.
	Sink sink.Sink
}

// Router dispatches log entries to the first matching Route, or to the
// fallback Sink when no route matches.
type Router struct {
	routes   []Route
	fallback sink.Sink
	index    map[string]map[string]sink.Sink // field -> value -> sink
}

// New constructs a Router. At least one route must be provided. fallback may
// be nil, in which case unmatched entries are silently dropped.
func New(routes []Route, fallback sink.Sink) (*Router, error) {
	if len(routes) == 0 {
		return nil, errors.New("router: at least one route is required")
	}
	idx := make(map[string]map[string]sink.Sink)
	for _, r := range routes {
		if r.Field == "" {
			return nil, errors.New("router: route field must not be empty")
		}
		if len(r.Values) == 0 {
			return nil, fmt.Errorf("router: route for field %q has no values", r.Field)
		}
		if r.Sink == nil {
			return nil, fmt.Errorf("router: route for field %q has nil sink", r.Field)
		}
		if idx[r.Field] == nil {
			idx[r.Field] = make(map[string]sink.Sink)
		}
		for _, v := range r.Values {
			idx[r.Field][v] = r.Sink
		}
	}
	return &Router{routes: routes, fallback: fallback, index: idx}, nil
}

// Route dispatches entry to the appropriate sink and returns any write error.
func (r *Router) Route(entry parser.Entry) error {
	for field, values := range r.index {
		raw, ok := entry.Fields[field]
		if !ok {
			continue
		}
		s, ok := values[fmt.Sprintf("%v", raw)]
		if !ok {
			continue
		}
		return s.Write(entry)
	}
	if r.fallback != nil {
		return r.fallback.Write(entry)
	}
	return nil
}
