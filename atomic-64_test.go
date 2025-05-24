/*
© 2025-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import (
	"testing"
	"time"
)

func TestAtomic64(t *testing.T) {
	const (
		expZero time.Duration = 0
		expBoth               = time.Second + time.Minute
		three   time.Duration = 3
		one     time.Duration = 1
		two     time.Duration = 2
	)
	var (
		actual     time.Duration
		actualBool bool
	)

	// Add() And() CompareAndSwap() Load() Or() Store() Swap()
	var atomicT Atomic64[time.Duration]

	// Load()
	actual = atomicT.Load()
	if actual != expZero {
		t.Errorf("Load %d exp %d", actual, expZero)
	}

	// Store()
	atomicT.Store(time.Second)
	actual = atomicT.Load()
	if actual != time.Second {
		t.Errorf("Load %d exp %d", actual, time.Second)
	}

	// Swap()
	actual = atomicT.Swap(time.Minute)
	_ = actual
	actual = atomicT.Load()
	if actual != time.Minute {
		t.Errorf("Load %d exp %d", actual, time.Minute)
	}

	// CompareAndSwap
	actualBool = atomicT.CompareAndSwap(time.Second, time.Second)
	if actualBool {
		t.Error("CompareAndSwap true")
	}
	actualBool = atomicT.CompareAndSwap(time.Minute, time.Minute)
	if !actualBool {
		t.Error("CompareAndSwap false")
	}

	// Add()
	actual = atomicT.Add(time.Second)
	if actual != expBoth {
		t.Errorf("Add %d exp %d", actual, expBoth)
	}

	// And()
	atomicT.Store(three)
	actual = atomicT.And(one)
	_ = actual
	actual = atomicT.Load()
	if actual != one {
		t.Errorf("And %d exp %d", actual, one)
	}

	// Or()
	actual = atomicT.Or(two)
	_ = actual
	actual = atomicT.Load()
	if actual != three {
		t.Errorf("And %d exp %d", actual, three)
	}
}
