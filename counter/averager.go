/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

type Averager struct {
	values []uint64
	n      int
	size   int
}

func InitAverager(averagerp *Averager, size int) {
	averagerp.values = make([]uint64, size)
	averagerp.size = size
	averagerp.n = 0
}

func (av *Averager) Add(value uint64) (average float64) {
	if av.n < av.size {
		av.values[av.n] = value // slice not full, use another position
		av.n++
	} else {
		copy(av.values, av.values[1:]) // drop the oldest value
		av.values[av.n-1] = value
	}

	var f float64
	for i := 0; i < av.n; i++ {
		f += float64(av.values[i])
	}
	average = f / float64(av.n)
	return
}
