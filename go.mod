// © 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl

// minimum version of Go required to
// compile packages in this module
//	- use of [iter.Seq] requires go1.23
//	- go1.24 supported 250211–2026-02 upon go1.26 released
//	- go1.23 supported 240813–2025-08 upon go1.25 released
//	- go1.23 is oldest supported Go release since 250211
go 1.23.0

require (
	github.com/google/uuid v1.6.0
	golang.org/x/exp v0.0.0-20250218142911-aa4b98e5adaa
	golang.org/x/sys v0.30.0
	golang.org/x/text v0.22.0
)
