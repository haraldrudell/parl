// © 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl/omaps

go 1.23.0

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/google/btree v1.1.3
	github.com/haraldrudell/parl v0.4.198
	golang.org/x/exp v0.0.0-20250218142911-aa4b98e5adaa
)

require golang.org/x/text v0.22.0 // indirect
