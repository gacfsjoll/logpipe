// Package filter implements predicate-based log entry filtering for logpipe.
//
// A Filter holds one or more Rules. Each Rule specifies a field name and a set
// of accepted values. An entry must satisfy ALL rules to pass through (AND
// semantics across rules; OR semantics across each rule's value list).
//
// Example usage:
//
//	f, err := filter.New([]filter.Rule{
//		{Field: "level",   Values: []string{"error", "warn"}},
//		{Field: "service", Values: []string{"payments", "auth"}},
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	if f.Keep(entry) {
//		// forward entry to sinks
//	}
//
// Value comparisons are case-insensitive. Entries that are missing a required
// field are always dropped.
package filter
