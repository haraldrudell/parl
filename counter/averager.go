/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import "time"

// Averager is a fixed-sized 64-bit container returning float64 averages per time
type Averager struct {
	values []datapoint
	n      int
	size   int
}

type datapoint struct {
	value    uint64
	duration time.Duration
}

// NewAverager returns an 64-bit averager over size values
func NewAverager(size int) (averager *Averager) {
	return &Averager{
		values: make([]datapoint, size),
		size:   size,
	}
}

// Add adds a new value and computes current average
func (av *Averager) Add(value uint64, duration time.Duration) (average float64) {
	if av.n < av.size {
		// slice not full, use another position
		var dpp = &av.values[av.n]
		dpp.value = value
		dpp.duration = duration
		av.n++
	} else {
		copy(av.values, av.values[1:]) // drop the oldest value
		var dpp = &av.values[av.n-1]
		dpp.value = value
		dpp.duration = duration
	}

	var valueTotal float64
	var durationTotal time.Duration
	for i := 0; i < av.n; i++ {
		valueTotal += float64(av.values[i].value)
		durationTotal += av.values[i].duration
	}
	average = valueTotal / durationTotal.Seconds()
	return
}
