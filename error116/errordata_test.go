/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorData(t *testing.T) {
	msg := "message"
	key := "key"
	value := "value"
	err := errors.New(msg)
	err2 := ErrorData(err, DataMap{key: value})

	expected := msg + "\n- " + key + ":\x20" + value

	actual := fmt.Sprintf("%+v", err2)
	if actual != expected {
		t.Fail()
	}
	actual = fmt.Sprintf("%-v", err2)
	if actual != expected {
		t.Fail()
	}

	expected = msg
	actual = fmt.Sprintf("%v", err2)
	if actual != expected {
		t.Fail()
	}
	actual = fmt.Sprintf("%s", err2)
	if actual != expected {
		t.Fail()
	}
	actual = fmt.Sprintf("%q", err2)
	if actual != "\x22"+expected+"\x22" {
		t.Fail()
	}
}
