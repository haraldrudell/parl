/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "sync/atomic"

var (
	// NoSTReader indicates that a ShouldTerminateReader argument is absent
	NoSTReader ShouldTerminateReader
	// NoSTWriter indicates that a ShouldTerminateWriter argument is absent
	NoSTWriter ShouldTerminateWriter
)

// ShouldTerminateReader allows an [GoGroup] or []
// error channel reader to detect whether the application
// terminate
type ShouldTerminateReader interface {
	// IsTerminate returns true if a prominent thread is about to exit
	//	- this signals for the app to exit
	IsTerminate() (isTerminate bool)
}

type ShouldTerminateWriter interface {
	// SetTerminate indicates that a prominent thread is about to exit
	//	- this signals for the app to exit
	SetTerminate()
}

// ShouldTerminate allows a main thread to terminate
// the thread-group or app upon its exit
//   - initialization-free
//   - provided as [ShouldTerminateWriter] [ShouldTerminateReader
type ShouldTerminate struct {
	// atomicBool true signals for the app to exit on next thread-exit
	atomicBool atomic.Bool
}

// ShouldTerminate is ShouldTerminateWriter
var _ ShouldTerminateWriter = &ShouldTerminate{}

// ShouldTerminate is ShouldTerminateReader
var _ ShouldTerminateReader = &ShouldTerminate{}

// SetTerminate indicates that a prominent thread is about to exit
//   - this signals for the app to exit
func (s *ShouldTerminate) SetTerminate() {
	if s.atomicBool.Load() {
		return
	}
	s.atomicBool.Store(true)
}

// IsTerminate returns true if a prominent thread is about to exit
//   - this signals for the app to exit
func (s *ShouldTerminate) IsTerminate() (isTerminate bool) { return s.atomicBool.Load() }
