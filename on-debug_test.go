/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"io"
	"testing"

	"github.com/haraldrudell/parl/plog"
	"github.com/haraldrudell/parl/pruntime"
)

type TestWriter struct {
	counter AtomicCounter
}

func (w *TestWriter) Write(b []byte) (n int, err error) {
	w.counter.Inc()
	n = len(b)
	return
}

var _ io.Writer = &TestWriter{}

func TestDebugThunk(t *testing.T) {
	var expNoDebug = 0
	var expDebug = 1
	var expRegexp = 1

	var testWriter = TestWriter{}
	var myLogger = plog.NewLog(&testWriter)

	var realStderrLogger = stderrLogger
	stderrLogger = myLogger
	defer func() { stderrLogger = realStderrLogger }()

	// Debug false
	testWriter.counter.Set(0)
	OnDebug(func() string { return "" })
	if c := testWriter.counter.Value(); c != uint64(expNoDebug) {
		t.Errorf("debug false testWriter.counter %d exp %d", c, expNoDebug)
	}

	// Debug true
	SetDebug(true)
	testWriter.counter.Set(0)
	OnDebug(func() string { return "" })
	if c := testWriter.counter.Value(); c != uint64(expDebug) {
		t.Errorf("debug true testWriter.counter %d exp %d", c, expDebug)
	}
	SetDebug(false)

	// regexp debug
	var funcName = pruntime.NewCodeLocation(0).FuncName
	t.Logf("funcName: %q", funcName)
	SetRegexp(funcName)
	testWriter.counter.Set(0)
	OnDebug(func() string { return "" })
	if c := testWriter.counter.Value(); c != uint64(expRegexp) {
		t.Errorf("regexp testWriter.counter %d exp %d", c, expRegexp)
	}
	SetRegexp("")
}
