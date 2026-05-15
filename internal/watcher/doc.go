// Package watcher provides lightweight file-availability monitoring for
// logpipe source paths.
//
// A Watcher polls a list of filesystem paths at a configurable interval and
// emits an Event on its Events channel whenever a path transitions between
// present (StateAvailable) and absent (StateMissing).  This allows the
// pipeline to react to log-file rotations or delayed file creation without
// relying on OS-specific inotify APIs.
//
// Usage:
//
//	w := watcher.New(paths, 5*time.Second)
//	w.Start()
//	for ev := range w.Events() {
//		log.Printf("path %s state %v", ev.Path, ev.State)
//	}
package watcher
