// Package redact implements field-level log redaction for logpipe.
//
// A Redactor holds a list of Rules, each targeting a named top-level field
// in a parser.Entry's Extra map. Rules can be unconditional (redact every
// value) or pattern-based (redact only when the field value matches a
// regular expression).
//
// Usage:
//
//	rules := []redact.Rule{
//		{Field: "password"},
//		{Field: "email", Pattern: regexp.MustCompile(`@`)},
//	}
//	r, err := redact.New(rules, "[REDACTED]")
//	if err != nil { ... }
//	safeEntry := r.Apply(entry)
//
// Apply never mutates the original entry; it always returns a new copy
// with a freshly allocated Extra map.
package redact
