// © 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl/watchfs

go 1.23.0

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/fsnotify/fsnotify v1.8.0
	github.com/google/uuid v1.6.0
	github.com/haraldrudell/parl v0.4.196
)

require (
	golang.org/x/exp v0.0.0-20250218142911-aa4b98e5adaa // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)
