// Package sink provides a collection of composable sink implementations for
// logpipe.
//
// # MultilineSink
//
// MultilineSink coalesces consecutive log entries whose configurable field
// value begins with one of a set of continuation prefixes into a single entry
// before forwarding downstream. This is useful for handling stack traces or
// other multi-line log patterns emitted by runtimes such as the JVM or Python.
//
// Example usage:
//
//	 inner, _ := sink.NewStdoutSink()
//	 s, err := sink.NewMultilineSink(inner, "message", []string{"\t", "at "})
//	 if err != nil {
//	     log.Fatal(err)
//	 }
//	 defer s.Flush()
//
// Flush must be called when the source is exhausted to ensure the last
// buffered entry group is forwarded to the inner sink.
package sink
