package sink

import (
	"errors"
	"fmt"

	"logpipe/internal/parser"
)

// levelOrder maps canonical level strings to a numeric severity.
var levelOrder = map[string]int{
	"debug": 0,
	"info":  1,
	"warn":  2,
	"error": 3,
	"fatal": 4,
}

// LevelFilterSink drops log entries whose level is below the configured
// minimum severity. Level comparison is case-insensitive.
type LevelFilterSink struct {
	inner    Sink
	minLevel int
	minName  string
}

// NewLevelFilterSink returns a Sink that forwards entries to inner only when
// their "level" field is greater than or equal to minLevel.
// minLevel must be one of: debug, info, warn, error, fatal.
func NewLevelFilterSink(inner Sink, minLevel string) (*LevelFilterSink, error) {
	if inner == nil {
		return nil, errors.New("level_filter: inner sink must not be nil")
	}
	if minLevel == "" {
		return nil, errors.New("level_filter: min_level must not be empty")
	}
	ord, ok := levelOrder[minLevel]
	if !ok {
		return nil, fmt.Errorf("level_filter: unknown level %q; must be one of debug, info, warn, error, fatal", minLevel)
	}
	return &LevelFilterSink{
		inner:   inner,
		minLevel: ord,
		minName:  minLevel,
	}, nil
}

// Write forwards entry to the inner sink if its level meets the minimum
// threshold. Entries with an absent or unrecognised level field are dropped.
func (s *LevelFilterSink) Write(entry parser.Entry) error {
	raw, ok := entry["level"]
	if !ok {
		return nil
	}
	lvlStr, ok := raw.(string)
	if !ok {
		return nil
	}
	ord, ok := levelOrder[lvlStr]
	if !ok {
		return nil
	}
	if ord < s.minLevel {
		return nil
	}
	return s.inner.Write(entry)
}
