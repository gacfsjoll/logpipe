// Package rotator detects log file rotation by polling file metadata.
//
// Rotation is signalled when either:
//   - the inode of a watched path changes (the file was replaced / renamed), or
//   - the file size decreases (the file was truncated in-place).
//
// Consumers receive [Event] values from the channel returned by [Rotator.Events]
// and should reopen their file handles in response.
//
// Example usage:
//
//	r := rotator.New([]string{"/var/log/app.log"}, 5*time.Second)
//	r.Start()
//	defer r.Stop()
//	for ev := range r.Events() {
//		log.Printf("rotation detected on %s: %s", ev.Path, ev.Reason)
//	}
package rotator
