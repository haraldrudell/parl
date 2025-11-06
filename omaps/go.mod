// © 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl/omaps

go 1.24.0

toolchain go1.24.5

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/google/btree v1.1.3
	github.com/haraldrudell/parl v0.4.242
	golang.org/x/exp v0.0.0-20251023183803-a4bb9ffd2546
)

require golang.org/x/text v0.30.0 // indirect
