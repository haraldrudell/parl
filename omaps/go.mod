module github.com/haraldrudell/parl/omaps

go 1.21

toolchain go1.21.3

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/google/btree v1.1.2
	github.com/haraldrudell/parl v0.4.135
	golang.org/x/exp v0.0.0-20231127185646-65229373498e
)

require golang.org/x/text v0.14.0 // indirect
