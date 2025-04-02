// © 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl/yamler

go 1.23.1

toolchain go1.24.0

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/haraldrudell/parl v0.4.213
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394
	gopkg.in/yaml.v3 v3.0.1
)

require (
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
)
