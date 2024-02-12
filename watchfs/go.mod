// © 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl/watchfs

go 1.21

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/fsnotify/fsnotify v1.7.0
	github.com/google/uuid v1.4.0
	github.com/haraldrudell/parl v0.4.161
)

require (
	golang.org/x/exp v0.0.0-20240119083558-1b970713d09a // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)
