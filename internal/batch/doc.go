// Package batch implements a size- and time-bounded entry accumulator for
// logpipe pipelines.
//
// # Overview
//
// A Batcher groups individual [parser.Entry] values into slices and delivers
// them to a [FlushFunc] in two situations:
//
//  1. The internal buffer reaches the configured maximum size (immediate flush).
//  2. The flush interval elapses without the buffer filling (periodic flush).
//
// # Usage
//
//	b := batch.New(50, 5*time.Second, func(entries []parser.Entry) {
//	    // forward entries to a bulk sink
//	})
//	go b.Run(ctx) // drive the interval ticker
//
//	for entry := range logCh {
//	    b.Add(entry)
//	}
//
// Calling [Batcher.Run] with a cancelled context triggers a final flush of any
// buffered entries before returning, ensuring no data is silently dropped on
// shutdown.
package batch
