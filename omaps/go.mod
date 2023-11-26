module github.com/haraldrudell/parl/omaps

go 1.21

toolchain go1.21.3

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/google/btree v1.1.2
	github.com/haraldrudell/parl v0.4.132
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa
)

require golang.org/x/text v0.14.0 // indirect
