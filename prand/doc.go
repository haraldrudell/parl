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
