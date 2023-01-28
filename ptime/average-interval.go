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

// Add adds a datapoint to the averager
func (ai *AverageInterval) Add(value float64) {
	ai.totalLock.Lock()
	defer ai.totalLock.Unlock()

	ai.count++
	ai.total += value
}

// Aggregate provides averaging content for calcukating the average
func (ai *AverageInterval) Aggregate(countp *uint64, floatp *float64) {
	ai.totalLock.RLock()
	defer ai.totalLock.RUnlock()

	*countp += ai.count
	*floatp += ai.total
}
