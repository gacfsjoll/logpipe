// Package sink provides the RouterSink, which dispatches log entries to
// different downstream sinks based on field-value matching rules.
//
// # Overview
//
// RouterSink wraps the internal/router package and exposes it as a Sink,
// making it composable with all other sink decorators in the pipeline.
//
// # Configuration
//
// Use type "router" in your sink configuration. Each route specifies a
// field name, a list of acceptable values, and a nested sink config that
// will receive matching entries:
//
//	- type: router
//	  routes:
//	    - field: level
//	      values: [error, fatal]
//	      sink:
//	        type: http
//	        url: https://alerts.example.com/logs
//	    - field: level
//	      values: [info, debug]
//	      sink:
//	        type: stdout
//
// Entries that do not match any route are silently dropped.
package sink
