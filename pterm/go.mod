// © 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl/pterm

go 1.23.1

toolchain go1.24.0

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/haraldrudell/parl v0.4.226
	golang.org/x/term v0.32.0
)

require (
	golang.org/x/exp v0.0.0-20250506013437-ce4c2cf36ca6 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
)
