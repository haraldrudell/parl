/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
)

func TestErrorWithStack(t *testing.T) {
	msg := "message"
	pc, filename, line, ok := runtime.Caller(0)
	if !ok {
		t.Log("runtime.Caller failed")
		t.Fail()
	}
	funcP := runtime.FuncForPC(pc)
	if funcP == nil {
		t.Log("runtime.FuncForPC failed")
		t.Fail()
	}
	funcName := funcP.Name()
	shortStack := fmt.Sprintf("%s at %s-%s:", msg, filepath.Base(funcName), filepath.Base(filename))
	longStack := fmt.Sprintf("%s\n%s\n  %s:", msg, funcName, filename)
	quote := "\x22"
	formats := []Format{DefaultFormat, LongFormat, LongSuffix, ShortFormat, ShortSuffix}
	formatMap := map[Format]string{DefaultFormat: "default", LongFormat: "long",
		LongSuffix: "longS", ShortFormat: "short", ShortSuffix: "shortS"}
	key := "key"
	value := "value"
	dataMap := DataMap{key: value}

	_ = line
	_ = shortStack
	_ = longStack
	_ = quote

	err := New(msg)
	for _, format := range formats {
		t.Logf("%s: %s\n", formatMap[format], ChainString(err, format))
	}

	err = Errorf("errors: '%w'", ErrorData(errors.New(msg), dataMap))
	for _, format := range formats {
		t.Logf("%s: %s\n", formatMap[format], ChainString(err, format))
	}
	/*
		t.Logf(ChainString(err, Def))
		t.Logf(ChainString(err, Def))

		actual := fmt.Sprintf("%+v", err2)
		expected := fmt.Sprintf("%s\n%s\n  %s:%d\n", msg, funcName, filename, line-1)
		if actual[:len(expected)] != expected {
			s := "Stack(err) %+v failed"
			t.Log(s)
			t.Fail()
		}

		actual = fmt.Sprintf("%-v", err2)
		expected =
		if actual != expected {
			s := "Stack(err) %-v failed"
			t.Log(s)
			t.Fail()
		}

		actual = fmt.Sprintf("%v", err2)
		if actual != msg {
			s := "Stack(err) %v failed"
			t.Log(s)
			t.Fail()
		}

		actual = fmt.Sprintf("%s", err2)
		if actual != msg {
			s := "Stack(err) %s failed"
			t.Log(s)
			t.Fail()
		}

		actual = fmt.Sprintf("%q", err2)
		if actual != "\x22"+msg+"\x22" {
			s := "Stack(err) %q failed"
			t.Log(s)
			t.Fail()
		}
	*/
}
