// Package router implements content-based routing for structured log entries.
//
// # Overview
//
// A Router inspects a nominated field of each [parser.Entry] and forwards the
// entry to the first [sink.Sink] whose accepted-values list contains the
// field's value.  When no route matches, the entry is sent to an optional
// fallback sink; if no fallback is configured the entry is silently discarded.
//
// # Usage
//
//	routes := []router.Route{
//	    {Field: "level",   Values: []string{"error", "fatal"}, Sink: alertSink},
//	    {Field: "service", Values: []string{"payments"},       Sink: paymentsSink},
//	}
//	r, err := router.New(routes, defaultSink)
//	if err != nil { … }
//	r.Route(entry)
//
// Routes are evaluated in field-index order; the first matching field+value
// pair wins.  If the same field appears in multiple routes the last-registered
// value mapping wins (last-write semantics within the internal index).
package router
