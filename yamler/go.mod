module github.com/haraldrudell/parl/yamler

go 1.21

replace github.com/haraldrudell/parl => ../../parl

replace github.com/haraldrudell/parl/mains => ../mains

require (
	github.com/haraldrudell/parl v0.4.112
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/google/btree v1.1.2 // indirect
	golang.org/x/text v0.13.0 // indirect
)
