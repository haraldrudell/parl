/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"testing"
)

func TestCachedLocation_PackFunc(t *testing.T) {
	var c0 CachedLocation
	// expCL NewCodeLocation is on same line as first invocation of c0
	var expCL = NewCodeLocation(len(c0.Short()) * 0)

	if c0.FuncIdentifier() != expCL.FuncIdentifier() {
		t.Errorf("FuncIdentifier: %q exp %q", c0.FuncIdentifier(), expCL.FuncIdentifier())
	}
	if c0.FuncIdentifier() == "" {
		t.Errorf("FuncIdentifier: empty")
	}
	if c0.PackFunc() != expCL.PackFunc() {
		t.Errorf("FuncName: %q exp %q", c0.PackFunc(), expCL.PackFunc())
	}
	if c0.Short() != expCL.Short() {
		t.Errorf("Short: %q exp %q", c0.Short(), expCL.Short())
	}
	if c0.FuncName() != expCL.FuncName {
		t.Errorf("FuncName: %q exp %q", c0.FuncName(), expCL.FuncName)
	}

	// FuncIdentifier: TestCachedLocation_PackFunc
	// PackFunc: pruntime.TestCachedLocation_PackFunc
	// Short: pruntime.TestCachedLocation_PackFunc()-cached-location_test.go:15
	// FuncName: github.com/haraldrudell/parl/pruntime.TestCachedLocation_PackFunc
	t.Logf("FuncIdentifier: %s\nPackFunc: %s\nShort: %s\nFuncName: %s\n",
		c0.FuncIdentifier(),
		c0.PackFunc(),
		c0.Short(),
		c0.FuncName(),
	)
	//t.Fail()
}
