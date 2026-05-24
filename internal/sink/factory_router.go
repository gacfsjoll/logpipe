package sink

import (
	"errors"
	"fmt"

	"logpipe/internal/config"
	"logpipe/internal/router"
)

// buildRouterSink constructs a RouterSink from a SinkConfig whose type is
// "router". Each entry in cfg.Routes is resolved to a concrete Sink via
// FromConfig, enabling fully composable routing pipelines.
func buildRouterSink(cfg config.SinkConfig) (Sink, error) {
	if len(cfg.Routes) == 0 {
		return nil, errors.New("router sink: routes must not be empty")
	}

	var routes []router.Route
	for i, rc := range cfg.Routes {
		if rc.Field == "" {
			return nil, fmt.Errorf("router sink: route[%d] missing field", i)
		}
		if len(rc.Values) == 0 {
			return nil, fmt.Errorf("router sink: route[%d] missing values", i)
		}
		if rc.Sink == nil {
			return nil, fmt.Errorf("router sink: route[%d] missing sink config", i)
		}
		inner, err := FromConfig(*rc.Sink)
		if err != nil {
			return nil, fmt.Errorf("router sink: route[%d] inner sink: %w", i, err)
		}
		routes = append(routes, router.Route{
			Field:  rc.Field,
			Values: rc.Values,
			Sink:   inner,
		})
	}

	return NewRouterSink(routes)
}
