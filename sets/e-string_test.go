/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import "testing"

func TestEpString(t *testing.T) {
	//t.Fail()
	var s, e string

	s = eString[*int](nil)
	e = "<nil>"
	t.Logf("epString(nil) → ‘%v’", s)
	if s != e {
		t.Errorf("%q exp %q", s, e)
	}

	var i = 1
	s = eString(i)
	e = "1"
	t.Logf("epString(1) → ‘%v’", s)
	if s != e {
		t.Errorf("%q exp %q", s, e)
	}

	s = eString(&i)
	e = "1"
	t.Logf("epString(&1) → ‘%v’", s)
	if s != e {
		t.Errorf("%q exp %q", s, e)
	}

	var text = "abc"
	s = eString(&text)
	e = "abc"
	t.Logf("epString(&abc) → ‘%v’", s)
	if s != e {
		t.Errorf("%q exp %q", s, e)
	}

	var h hasString
	s = eString(h)
	e = "string"
	t.Logf("epString((*) String()) → ‘%v’", s)
	if s != e {
		t.Errorf("%q exp %q", s, e)
	}

	s = eString(&h)
	e = "string"
	t.Logf("epString(&(*) String()) → ‘%v’", s)
	if s != e {
		t.Errorf("%q exp %q", s, e)
	}

	var v valueString
	s = eString(v)
	e = "string"
	t.Logf("epString(() String()) → ‘%v’", s)
	if s != e {
		t.Errorf("%q exp %q", s, e)
	}

	s = eString(&v)
	e = "string"
	t.Logf("epString(&() String()) → ‘%v’", s)
	if s != e {
		t.Errorf("%q exp %q", s, e)
	}
}

type hasString struct{}

func (h *hasString) String() (s string) {
	return "string"
}

type valueString struct{}

func (h valueString) String() (s string) {
	return "string"
}
