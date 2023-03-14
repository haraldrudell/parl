/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package prand provides a fast and thread-safe random number generation.
//   - prand.Uint32: 2 ns ±0.5
//   - math/rand.Uint32: 14 ns ±0.5
//   - /crypto/rand.Read: 330 ns ±0.5
//   - same methods as math/rand package
//   - based on runtime.fastrand
package prand

import (
	"encoding/binary"
	"unsafe"
)

const (
	bitsPerByte  = 8
	sizeOfUint32 = int(unsafe.Sizeof(uint32(1)))
)

// Uint32 returns a 32-bit unsigned random number using runtime.fastrand. Thread-safe
func Uint32() (random uint32) {
	return fastrand()
}

// Uint32n returns a 32-bit unsigned random number using runtime.fastrand. Thread-safe
func Uint32n(n uint32) (random uint32) {
	return fastrandn(n)
}

// Uint64 returns a 64-bit unsigned random number using runtime.fastrand. Thread-safe
func Uint64() (random uint64) {
	return uint64(fastrand())<<32 | uint64(fastrand())
}

// Int31n returns, as an int32, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
func Int31n(n int32) (i32 int32) {
	if n <= 0 {
		panic("invalid argument to Int31n")
	}
	i32 = int32(fastrandn(uint32(n)))
	return
}

// Read reads n random bytes into p. Thread-Safe.
//   - n always len(p), err always nil
func Read(p []byte) (n int, err error) {
	n = len(p)

	// randomize using 32-bit integers
	index := 0
	lengthMod4 := n &^ (sizeOfUint32 - 1)
	for index < lengthMod4 {
		binary.LittleEndian.PutUint32(p[index:], Uint32())
		index += sizeOfUint32
	}
	if index == n {
		return
	}

	// odd bytes at end
	v := Uint32()
	for index < n {
		p[index] = byte(v)
		index++
		v >>= bitsPerByte
	}

	return
}

// Int63 returns a non-negative pseudo-random 63-bit integer as an int64.
func Int63() (random int64) { return int64(Uint64() >> 1) }

// Int31 returns a non-negative pseudo-random 31-bit integer as an int32.
func Int31() int32 { return int32(Uint32() >> 1) }
