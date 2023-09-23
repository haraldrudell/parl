/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"sync"
)

// AverageInterval is a container for the datapoint aggregate over a single period
type AverageInterval struct {
	index PeriodIndex

	totalLock sync.RWMutex
	count     uint64
	total     float64
}

// NewAverageInterval returns an avareging container for one averaging interval
func NewAverageInterval(index PeriodIndex) (ai *AverageInterval) {
	return &AverageInterval{index: index}
}

func (a *AverageInterval) Index() (index PeriodIndex) {
	return a.index
}

// Add adds a datapoint to the averager
//   - count of values incremented
//   - value added to total
func (a *AverageInterval) Add(value float64) {
	a.totalLock.Lock()
	defer a.totalLock.Unlock()

	a.count++
	a.total += value
}

// Aggregate provides averaging content for calculating the average
func (a *AverageInterval) Aggregate(countp *uint64, floatp *float64) {
	a.totalLock.RLock()
	defer a.totalLock.RUnlock()

	*countp += a.count
	*floatp += a.total
}
