module github.com/haraldrudell/parl/watchfs

go 1.19

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/fsnotify/fsnotify v1.6.0
	github.com/google/uuid v1.3.0
	github.com/haraldrudell/parl v0.4.90
)

require (
	golang.org/x/exp v0.0.0-20230515195305-f3d0a9c9a5cc // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
)
