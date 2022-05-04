/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package threadprof

type CounterValue struct {
	value   uint64
	running uint64
	max     uint64
	incRate uint64
	decRate uint64
}

func (cv *CounterValue) Value() (value uint64) {
	return cv.value
}

func (cv *CounterValue) Running() (ops uint64) {
	return cv.running
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

func (cv *CounterValue) Get() (value uint64, running uint64, max uint64, incRate uint64, decRate uint64) {
	value = cv.value
	running = cv.running
	max = cv.max
	incRate = cv.incRate
	decRate = cv.decRate
	return
}
