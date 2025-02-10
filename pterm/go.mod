// © 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl/pterm

go 1.22.0

toolchain go1.23.2

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/haraldrudell/parl v0.4.188
	golang.org/x/term v0.29.0
)

require (
	golang.org/x/exp v0.0.0-20250207012021-f9890c6ad9f3 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)
