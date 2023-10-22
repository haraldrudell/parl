/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/plog"
	"github.com/haraldrudell/parl/pruntime"
)

func TestLogLog(t *testing.T) {
	// reset of static loggings logInstance object
	defer func(stderrLogger0 *plog.LogInstance) {
		stderrLogger = stderrLogger0
	}(stderrLogger)
	defer SetDebug(false)

	text1, textNewline, expectedLocation, _, writer, _ := mocksLogStat()
	stderrLogger = plog.NewLogFrames(writer, 1)

	var actualSlice []string
	var actual string

	// Log
	Log(text1)
	actualSlice = writer.getData()
	if len(actualSlice) != 1 || actualSlice[0] != textNewline {
		t.Logf(".Log failed: expected: %q actual: %s", textNewline, quoteSliceLogStat(actualSlice))
		t.Fail()
	}

	// Log with Location
	SetDebug(true)
	Log(text1 + "\n")
	actualSlice = writer.getData()
	if len(actualSlice) != 1 {
		t.Logf("Log SetDebug: invocations not 1: %d", len(actualSlice))
	} else {
		actual = actualSlice[0]
	}
	if !strings.HasPrefix(actual, text1) {
		t.Logf("Log SetDebug: no text1 prefix: %q", actual)
		t.Fail()
	}
	if !strings.Contains(actual, expectedLocation) {
		t.Logf("Log SetDebug: no location: actual: %q expected: %q", actual, expectedLocation)
		t.Fail()
	}
	if strings.Index(actual, "\n") != len(actual)-1 {
		t.Logf("Log SetDebug: newline not at end: actual: %q expected: %q", actual, expectedLocation)
		t.Fail()
	}
}

func TestInfoLog(t *testing.T) {
	defer func(stderrLogger0 *plog.LogInstance) {
		stderrLogger = stderrLogger0
	}(stderrLogger)
	defer SetSilent(false)

	text1, textNewline, _, _, writer, _ := mocksLogStat()
	stderrLogger = plog.NewLogFrames(writer, 1)

	var actualSlice []string

	// Info
	Info(text1)
	actualSlice = writer.getData()
	if len(actualSlice) != 1 || actualSlice[0] != textNewline {
		t.Logf(".Log failed: expected:\n%q actual:\n%+v", textNewline, quoteSliceLogStat(actualSlice))
		t.Fail()
	}
	if IsSilent() {
		t.Logf("SetSilent default true")
		t.Fail()
	}

	// SetSilent
	SetSilent(true)
	if !IsSilent() {
		t.Logf("SetSilent ineffective")
		t.Fail()
	}
	Info(text1)
	actualSlice = writer.getData()
	if len(actualSlice) != 0 {
		t.Logf("SetSilent true: Info still prints")
		t.Fail()
	}
}

func TestDebugLog(t *testing.T) {
	defer func(stderrLogger0 *plog.LogInstance) {
		stderrLogger = stderrLogger0
	}(stderrLogger)
	defer SetDebug(false)

	text1, textNewline, expectedLocation, _, writer, _ := mocksLogStat()
	stderrLogger = plog.NewLogFrames(writer, 1)

	var actualSlice []string
	var actual string

	// Debug off
	if IsThisDebug() {
		t.Logf("IsThisDebug default true")
		t.Fail()
	}
	Debug(text1)
	actualSlice = writer.getData()
	if len(actualSlice) != 0 {
		t.Logf("Debug prints as default")
		t.Fail()
	}

	// Debug on
	SetDebug(true)
	if !IsThisDebug() {
		t.Logf("IsThisDebug ineffective")
		t.Fail()
	}
	Debug(textNewline)
	actualSlice = writer.getData()
	if len(actualSlice) != 1 {
		t.Logf("Log SetDebug: invocations not 1: %d", len(actualSlice))
		t.FailNow()
	}
	actual = actualSlice[0]
	if !strings.HasPrefix(actual, text1) {
		t.Logf("Log SetDebug: no text1 prefix: %q", actual)
		t.Fail()
	}
	if !strings.Contains(actual, expectedLocation) {
		t.Logf("Log SetDebug: no location: actual: %q expected: %q", actual, expectedLocation)
		t.Fail()
	}
}

