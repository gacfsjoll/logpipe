package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

// SnapshotResponse is the JSON shape returned by the metrics HTTP handler.
type SnapshotResponse struct {
	UptimeSeconds float64 `json:"uptime_seconds"`
	LinesRead     int64   `json:"lines_read"`
	LinesParsed   int64   `json:"lines_parsed"`
	LinesDropped  int64   `json:"lines_dropped"`
	SinkErrors    int64   `json:"sink_errors"`
}

// Handler returns an http.HandlerFunc that serialises the current metrics
// snapshot as a JSON response. It is safe to register on any ServeMux.
func (m *Metrics) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := m.Snapshot()
		uptime := time.Since(m.startTime).Seconds()

		resp := SnapshotResponse{
			UptimeSeconds: uptime,
			LinesRead:     snap.LinesRead,
			LinesParsed:   snap.LinesParsed,
			LinesDropped:  snap.LinesDropped,
			SinkErrors:    snap.SinkErrors,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}
}
