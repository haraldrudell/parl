// © 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl/yamler

go 1.24.0

toolchain go1.24.5

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/haraldrudell/parl v0.4.242
	golang.org/x/exp v0.0.0-20251023183803-a4bb9ffd2546
	gopkg.in/yaml.v3 v3.0.1
)

require (
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
)
