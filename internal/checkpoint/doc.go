// Package checkpoint provides durable, atomic persistence of per-file read
// offsets for logpipe.
//
// # Overview
//
// When logpipe tails log files it tracks how many bytes have been consumed
// from each file. The checkpoint Store writes these offsets to a JSON file so
// that, after a process restart, tailing can resume from the last committed
// position rather than re-processing already-forwarded lines.
//
// # Usage
//
//	s, err := checkpoint.New("/var/lib/logpipe/offsets.json")
//	if err != nil { /* handle */ }
//
//	// After processing a line:
//	s.Set("/var/log/app.log", newOffset)
//
//	// Periodically, or on shutdown:
//	if err := s.Flush(); err != nil { /* handle */ }
//
// Flush writes to a sibling ".tmp" file and then renames it over the target,
// ensuring the on-disk state is never partially written.
package checkpoint
