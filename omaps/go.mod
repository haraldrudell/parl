// © 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl/omaps

go 1.23.1

toolchain go1.24.0

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/google/btree v1.1.3
	github.com/haraldrudell/parl v0.4.223
	golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0
)

require golang.org/x/text v0.24.0 // indirect
