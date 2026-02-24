// © 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl

// minimum version of Go required to
// compile packages in this module
//	- golang.org/x/exp requires go1.25
//	- Ubuntu 24.04.3 has golang-1.23-go go1.23.1 since 240923
//	- go1.25 supported 250812–2026-08 upon go1.27 released
//	- go1.24 supported 250211–2026-02 upon go1.26 released
//	- go1.24 is oldest supported Go release since 250812
go 1.25.0

require (
	github.com/google/uuid v1.6.0
	golang.org/x/exp v0.0.0-20260218203240-3dfff04db8fa
	golang.org/x/sys v0.41.0
	golang.org/x/text v0.34.0
)
