package sink_test

import (
	"testing"
	"time"

	"logpipe/internal/parser"
	"logpipe/internal/sink"
)

func dedupEntry(raw string) parser.Entry {
	return parser.Entry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "test",
		Raw:       raw,
		Fields:    map[string]any{},
	}
}

func validDedupCfg() sink.DedupConfig {
	return sink.DedupConfig{WindowSeconds: 5, MaxEntries: 100}
}

func TestNewDedupSink_NilInnerReturnsError(t *testing.T) {
	_, err := sink.NewDedupSink(nil, validDedupCfg())
	if err == nil {
		t.Fatal("expected error for nil inner sink")
	}
}

func TestNewDedupSink_InvalidWindowReturnsError(t *testing.T) {
	collector := &captureSink{}
	_, err := sink.NewDedupSink(collector, sink.DedupConfig{WindowSeconds: 0, MaxEntries: 10})
	if err == nil {
		t.Fatal("expected error for zero WindowSeconds")
	}
}

func TestNewDedupSink_InvalidMaxEntriesReturnsError(t *testing.T) {
	collector := &captureSink{}
	_, err := sink.NewDedupSink(collector, sink.DedupConfig{WindowSeconds: 5, MaxEntries: 0})
	if err == nil {
		t.Fatal("expected error for zero MaxEntries")
	}
}

func TestDedupSink_Write_ForwardsUniqueEntry(t *testing.T) {
	collector := &captureSink{}
	ds, err := sink.NewDedupSink(collector, validDedupCfg())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := ds.Write(dedupEntry(`{"msg":"hello"}`)); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	if len(collector.entries) != 1 {
		t.Fatalf("expected 1 forwarded entry, got %d", len(collector.entries))
	}
}

func TestDedupSink_Write_DropsDuplicate(t *testing.T) {
	collector := &captureSink{}
	ds, err := sink.NewDedupSink(collector, validDedupCfg())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw := `{"msg":"repeated"}`
	_ = ds.Write(dedupEntry(raw))
	_ = ds.Write(dedupEntry(raw))
	_ = ds.Write(dedupEntry(raw))

	if len(collector.entries) != 1 {
		t.Fatalf("expected 1 forwarded entry after dedup, got %d", len(collector.entries))
	}
}

func TestDedupSink_Write_AllowsDifferentLines(t *testing.T) {
	collector := &captureSink{}
	ds, err := sink.NewDedupSink(collector, validDedupCfg())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = ds.Write(dedupEntry(`{"msg":"a"}`))
	_ = ds.Write(dedupEntry(`{"msg":"b"}`))
	_ = ds.Write(dedupEntry(`{"msg":"c"}`))

	if len(collector.entries) != 3 {
		t.Fatalf("expected 3 forwarded entries, got %d", len(collector.entries))
	}
}
