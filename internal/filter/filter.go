// Package filter provides log entry filtering based on field-level predicates.
// Entries that do not match all configured rules are dropped from the pipeline.
package filter

import (
	"fmt"
	"strings"

	"logpipe/internal/parser"
)

// Rule describes a single field-match predicate.
type Rule struct {
	// Field is the top-level JSON key to inspect (e.g. "level", "service").
	Field string
	// Values holds the accepted values; the rule passes when the field matches
	// any one of them (OR semantics). Comparison is case-insensitive.
	Values []string
}

// Filter drops log entries that fail one or more rules.
type Filter struct {
	rules []Rule
}

// New creates a Filter from the supplied rules.
// Returns an error when any rule has an empty Field or no Values.
func New(rules []Rule) (*Filter, error) {
	for i, r := range rules {
		if strings.TrimSpace(r.Field) == "" {
			return nil, fmt.Errorf("filter rule %d: field must not be empty", i)
		}
		if len(r.Values) == 0 {
			return nil, fmt.Errorf("filter rule %d: values must not be empty", i)
		}
	}
	return &Filter{rules: rules}, nil
}

// Keep returns true when the entry satisfies every rule.
// If no rules are configured every entry is kept.
func (f *Filter) Keep(e parser.Entry) bool {
	for _, rule := range f.rules {
		raw, ok := e.Fields[rule.Field]
		if !ok {
			return false
		}
		val := strings.ToLower(fmt.Sprintf("%v", raw))
		matched := false
		for _, accepted := range rule.Values {
			if strings.ToLower(accepted) == val {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

// Stats returns the number of configured rules.
func (f *Filter) Stats() int {
	return len(f.rules)
}
