// © 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl/watchfs

go 1.21

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/fsnotify/fsnotify v1.7.0
	github.com/google/uuid v1.4.0
	github.com/haraldrudell/parl v0.4.162
)

require (
	golang.org/x/exp v0.0.0-20240205201215-2c58cdc269a3 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)
