package metrics_test

import (
	"testing"
	"time"

	"logpipe/internal/metrics"
)

func TestNew_InitialisesStartTime(t *testing.T) {
	before := time.Now()
	m := metrics.New()
	after := time.Now()

	if m.StartTime.Before(before) || m.StartTime.After(after) {
		t.Errorf("StartTime %v not in expected range [%v, %v]", m.StartTime, before, after)
	}
}

func TestCounters_IncrementAndSnapshot(t *testing.T) {
	m := metrics.New()

	m.LinesRead.Add(10)
	m.LinesParsed.Add(8)
	m.ParseErrors.Add(2)
	m.SinkWrites.Add(7)
	m.SinkErrors.Add(1)

	snap := m.Snapshot()

	if snap.LinesRead != 10 {
		t.Errorf("LinesRead: want 10, got %d", snap.LinesRead)
	}
	if snap.LinesParsed != 8 {
		t.Errorf("LinesParsed: want 8, got %d", snap.LinesParsed)
	}
	if snap.ParseErrors != 2 {
		t.Errorf("ParseErrors: want 2, got %d", snap.ParseErrors)
	}
	if snap.SinkWrites != 7 {
		t.Errorf("SinkWrites: want 7, got %d", snap.SinkWrites)
	}
	if snap.SinkErrors != 1 {
		t.Errorf("SinkErrors: want 1, got %d", snap.SinkErrors)
	}
	if snap.Uptime <= 0 {
		t.Errorf("Uptime should be positive, got %v", snap.Uptime)
	}
}

func TestCounters_Reset(t *testing.T) {
	m := metrics.New()
	m.LinesRead.Add(5)
	m.SinkErrors.Add(3)

	m.Reset()
	snap := m.Snapshot()

	if snap.LinesRead != 0 {
		t.Errorf("LinesRead after reset: want 0, got %d", snap.LinesRead)
	}
	if snap.SinkErrors != 0 {
		t.Errorf("SinkErrors after reset: want 0, got %d", snap.SinkErrors)
	}
}

func TestCounters_ConcurrentIncrements(t *testing.T) {
	m := metrics.New()
	done := make(chan struct{})

	for i := 0; i < 100; i++ {
		go func() {
			m.LinesRead.Add(1)
			done <- struct{}{}
		}()
	}
	for i := 0; i < 100; i++ {
		<-done
	}

	if got := m.LinesRead.Load(); got != 100 {
		t.Errorf("concurrent LinesRead: want 100, got %d", got)
	}
}
