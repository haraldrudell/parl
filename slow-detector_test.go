/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"testing"
)

func TestNewSlowDetector(t *testing.T) {

	didPrintf := 0
	printf := func(format string, a ...any) {
		didPrintf++
	}

	var slowDetector *SlowDetector = NewSlowDetectorPrintf("", printf, 0)
	slowDetector.Start()
	slowDetector.Stop()

	if didPrintf != 1 {
		t.Errorf("didPrintf not 1: %d", didPrintf)
	}
}
