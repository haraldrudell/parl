/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"testing"
)

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
		err = errors.Unwrap(err)
	}
	return
}

func TestMap(t *testing.T) {
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

func TestSlice(t *testing.T) {
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
