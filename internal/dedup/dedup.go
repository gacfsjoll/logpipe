// Package dedup provides a log entry deduplicator that suppresses repeated
// identical log lines within a configurable time window.
package dedup

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// entry tracks the last seen time and occurrence count for a given hash.
type entry struct {
	lastSeen time.Time
	count    int
}

// Deduplicator suppresses duplicate log lines seen within a sliding window.
type Deduplicator struct {
	mu      sync.Mutex
	window  time.Duration
	seen    map[string]*entry
	stopCh  chan struct{}
}

// New creates a Deduplicator with the given deduplication window.
// A background goroutine evicts stale entries every window interval.
func New(window time.Duration) *Deduplicator {
	d := &Deduplicator{
		window: window,
		seen:   make(map[string]*entry),
		stopCh: make(chan struct{}),
	}
	go d.evictLoop()
	return d
}

// IsDuplicate returns true if the given raw line was already seen within the
// configured window. It always records the line on first sight.
func (d *Deduplicator) IsDuplicate(line string) bool {
	h := hash(line)
	now := time.Now()

	d.mu.Lock()
	defer d.mu.Unlock()

	if e, ok := d.seen[h]; ok && now.Sub(e.lastSeen) < d.window {
		e.count++
		e.lastSeen = now
		return true
	}

	d.seen[h] = &entry{lastSeen: now, count: 1}
	return false
}

// Stop halts the background eviction goroutine.
func (d *Deduplicator) Stop() {
	close(d.stopCh)
}

// Size returns the number of unique entries currently tracked.
func (d *Deduplicator) Size() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.seen)
}

func (d *Deduplicator) evictLoop() {
	ticker := time.NewTicker(d.window)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			d.evict()
		case <-d.stopCh:
			return
		}
	}
}

func (d *Deduplicator) evict() {
	cutoff := time.Now().Add(-d.window)
	d.mu.Lock()
	defer d.mu.Unlock()
	for k, e := range d.seen {
		if e.lastSeen.Before(cutoff) {
			delete(d.seen, k)
		}
	}
}

func hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
