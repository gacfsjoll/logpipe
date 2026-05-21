// Package healthcheck exposes a lightweight HTTP endpoint (/healthz) that
// aggregates the health of all logpipe components.
//
// Usage:
//
//	checker := healthcheck.New()
//
//	// Register a check for each source file watcher.
//	checker.Register("source:app.log", func() (bool, string) {
//		if watcherIsRunning {
//			return true, ""
//		}
//		return false, "watcher stopped unexpectedly"
//	})
//
//	srv := healthcheck.NewServer(":9091", checker)
//	if err := srv.Start(); err != nil {
//		log.Fatal(err)
//	}
//
// The handler responds with HTTP 200 and a JSON Report when all registered
// checks pass, or HTTP 503 when one or more checks fail.
package healthcheck
