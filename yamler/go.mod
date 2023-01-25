module github.com/haraldrudell/parl/yamler

go 1.19

replace github.com/haraldrudell/parl => ../../parl

replace github.com/haraldrudell/parl/mains => ../mains

require (
	github.com/haraldrudell/parl v0.4.45
	github.com/haraldrudell/parl/mains v0.4.45
	gopkg.in/yaml.v3 v3.0.1
)

require (
	golang.org/x/exp v0.0.0-20230124195608-d38c7dcee874 // indirect
	golang.org/x/sys v0.4.0 // indirect
	golang.org/x/text v0.6.0 // indirect
)
