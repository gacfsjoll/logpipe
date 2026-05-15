// Package watcher monitors a set of log source paths for availability
// and notifies the pipeline when files appear or disappear.
package watcher

import (
	"os"
	"sync"
	"time"
)

// State represents whether a watched path is currently accessible.
type State int

const (
	StateAvailable State = iota
	StateMissing
)

// Event is emitted whenever a file's availability changes.
type Event struct {
	Path  string
	State State
}

// Watcher periodically checks a list of paths and emits Events on change.
type Watcher struct {
	paths    []string
	interval time.Duration
	events   chan Event
	stop     chan struct{}
	mu       sync.Mutex
	last     map[string]State
}

// New creates a Watcher that polls the given paths at the given interval.
func New(paths []string, interval time.Duration) *Watcher {
	last := make(map[string]State, len(paths))
	for _, p := range paths {
		last[p] = StateMissing
	}
	return &Watcher{
		paths:    paths,
		interval: interval,
		events:   make(chan Event, len(paths)*2),
		stop:     make(chan struct{}),
		last:     last,
	}
}

// Events returns the read-only channel of file state change events.
func (w *Watcher) Events() <-chan Event {
	return w.events
}

// Start begins polling in a background goroutine.
func (w *Watcher) Start() {
	go w.run()
}

// Stop shuts down the polling goroutine.
func (w *Watcher) Stop() {
	close(w.stop)
}

func (w *Watcher) run() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	w.check() // immediate first check
	for {
		select {
		case <-ticker.C:
			w.check()
		case <-w.stop:
			return
		}
	}
}

func (w *Watcher) check() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, p := range w.paths {
		_, err := os.Stat(p)
		current := StateAvailable
		if err != nil {
			current = StateMissing
		}
		if prev, ok := w.last[p]; !ok || prev != current {
			w.last[p] = current
			w.events <- Event{Path: p, State: current}
		}
	}
}
