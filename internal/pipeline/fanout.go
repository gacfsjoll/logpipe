package pipeline

import (
	"fmt"
	"strings"

	"logpipe/internal/parser"
)

// Sink is the interface that wraps the Write method.
type Sink interface {
	Write(entry parser.Entry) error
}

// fanOut writes a single log entry to all provided sinks.
// It collects errors from all sinks and returns a combined error if any failed.
func fanOut(entry parser.Entry, sinks []Sink) error {
	if len(sinks) == 0 {
		return nil
	}

	var errs []string
	for _, s := range sinks {
		if err := s.Write(entry); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("fanout errors: %s", strings.Join(errs, "; "))
	}
	return nil
}
