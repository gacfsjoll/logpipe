// Package metrics provides thread-safe atomic counters used throughout the
// logpipe pipeline to track operational statistics such as lines read,
// successfully parsed entries, parse errors, sink writes, and sink errors.
//
// Usage:
//
//	m := metrics.New()
//	m.LinesRead.Add(1)
//	snap := m.Snapshot()
//	fmt.Println(snap.LinesRead, snap.Uptime)
package metrics
