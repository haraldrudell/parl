/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"testing"
	"time"
)

func TestNewAverager(t *testing.T) {
	value1 := time.Second
	value2 := time.Minute
	exp := (value1 + value2) / 2

	var a *Averager[time.Duration]
	var b float64
	var duration time.Duration

	a = NewAverager[time.Duration]()
	a.Add(value1)
	a.Add(value2)

	b, _ = a.Average()
	duration = time.Duration(b)
	if duration != exp {
		t.Errorf("average: %s exp %s", duration, exp)
	}
}
