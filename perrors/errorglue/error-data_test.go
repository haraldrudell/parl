/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"reflect"
	"testing"
)

func TestErrorData(t *testing.T) {
	//t.Errorf("Logging on")
	var key, value = "key", "value"
	var message = "error-message"
	var eData *errorData
	var typeName = reflect.TypeOf(eData).String()
	var m = map[CSFormat]string{
		DefaultFormat: message,
		CodeLocation:  "",
		ShortFormat:   message,
		LongFormat:    message + " [" + typeName + "]\n" + key + ": " + value,
		ShortSuffix:   "",
		LongSuffix:    key + ": " + value,
	}

	var formatAct, formatExp, keyAct, valueAct string
	var ok bool
	var err error
	var err0 = errors.New(message)
	var hasData ErrorHasData

	// ChainString() KeyValue()
	var _ *errorData

	err = NewErrorData(err0, key, value)

	// err should be ErrorHasData
	hasData, ok = err.(ErrorHasData)
	if !ok {
		t.Fatalf("NewErrorData not ErrorHasData")
	}

	// err should be eData
	eData, ok = err.(*errorData)
	if !ok {
		t.Fatalf("NewErrorData not ErrorHasData")
	}

	// KeyValue should match
	keyAct, valueAct = hasData.KeyValue()
	if keyAct != key {
		t.Errorf("key %q exp %q", keyAct, key)
	}
	if valueAct != value {
		t.Errorf("value %q exp %q", valueAct, value)
	}

	for _, csFormat := range csFormatList {
		if formatExp, ok = m[csFormat]; !ok {
			t.Errorf("no expected value for format: %s", csFormat)
		}
		formatAct = eData.ChainString(csFormat)

		// DefaultFormat: "error-message"
		// CodeLocation: ""
		// ShortFormat: "error-message"
		//  LongFormat: "error-message [*errorglue.errorData]\nkey: value"
		// ShortSuffix: ""
		// LongSuffix: "key: value"
		t.Logf("%s: %q", csFormat, formatAct)

		// ChainString should match
		if formatAct != formatExp {
			t.Errorf("FAIL %s:\n%q exp\n%q",
				csFormat, formatAct, formatExp,
			)
		}
	}
}

func TestErrorDataMap(t *testing.T) {
	k1 := "k1"
	v1 := "s1"
	k2 := "k2"
	v2 := "s2"
	e1error := "e1"

	var expectedInt int

	e1 := errors.New(e1error)
	e2 := NewErrorData(e1, k1, v1)
	e3 := NewErrorData(e2, k2, v2)
	strs, values := eErrorData(e3)
	expectedInt = len(strs)
	if expectedInt > 0 {
		t.Errorf("slice has unexpected values: %d", expectedInt)
	}
	if len(values) != 2 || values[k1] != v1 || values[k2] != v2 {
		t.Errorf("bad map: %#v expected: %#v", values, map[string]string{k1: v1, k2: v2})
	}
}

func TestErrorDataSlice(t *testing.T) {
	s1 := "s1"
	s2 := "s2"
	e1error := "e1"

	var expectedInt int

	e1 := errors.New(e1error)
	e2 := NewErrorData(e1, "", s1)
	e3 := NewErrorData(e2, "", s2)
	strs, values := eErrorData(e3)
	expectedInt = len(values)
	if expectedInt > 0 {
		t.Errorf("map has unexpected values: %d", expectedInt)
	}
	if len(strs) != 2 || strs[0] != s1 || strs[1] != s2 {
		t.Errorf("bad slice: %#v expected: %#v", strs, []string{s1, s2})
	}
}

func eErrorData(err error) (list []string, keyValues map[string]string) {
	for err != nil {
		if e, ok := err.(ErrorHasData); ok {
			key, value := e.KeyValue()
			if key == "" { // for the slice
				list = append([]string{value}, list...)
			} else { // for the map
				if keyValues == nil {
					keyValues = map[string]string{key: value}
				} else if _, ok := keyValues[key]; !ok {
					keyValues[key] = value
				}
			}
		}
		err, _, _ = Unwrap(err)
	}
	return
}
