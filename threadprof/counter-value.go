/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

// Package ghandi interfaces Android devices
package threadprof

type CounterValue struct {
	value   uint64
	ops     uint64
	max     uint64
	incRate uint64
	decRate uint64
}

func (cv *CounterValue) Value() (value uint64) {
	return cv.value
}

func (cv *CounterValue) Ops() (ops uint64) {
	return cv.ops
}

func (cv *CounterValue) Max() (ops uint64) {
	return cv.max
}
func (cv *CounterValue) IncRate() (ops uint64) {
	return cv.incRate
}
func (cv *CounterValue) DecRate() (ops uint64) {
	return cv.decRate
}

func (cv *CounterValue) Get() (value uint64, ops uint64, max uint64, incRate uint64, decRate uint64) {
	value = cv.value
	ops = cv.ops
	max = cv.max
	incRate = cv.incRate
	decRate = cv.decRate
	return
}
