package sink

import (
	"errors"

	"logpipe/internal/parser"
	"logpipe/internal/sampler"
)

// SampledSink wraps an inner Sink and forwards only a statistical sample of
// log entries determined by the configured keep rate.
type SampledSink struct {
	inner   Sink
	sampler *sampler.Sampler
}

// SamplerConfig holds the configuration for a SampledSink.
type SamplerConfig struct {
	// Rate is the fraction of entries to keep, in the range (0, 1].
	Rate float64
	// Seed is an optional fixed seed for the PRNG; 0 means random.
	Seed int64
}

// NewSampledSink constructs a SampledSink that forwards approximately Rate
// fraction of entries to inner.
func NewSampledSink(inner Sink, cfg SamplerConfig) (*SampledSink, error) {
	if inner == nil {
		return nil, errors.New("sampled sink: inner sink must not be nil")
	}
	s, err := sampler.New(cfg.Rate, cfg.Seed)
	if err != nil {
		return nil, err
	}
	return &SampledSink{inner: inner, sampler: s}, nil
}

// Write forwards entry to the inner sink only when the sampler decides to keep
// it. Dropped entries are silently discarded and no error is returned.
func (s *SampledSink) Write(entry parser.Entry) error {
	if !s.sampler.Keep() {
		return nil
	}
	return s.inner.Write(entry)
}
