// Package pipeline wires together tailers, parsers, and sinks into a
// running log-forwarding pipeline.
package pipeline

import (
	"context"
	"log"
	"sync"

	"logpipe/internal/config"
	"logpipe/internal/parser"
	"logpipe/internal/sink"
	"logpipe/internal/tail"
)

// Pipeline coordinates all components for a given configuration.
type Pipeline struct {
	cfg   *config.Config
	sinks []sink.Sink
	wg    sync.WaitGroup
}

// New creates a Pipeline from the provided configuration.
// It initialises all configured sinks and returns any construction error.
func New(cfg *config.Config) (*Pipeline, error) {
	var sinks []sink.Sink
	for _, sc := range cfg.Sinks {
		if sc.Type == "http" {
			s, err := sink.NewHTTPSink(sc)
			if err != nil {
				return nil, err
			}
			sinks = append(sinks, s)
		}
	}
	return &Pipeline{cfg: cfg, sinks: sinks}, nil
}

// Run starts tailing all configured source files and forwards parsed entries
// to every sink. It blocks until ctx is cancelled.
func (p *Pipeline) Run(ctx context.Context) {
	for _, src := range p.cfg.Sources {
		src := src // capture
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.tailSource(ctx, src)
		}()
	}
	p.wg.Wait()
}

func (p *Pipeline) tailSource(ctx context.Context, src config.Source) {
	pr := parser.New(src.TimestampKey, src.LevelKey, src.MessageKey)
	t, err := tail.New(src.Path)
	if err != nil {
		log.Printf("pipeline: failed to tail %s: %v", src.Path, err)
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-t.Lines():
			if !ok {
				return
			}
			entry, err := pr.Parse(line)
			if err != nil {
				log.Printf("pipeline: parse error (%s): %v", src.Path, err)
				continue
			}
			p.writeToSinks(ctx, entry)
		}
	}
}

// writeToSinks forwards a parsed log entry to all configured sinks.
// Errors from individual sinks are logged but do not stop delivery to
// the remaining sinks.
func (p *Pipeline) writeToSinks(ctx context.Context, entry parser.Entry) {
	for _, s := range p.sinks {
		if err := s.Write(ctx, entry); err != nil {
			log.Printf("pipeline: sink write error: %v", err)
		}
	}
}
