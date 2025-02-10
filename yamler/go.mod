// © 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl/yamler

go 1.22.0

toolchain go1.23.2

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/haraldrudell/parl v0.4.188
	golang.org/x/exp v0.0.0-20250207012021-f9890c6ad9f3
	gopkg.in/yaml.v3 v3.0.1
)

require (
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)
