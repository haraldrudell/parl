/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptesting

import (
	"math"
)

// Uint32 returns a 32-bit unsigned random number using runtime.fastrand. Thread-safe
func Uint32() (random uint32) {
	return fastrand()
}

// Uint32n returns a 32-bit unsigned random number using runtime.fastrand. Thread-safe
func Uint32n(n uint32) (random uint32) {
	return fastrandn(n)
}

// Uint32 returns a 64-bit unsigned random number using runtime.fastrand. Thread-safe
func Uint64() (random uint64) {
	return uint64(fastrand())<<32 | uint64(fastrand())
}

// Shuffle pseudo-randomizes the order of elements.
// n is the number of elements. Shuffle panics if n < 0.
// swap swaps the elements with indexes i and j. Thread-safe.
//   - from: math/rand.Shuffle
func Shuffle(n int, swap func(i, j int)) {
	if n < 0 {
		panic("invalid argument to Shuffle")
	}

	// Fisher-Yates shuffle: https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
	// Shuffle really ought not be called with n that doesn't fit in 32 bits.
	// Not only will it take a very long time, but with 2³¹! possible permutations,
	// there's no way that any PRNG can have a big enough internal state to
	// generate even a minuscule percentage of the possible permutations.
	// Nevertheless, the right API signature accepts an int n, so handle it as best we can.
	i := n - 1
	for ; i > 1<<31-1-1; i-- {
		j := int(Int63n(int64(i + 1)))
		swap(i, j)
	}
	for ; i > 0; i-- {
		j := int(Int31n(int32(i + 1)))
		swap(i, j)
	}
}

// Int63n returns, as an int64, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0. Thread-safe.
//   - from: math/rand.Int63n
func Int63n(n int64) int64 {
	if n <= 0 {
		panic("invalid argument to Int63n")
	}
	return int64(Uint64n(uint64(n)))
}

// Int31n returns, as an int32, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
func Int31n(n int32) int32 {
	if n <= 0 {
		panic("invalid argument to Int31n")
	}
	var u32 uint32 = fastrandn(uint32(n))
	return int32(u32)
}

// Uint64n returns, as a uint64, a pseudo-random number in [0,n).
// It is guaranteed more uniform than taking a Source value mod n
// for any n that is not a power of 2. Thread-safe.
//   - from: math/rand.Rand.Uint64n
func Uint64n(n uint64) uint64 {
	if n&(n-1) == 0 { // n is power of two, can mask
		if n == 0 {
			panic("invalid argument to Uint64n")
		}
		return Uint64() & (n - 1)
	}
	// If n does not divide v, to avoid bias we must not use
	// a v that is within maxUint64%n of the top of the range.
	v := Uint64()
	if v > math.MaxUint64-n { // Fast check.
		ceiling := math.MaxUint64 - math.MaxUint64%n
		for v >= ceiling {
			v = Uint64()
		}
	}

	return v % n
}
