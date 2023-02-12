module github.com/haraldrudell/parl/watchfs

go 1.19

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/fsnotify/fsnotify v1.6.0
	github.com/google/uuid v1.3.0
	github.com/haraldrudell/parl v0.4.56
)

require (
	golang.org/x/exp v0.0.0-20230210204819-062eb4c674ab // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
)
