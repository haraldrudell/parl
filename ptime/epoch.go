/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import "time"

var ptimeEpoch = time.Now()

type Epoch time.Duration

// Epoch translates a time value to a 64-bit value that can be used atomically
func EpochNow(t ...time.Time) (epoch Epoch) {
	var t0 time.Time
	if len(t) > 0 {
		t0 = t[0]
	} else {
		t0 = time.Now()
	}
	if t0.IsZero() {
		return // zero value: epoch 0
	}
	return Epoch(t0.Sub(ptimeEpoch))
}

func (epoch Epoch) Time() (t time.Time) {
	if epoch == 0 {
		return // epoch zero means time.Time{} ie. time.IsZero()
	}
	return ptimeEpoch.Add(time.Duration(epoch))
}

func (epoch Epoch) IsValid() (isValid bool) {
	return epoch != 0
}
