/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"testing"
	"time"
	_ "unsafe"
)

func TestCpuTicks(t *testing.T) {
	//t.Fail()

	var t0 = time.Now()
	var a = CpuTicks() // uint32
	time.Sleep(time.Microsecond)
	var t1 = time.Now()
	var b = CpuTicks()
	var diff = int(b) - int(a)
	// time difference is more than 1,000 ns

	// CpuTicks: 3073095747 3073101414 change: 5667 duration per tick: 9.925887e-10
	t.Logf("CpuTicks: %d %d change: %d duration per tick s: %e", a, b, diff, float64(t1.Sub(t0))/1e9/float64(diff))

	if a == b {
		t.Errorf("CputTicks not advancing: %d %d", a, b)
	}
}
