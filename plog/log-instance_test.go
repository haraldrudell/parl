/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package plog

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

func TestLogLogI(t *testing.T) {

	// the code location in printouts is code location Short:
	// parl.TestLogLogI-log-instance_test.go:22
	// it has a line number that  mocksLogI removes
	if false {
		t.Logf("Short: %q", pruntime.NewCodeLocation(0).Short())
		t.FailNow()
	}

	text1, textNewline, expectedLocation, _, writer, lg := mocksLogI()
	// expectedLocation:
	// parl.TestLogLogI-log-instance_test.go
	if false {
		t.Logf("mocksLogI Short: %q", expectedLocation)
		t.FailNow()
	}
	var actualSlice []string
	var actual string

	// Log
	lg.Log(text1)
	actualSlice = writer.getData()
	if len(actualSlice) != 1 || actualSlice[0] != textNewline {
		t.Logf("len actual: %d", len(actualSlice))
		t.Logf(".Log failed: expected: %q actual: %s", textNewline, quoteSliceLogI(actualSlice))
		t.Fail()
	}

	// Log with Location
	lg.SetDebug(true)
	lg.Log(text1 + "\n")
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
		t.Logf("Log SetDebug: no location: actual:\n%q expected:\n%q", actual, expectedLocation)
		t.Fail()
	}
	if strings.Index(actual, "\n") != len(actual)-1 {
		t.Logf("Log SetDebug: newline not at end: actual: %q expected: %q", actual, expectedLocation)
		t.Fail()
	}
}

func TestInfoLogI(t *testing.T) {
	text1, textNewline, _, _, writer, lg := mocksLogI()
	var actualSlice []string

	// Info
	lg.Info(text1)
	actualSlice = writer.getData()
	if len(actualSlice) != 1 || actualSlice[0] != textNewline {
		t.Logf(".Log failed: expected: %q actual: %+v", textNewline, quoteSliceLogI(actualSlice))
		t.Fail()
	}
	if lg.IsSilent() {
		t.Logf("SetSilent default true")
		t.Fail()
	}

	// SetSilent
	lg.SetSilent(true)
	if !lg.IsSilent() {
		t.Logf("SetSilent ineffective")
		t.Fail()
	}
	lg.Info(text1)
	actualSlice = writer.getData()
	if len(actualSlice) != 0 {
		t.Logf("SetSilent true: Info still prints")
		t.Fail()
	}
}

func TestDebugLogI(t *testing.T) {
	text1, textNewline, expectedLocation, _, writer, lg := mocksLogI()
	var actualSlice []string
	var actual string

	// Debug off
	if lg.IsThisDebug() {
		t.Logf("IsThisDebug default true")
		t.Fail()
	}
	lg.Debug(text1)
	actualSlice = writer.getData()
	if len(actualSlice) != 0 {
		t.Logf("Debug prints as default")
		t.Fail()
	}

	// Debug on
	lg.SetDebug(true)
	if !lg.IsThisDebug() {
		t.Logf("IsThisDebug ineffective")
		t.Fail()
	}
	lg.Debug(textNewline)
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

func TestRegexpLogI(t *testing.T) {

	// What is matched against the regexp is FuncName:
	// github.com/haraldrudell/parl.TestRegexpLogI
	// fully qualified package name and function name
	if false {
		t.Logf("FuncName: %q", pruntime.NewCodeLocation(0).FuncName)
		t.FailNow()
	}

	text1, textNewline, expectedLocation, regexpLocation, writer, lg := mocksLogI()
	var actualSlice []string
	var actual string

	nonMatchingRegexp := "aaa"
	matchingRegexp := regexpLocation

	// string that regExp is matched against: "github.com/haraldrudell/parl.TestRegexpLogI"
	//t.Logf("string that regExp is matched against: %q", error116.NewCodeLocation(0).FuncName)

	// matching regexp
	if err := lg.SetRegexp(matchingRegexp); err != nil {
		t.Logf("SetRegexp failed: input: %q err: %+v", matchingRegexp, err)
		t.Fail()
	}
	lg.Debug(textNewline)
	actualSlice = writer.getData()
	if len(actualSlice) != 1 {
		var r = lg.infoRegexp.Load()
		t.Logf("matching regexp did not print 1: %d regexp input:\n%q compiled:\n%+v",
			len(actualSlice),
			matchingRegexp,
			r)
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
		t.Logf("matching regexp: newline not at end: actual: %q expected: %q", actual, regexpLocation)
		t.Fail()
	}

	// non-matching regexp
	if err := lg.SetRegexp(nonMatchingRegexp); err != nil {
		panic(err)
	}
	lg.Debug(text1)
	actualSlice = writer.getData()
	if len(actualSlice) > 0 {
		t.Logf("non-matching regexp did print: %d", len(actualSlice))
		t.Fail()
	}
}

type mockWriterLogI struct {
	lock sync.Mutex
	buf  []string
}

func (w *mockWriterLogI) Write(p []byte) (n int, err error) {
	n = len(p)
	w.lock.Lock()
	defer w.lock.Unlock()
	w.buf = append(w.buf, string(p))
	return
}

func (w *mockWriterLogI) getData() (sList []string) {
	w.lock.Lock()
	defer w.lock.Unlock()
	sList = w.buf
	w.buf = nil
	return
}

func quoteSliceLogI(sList []string) (s string) {
	var s2 []string
	for _, sx := range sList {
		s2 = append(s2, fmt.Sprintf("%q", sx))
	}
	return strings.Join(s2, "\x20")
}

func mocksLogI() (text1, textNewline, expectedLocation, regexpLocation string, writer *mockWriterLogI, lg *LogInstance) {
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
	writer = &mockWriterLogI{}
	lg = NewLog(writer)
	return
}
