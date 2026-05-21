// Package rotator provides log file rotation detection for tailed files.
// It monitors inode changes and truncation events to signal when a tailer
// should reopen its file handle.
package rotator

import (
	"os"
	"sync"
	"time"
)

// Event describes a rotation occurrence on a watched path.
type Event struct {
	Path   string
	Reason string // "inode_changed" | "truncated"
}

// Rotator polls a set of file paths for rotation signals.
type Rotator struct {
	paths    []string
	interval time.Duration
	events   chan Event
	stop     chan struct{}
	wg       sync.WaitGroup
	inodes   map[string]uint64
	sizes    map[string]int64
}

// New creates a Rotator that checks the given paths at the given interval.
func New(paths []string, interval time.Duration) *Rotator {
	return &Rotator{
		paths:    paths,
		interval: interval,
		events:   make(chan Event, len(paths)*2),
		stop:     make(chan struct{}),
		inodes:   make(map[string]uint64),
		sizes:    make(map[string]int64),
	}
}

// Events returns the channel on which rotation events are delivered.
func (r *Rotator) Events() <-chan Event { return r.events }

// Start begins polling in a background goroutine.
func (r *Rotator) Start() {
	r.snapshot() // establish baseline
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				r.check()
			case <-r.stop:
				return
			}
		}
	}()
}

// Stop halts the polling goroutine and waits for it to exit.
func (r *Rotator) Stop() {
	close(r.stop)
	r.wg.Wait()
}

func (r *Rotator) snapshot() {
	for _, p := range r.paths {
		if fi, err := os.Stat(p); err == nil {
			r.inodes[p] = inode(fi)
			r.sizes[p] = fi.Size()
		}
	}
}

func (r *Rotator) check() {
	for _, p := range r.paths {
		fi, err := os.Stat(p)
		if err != nil {
			continue
		}
		curInode := inode(fi)
		prevInode, seenInode := r.inodes[p]
		prevSize := r.sizes[p]

		switch {
		case seenInode && curInode != prevInode:
			r.emit(Event{Path: p, Reason: "inode_changed"})
		case fi.Size() < prevSize:
			r.emit(Event{Path: p, Reason: "truncated"})
		}
		r.inodes[p] = curInode
		r.sizes[p] = fi.Size()
	}
}

func (r *Rotator) emit(e Event) {
	select {
	case r.events <- e:
	default:
	}
}
