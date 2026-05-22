package router_test

import (
	"strings"
	"testing"

	"logpipe/internal/parser"
	"logpipe/internal/router"
	"logpipe/internal/sink"
)

// TestRouter_MultiFieldRouting verifies that routes on different fields work
// independently and that the first matching field dispatches correctly.
func TestRouter_MultiFieldRouting(t *testing.T) {
	var errBuf, svcBuf, fallBuf bytes.Buffer

	errSink := sink.NewStdoutSinkWithWriter(&errBuf)
	svcSink := sink.NewStdoutSinkWithWriter(&svcBuf)
	fallSink := sink.NewStdoutSinkWithWriter(&fallBuf)

	r, err := router.New([]router.Route{
		{Field: "level", Values: []string{"error"}, Sink: errSink},
		{Field: "service", Values: []string{"payments"}, Sink: svcSink},
	}, fallSink)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		fields      map[string]any
		wantErr     bool
		wantSvc     bool
		wantFall    bool
	}{
		{map[string]any{"level": "error"}, true, false, false},
		{map[string]any{"service": "payments"}, false, true, false},
		{map[string]any{"level": "info"}, false, false, true},
	}

	for _, tc := range cases {
		errBuf.Reset()
		svcBuf.Reset()
		fallBuf.Reset()

		entry := parser.Entry{
			Level:   "info",
			Message: "msg",
			Fields:  tc.fields,
		}
		if err := r.Route(entry); err != nil {
			t.Fatalf("Route() error: %v", err)
		}
		if got := errBuf.Len() > 0; got != tc.wantErr {
			t.Errorf("fields=%v: errSink got=%v want=%v", tc.fields, got, tc.wantErr)
		}
		if got := svcBuf.Len() > 0; got != tc.wantSvc {
			t.Errorf("fields=%v: svcSink got=%v want=%v", tc.fields, got, tc.wantSvc)
		}
		if got := fallBuf.Len() > 0; got != tc.wantFall {
			t.Errorf("fields=%v: fallSink got=%v want=%v", tc.fields, got, tc.wantFall)
		}
	}
}

// TestRouter_HighVolume ensures the router handles many entries without error.
func TestRouter_HighVolume(t *testing.T) {
	var buf bytes.Buffer
	s := sink.NewStdoutSinkWithWriter(&buf)
	r, _ := router.New([]router.Route{
		{Field: "level", Values: []string{"error"}, Sink: s},
	}, nil)

	for i := 0; i < 1000; i++ {
		entry := parser.Entry{Level: "error", Message: "boom", Fields: map[string]any{"level": "error"}}
		if err := r.Route(entry); err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
	}
	lines := strings.Count(buf.String(), "\n")
	if lines != 1000 {
		t.Errorf("expected 1000 lines, got %d", lines)
	}
}
