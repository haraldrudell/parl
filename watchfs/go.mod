module github.com/haraldrudell/parl/watchfs

go 1.19

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/fsnotify/fsnotify v1.6.0
	github.com/google/uuid v1.3.0
	github.com/haraldrudell/parl v0.4.71
)

require (
	golang.org/x/exp v0.0.0-20230315142452-642cacee5cc0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
)
