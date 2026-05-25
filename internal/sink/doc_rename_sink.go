// Package sink provides composable sink implementations for logpipe.
//
// # RenameSink
//
// RenameSink rewrites field keys in a log entry before forwarding to an inner
// sink. It is useful for normalising field names across heterogeneous log
// sources, for example renaming "msg" to "message" or "ts" to "timestamp"
// before the entry reaches a downstream sink that expects a canonical schema.
//
// Rules are applied in a single pass over a shallow copy of the entry so the
// original entry is never mutated. If a rule's source key is absent the rule
// is silently skipped.
//
// Example:
//
//	rules := []sink.RenameRule{
//		{From: "msg",  To: "message"},
//		{From: "ts",   To: "timestamp"},
//		{From: "svc",  To: "service"},
//	}
//	s, err := sink.NewRenameSink(inner, rules)
package sink
