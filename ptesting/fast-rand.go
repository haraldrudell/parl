/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptesting

import "golang.org/x/exp/rand"

type fastRand struct{}

var _ rand.Source = &fastRand{}

// NewFastRand returns a fast random number generator
//   - not thread-safe
//   - cannot be Seeded
func NewFastRand() (generator *rand.Rand) {
	return rand.New(&fastRand{})
}

func (f fastRand) Seed(seed uint64) {}
func (f fastRand) Uint64() (random uint64) {
	return uint64(fastrand())<<32 | uint64(fastrand())
}
