/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"testing"
	"time"
)

const testDuration = time.Second

func TestClosingTicker(t *testing.T) {
	ticker := NewClosingTicker(testDuration)
	ticker.Shutdown()
	_, ok := <-ticker.C
	if ok {
		t.Logf("ticker.C did not close on ticker.Shutdown")
		t.Fail()
	}
}
