package sink

import (
	"errors"

	"logpipe/internal/parser"
	"logpipe/internal/router"
)

// RouterSink wraps a router.Router and implements the Sink interface,
// dispatching each log entry to the appropriate downstream sink based
// on field-value routing rules.
type RouterSink struct {
	router *router.Router
}

// NewRouterSink constructs a RouterSink from the provided router.Routes.
// Returns an error if the routes are invalid or contain nil sinks.
func NewRouterSink(routes []router.Route) (*RouterSink, error) {
	if len(routes) == 0 {
		return nil, errors.New("router sink: at least one route is required")
	}
	r, err := router.New(routes)
	if err != nil {
		return nil, err
	}
	return &RouterSink{router: r}, nil
}

// Write dispatches the entry to matching downstream sinks via the router.
// If no route matches, the entry is silently dropped.
func (s *RouterSink) Write(entry parser.Entry) error {
	return s.router.Route(entry)
}
