package sampler_test

import (
	"testing"

	"github.com/your-org/logpipe/internal/sampler"
)

func TestNew_InvalidRate(t *testing.T) {
	cases := []struct {
		name string
		rate float64
	}{
		{"zero", 0},
		{"negative", -0.5},
		{"above one", 1.1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sampler.New(tc.rate, 42)
			if err == nil {
				t.Fatalf("expected error for rate %v, got nil", tc.rate)
			}
		})
	}
}

func TestNew_ValidRate(t *testing.T) {
	s, err := sampler.New(1.0, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil sampler")
	}
}

func TestKeep_RateOne_KeepsAll(t *testing.T) {
	s, _ := sampler.New(1.0, 42)
	const n = 1000
	for i := 0; i < n; i++ {
		if !s.Keep() {
			t.Fatal("rate=1.0 should keep every entry")
		}
	}
	kept, dropped := s.Stats()
	if kept != n {
		t.Errorf("kept=%d, want %d", kept, n)
	}
	if dropped != 0 {
		t.Errorf("dropped=%d, want 0", dropped)
	}
}

func TestKeep_LowRate_DropsEntries(t *testing.T) {
	// With rate=0.1 and a large sample we expect far fewer than all kept.
	s, _ := sampler.New(0.1, 99)
	const n = 10_000
	var kept int
	for i := 0; i < n; i++ {
		if s.Keep() {
			kept++
		}
	}
	// Allow generous tolerance: expect between 5% and 20%.
	if kept < n*5/100 || kept > n*20/100 {
		t.Errorf("kept=%d out of %d; expected roughly 10%%", kept, n)
	}
}

func TestStats_AndReset(t *testing.T) {
	s, _ := sampler.New(1.0, 1)
	s.Keep()
	s.Keep()
	kept, dropped := s.Stats()
	if kept != 2 || dropped != 0 {
		t.Fatalf("want kept=2 dropped=0, got kept=%d dropped=%d", kept, dropped)
	}
	s.Reset()
	kept, dropped = s.Stats()
	if kept != 0 || dropped != 0 {
		t.Fatalf("after Reset want 0/0, got %d/%d", kept, dropped)
	}
}
