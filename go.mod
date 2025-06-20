// © 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl

// minimum version of Go required to
// compile packages in this module
//	- use of [iter.Seq] requires go1.23
//	- Ubuntu 24.04 has golang-1.23-go go1.23.1 since 240923
//	- go1.24 supported 250211–2026-02 upon go1.26 released
//	- go1.23 supported 240813–2025-08 upon go1.25 released
//	- go1.23 is oldest supported Go release since 250211
go 1.23.1

require (
	github.com/google/uuid v1.6.0
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b
	golang.org/x/sys v0.33.0
	golang.org/x/text v0.26.0
)
