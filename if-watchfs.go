/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// a path can be absolute or relative
// — a relative path uses the process’ current working directory
//
// func filepath.Abs(path string) (string, error)
// func filepath.IsAbs(path string) bool

// a path can contain symlinks
//
// func filepath.EvalSymlinks(path string) (string, error)

// a path can contain .
// a path can contain multiple separators in sequence
// a path can contain ..:
// — a path can contain inner ..
// — a path can seek the root parent directory: "/.."
//
// func filepath.Clean(path string) string

// a path can end with separator

// watch a symlink that is a dir
// watch special files

// Operator read-write-create-delete etc.

// event contains:
// unique event ID, ns timestamp, path, operations

// event should be an interface that can have
// implementation-specific extensions

// event api is callback
// errors and events can happen at any time
// therefore the watcher must have some sort of thread
// the thread in blocking listen or channel listen

// Should it be a parl.Go goroutine
// or an object with Shutdown
// that depends on implementation.
