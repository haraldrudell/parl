/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"testing"
)

func TestCachedLocation(t *testing.T) {
	var c CachedLocation
	var exp string

	// initial Short should match CodeLocation
	var cL, act = NewCodeLocation(0), c.Short()
	if exp = cL.Short(); act != exp {
		t.Errorf("initial Short: %q exp %q", act, exp)
	}

	// cached Short should match Code Location
	if act, exp = c.Short(), cL.Short(); act != exp {
		t.Errorf("cached Short: %q exp %q", act, exp)
	}

	// FuncIdentifier
	if act, exp = c.FuncIdentifier(), cL.FuncIdentifier(); act != exp {
		t.Errorf("FuncIdentifier: %q exp %q", act, exp)
	}

	// FuncName
	if act, exp = c.FuncName(), cL.FuncName; act != exp {
		t.Errorf("FuncName: %q exp %q", act, exp)
	}

	// PackFunc
	if act, exp = c.PackFunc(), cL.PackFunc(); act != exp {
		t.Errorf("PackFunc: %q exp %q", act, exp)
	}
}
