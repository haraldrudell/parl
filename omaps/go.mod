// © 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl/omaps

go 1.21

toolchain go1.21.3

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/google/btree v1.1.2
	github.com/haraldrudell/parl v0.4.156
	golang.org/x/exp v0.0.0-20240103183307-be819d1f06fc
)

require golang.org/x/text v0.14.0 // indirect
