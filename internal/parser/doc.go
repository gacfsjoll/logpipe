// Package parser provides JSON log line parsing for logpipe.
//
// It accepts raw log lines from tailed files and converts them into
// structured Entry values that downstream sinks can consume. The parser
// recognises common field aliases (e.g. "msg" / "message", "ts" / "time" /
// "timestamp", "severity" / "level") so that logs from different frameworks
// are normalised to a consistent schema. Any unrecognised top-level keys are
// collected into the Entry.Fields map and forwarded without modification.
package parser
