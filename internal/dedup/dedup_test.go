package dedup

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestIsDuplicate_FirstOccurrenceNotDuplicate(t *testing.T) {
	d := New(100 * time.Millisecond)
	defer d.Stop()

	if d.IsDuplicate(`{"msg":"hello"}`) {
		t.Fatal("expected first occurrence to not be a duplicate")
	}
}

func TestIsDuplicate_SecondOccurrenceIsDuplicate(t *testing.T) {
	d := New(100 * time.Millisecond)
	defer d.Stop()

	line := `{"msg":"repeated"}`
	d.IsDuplicate(line)

	if !d.IsDuplicate(line) {
		t.Fatal("expected second occurrence within window to be a duplicate")
	}
}

func TestIsDuplicate_DifferentLinesNotDuplicate(t *testing.T) {
	d := New(100 * time.Millisecond)
	defer d.Stop()

	if d.IsDuplicate(`{"msg":"a"}`) {
		t.Fatal("first line should not be duplicate")
	}
	if d.IsDuplicate(`{"msg":"b"}`) {
		t.Fatal("different line should not be duplicate")
	}
}

func TestIsDuplicate_ExpiredWindowAllowsReentry(t *testing.T) {
	d := New(30 * time.Millisecond)
	defer d.Stop()

	line := `{"msg":"expire"}`
	d.IsDuplicate(line)

	time.Sleep(50 * time.Millisecond)

	if d.IsDuplicate(line) {
		t.Fatal("expected line to be accepted again after window expiry")
	}
}

func TestSize_TracksEntries(t *testing.T) {
	d := New(200 * time.Millisecond)
	defer d.Stop()

	for i := 0; i < 5; i++ {
		d.IsDuplicate(fmt.Sprintf(`{"i":%d}`, i))
	}

	if got := d.Size(); got != 5 {
		t.Fatalf("expected size 5, got %d", got)
	}
}

func TestIsDuplicate_ConcurrentSafe(t *testing.T) {
	d := New(500 * time.Millisecond)
	defer d.Stop()

	var wg sync.WaitGroup
	line := `{"msg":"concurrent"}`

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d.IsDuplicate(line)
		}()
	}
	wg.Wait()
}
