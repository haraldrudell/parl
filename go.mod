// © 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl

// minimum version of Go required to
// compile packages in this module
//	- use of [iter.Seq] requires go1.23
//	- go1.23 supported 240813–2025-08 upon go1.25 released
//	- go1.23 will be oldest supported Go release 2025-02
go 1.23

require (
	github.com/google/uuid v1.6.0
	golang.org/x/exp v0.0.0-20250210185358-939b2ce775ac
	golang.org/x/sys v0.30.0
	golang.org/x/text v0.22.0
)