func TestRegexpLog(t *testing.T) {
	defer func(stderrLogger0 *plog.LogInstance) {
		stderrLogger = stderrLogger0
	}(stderrLogger)
	defer SetRegexp("")

	text1, textNewline, expectedLocation, regexpLocation, writer, _ := mocksLogStat()
	stderrLogger = plog.NewLogFrames(writer, 1)

	matchingRegexp := regexpLocation
	nonMatchingRegexp := "aaa"

	var actualSlice []string
	var actual string

	// matching regexp
	if err := SetRegexp(matchingRegexp); err != nil {
		t.Logf("SetRegexp failed: input: %q err: %+v", matchingRegexp, err)
		t.Fail()
	}
	Debug(textNewline)
	actualSlice = writer.getData()
	if len(actualSlice) != 1 {
		t.Logf("matching regexp did not print 1: %d regexp input:\n%q",
			len(actualSlice),
			matchingRegexp)
		t.Fail()
	}
	actual = actualSlice[0]
	if !strings.HasPrefix(actual, text1) {
		t.Logf("matching regexp: missing prefix: %q text: %q", text1, actual)
		t.Fail()
	}
	if !strings.Contains(actual, expectedLocation) {
		t.Logf("matching regexp: no location: actual:\n%q expected:\n%q",
			actual,
			expectedLocation)
		t.Fail()
	}
	if strings.Index(actual, "\n") != len(actual)-1 {
		t.Logf("matching regexp: newline not at end: actual: %q expected: %q", actual, expectedLocation)
		t.Fail()
	}

	// non-matching regexp
	if err := SetRegexp(nonMatchingRegexp); err != nil {
		panic(err)
	}
	Debug(text1)
	actualSlice = writer.getData()
	if len(actualSlice) > 0 {
		t.Logf("non-matching regexp did print: %d", len(actualSlice))
		t.Fail()
	}
}

type mockWriterLogStat struct {
	lock sync.Mutex
	buf  []string
}

func (w *mockWriterLogStat) Write(p []byte) (n int, err error) {
	n = len(p)
	w.lock.Lock()
	defer w.lock.Unlock()
	w.buf = append(w.buf, string(p))
	return
}

func (w *mockWriterLogStat) getData() (sList []string) {
	w.lock.Lock()
	defer w.lock.Unlock()
	sList = w.buf
	w.buf = nil
	return
}

func quoteSliceLogStat(sList []string) (s string) {
	var s2 []string
	for _, sx := range sList {
		s2 = append(s2, fmt.Sprintf("%q", sx))
	}
	return strings.Join(s2, "\x20")
}

func mocksLogStat() (text1, textNewline, expectedLocation, regexpLocation string, writer *mockWriterLogStat, mockOutput func(n int, s string) (err error)) {
	text1 = "abc"
	textNewline = text1 + "\n"

	// location text for this file
	location := pruntime.NewCodeLocation(1)
	expectedLocation = location.Short()
	// remove line number since this changes
	if index := strings.Index(expectedLocation, ":"); index == -1 {
		panic(perrors.Errorf("error116.NewCodeLocation failed: %q", expectedLocation))
	} else {
		expectedLocation = expectedLocation[0:index]
	}
	regexpLocation = location.FuncName
	writer = &mockWriterLogStat{}
	mockOutput = func(n int, s string) (err error) {
		if !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		_, err = writer.Write([]byte(s))
		return
	}
	return
}

func TestIsThisDebugN(t *testing.T) {
	SetDebug(false)

	// matching Regexp should…
	SetRegexp("TestIsThisDebugN")
	if !IsThisDebug() {
		t.Error("IsThisDebug false")
	}
	if !IsThisDebugN(0) {
		t.Error("IsThisDebugN(0) false")
	}
	if IsThisDebugN(1) {
		t.Error("IsThisDebugN(1) true")
	}

	// no debug should…
	SetRegexp("")
	if IsThisDebug() {
		t.Error("IsThisDebug true")
	}
	if IsThisDebugN(0) {
		t.Error("IsThisDebugN(0) true")
	}
	if IsThisDebugN(1) {
		t.Error("IsThisDebugN(1) true")
	}
}
