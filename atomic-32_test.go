/*
© 2025-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import (
	"os"
	"testing"
)

func TestAtomic32(t *testing.T) {
	const (
		expZero os.FileMode = 0
		expBoth             = os.ModeDir + os.ModeAppend
	)
	var (
		actual     os.FileMode
		actualBool bool
	)

	// Add() And() CompareAndSwap() Load() Or() Store() Swap()
	var atomicT Atomic32[os.FileMode]

	// Load()
	actual = atomicT.Load()
	if actual != expZero {
		t.Errorf("Load %d exp %d", actual, expZero)
	}

	// Store()
	atomicT.Store(os.ModeDir)
	actual = atomicT.Load()
	if actual != os.ModeDir {
		t.Errorf("Load %d exp %d", actual, os.ModeDir)
	}

	// Swap()
	actual = atomicT.Swap(os.ModeAppend)
	_ = actual
	actual = atomicT.Load()
	if actual != os.ModeAppend {
		t.Errorf("Load %d exp %d", actual, os.ModeAppend)
	}

	// CompareAndSwap
	actualBool = atomicT.CompareAndSwap(os.ModeDir, os.ModeDir)
	if actualBool {
		t.Error("CompareAndSwap true")
	}
	actualBool = atomicT.CompareAndSwap(os.ModeAppend, os.ModeAppend)
	if !actualBool {
		t.Error("CompareAndSwap false")
	}

	// Add()
	actual = atomicT.Add(os.ModeDir)
	if actual != expBoth {
		t.Errorf("Add %d exp %d", actual, expBoth)
	}

	// And()
	actual = atomicT.And(os.ModeDir)
	_ = actual
	actual = atomicT.Load()
	if actual != os.ModeDir {
		t.Errorf("And %d exp %d", actual, os.ModeDir)
	}

	// Or()
	actual = atomicT.Or(os.ModeAppend)
	_ = actual
	actual = atomicT.Load()
	if actual != expBoth {
		t.Errorf("And %d exp %d", actual, expBoth)
	}
}
