// Package redact provides field-level redaction for log entries,
// replacing sensitive values with a configurable mask string before
// the entry is forwarded to any sink.
package redact

import (
	"fmt"
	"regexp"

	"github.com/yourorg/logpipe/internal/parser"
)

const defaultMask = "[REDACTED]"

// Rule describes a single redaction rule.
type Rule struct {
	// Field is the top-level JSON key to inspect.
	Field string
	// Pattern, when non-nil, only redacts values that match the regexp.
	// When nil every value for Field is redacted.
	Pattern *regexp.Regexp
}

// Redactor applies a set of Rules to log entries.
type Redactor struct {
	rules []Rule
	mask  string
}

// New creates a Redactor with the supplied rules and mask string.
// If mask is empty the default mask "[REDACTED]" is used.
func New(rules []Rule, mask string) (*Redactor, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("redact: at least one rule is required")
	}
	for i, r := range rules {
		if r.Field == "" {
			return nil, fmt.Errorf("redact: rule[%d] has empty field", i)
		}
	}
	if mask == "" {
		mask = defaultMask
	}
	return &Redactor{rules: rules, mask: mask}, nil
}

// Apply returns a shallow copy of entry with sensitive fields masked.
// The original entry is never mutated.
func (r *Redactor) Apply(entry parser.Entry) parser.Entry {
	out := entry
	out.Extra = make(map[string]any, len(entry.Extra))
	for k, v := range entry.Extra {
		out.Extra[k] = v
	}

	for _, rule := range r.rules {
		v, ok := out.Extra[rule.Field]
		if !ok {
			continue
		}
		if rule.Pattern == nil {
			out.Extra[rule.Field] = r.mask
			continue
		}
		if s, ok := v.(string); ok && rule.Pattern.MatchString(s) {
			out.Extra[rule.Field] = r.mask
		}
	}
	return out
}
