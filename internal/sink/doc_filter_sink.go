// Package sink provides Sink implementations for forwarding parsed log entries
// to various destinations.
//
// # FilterSink
//
// FilterSink wraps any inner Sink and applies one or more field-based filter
// rules before forwarding entries. Entries that do not satisfy every rule are
// silently dropped; entries that pass all rules are forwarded unchanged.
//
// Rules are evaluated using the filter.Filter type from internal/filter, which
// performs exact-match comparisons on string-valued entry fields.
//
// Example usage:
//
//	rules := []filter.Rule{
//		{Field: "level",   Values: []string{"error", "fatal"}},
//		{Field: "service", Values: []string{"payments"}},
//	}
//	fs, err := sink.NewFilterSink(inner, rules)
//	if err != nil {
//		log.Fatal(err)
//	}
package sink
