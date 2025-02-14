// © 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl/pterm

go 1.23

toolchain go1.23.5

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/haraldrudell/parl v0.4.192
	golang.org/x/term v0.29.0
)

require (
	golang.org/x/exp v0.0.0-20250210185358-939b2ce775ac // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)
