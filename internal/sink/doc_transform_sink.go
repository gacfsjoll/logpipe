// Package sink provides Sink implementations for logpipe.
//
// # TransformSink
//
// TransformSink decorates any Sink with a [transform.Transformer], applying
// field mutations (add, remove, normalise) to each log entry before it is
// forwarded to the wrapped sink.
//
// Usage:
//
//	tr := transform.New(
//		transform.AddField("service", "api"),
//		transform.NormaliseLevel(),
//	)
//	inner, _ := sink.NewStdoutSink()
//	s, err := sink.NewTransformSink(inner, tr)
//
// The original [parser.Entry] is never mutated; Apply returns a shallow copy
// with the requested changes applied.
package sink
