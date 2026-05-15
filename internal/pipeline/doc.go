// Package pipeline wires together a tailer, parser, and one or more sinks
// for a single log-source configuration entry.
//
// A Pipeline reads raw log lines produced by a tail.Tailer, parses each line
// into a parser.Entry via a parser.Parser, and fans the result out to every
// configured sink.Sink concurrently.  Metrics are recorded for each line
// processed and for any errors encountered.
//
// Usage:
//
//	p, err := pipeline.New(cfg, sinks, metrics)
//	if err != nil {
//		log.Fatal(err)
//	}
//	go p.Run(ctx)
package pipeline
