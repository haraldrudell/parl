module github.com/haraldrudell/parl/watchfs

go 1.21

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/fsnotify/fsnotify v1.7.0
	github.com/google/uuid v1.4.0
	github.com/haraldrudell/parl v0.4.128
)

require (
	github.com/google/btree v1.1.2 // indirect
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)
