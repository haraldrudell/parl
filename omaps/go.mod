// © 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl/omaps

go 1.23.0

toolchain go1.24.0

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/google/btree v1.1.3
	github.com/haraldrudell/parl v0.4.194
	golang.org/x/exp v0.0.0-20250215185904-eff6e970281f
)

require golang.org/x/text v0.22.0 // indirect
