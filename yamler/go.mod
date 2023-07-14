module github.com/haraldrudell/parl/yamler

go 1.19

replace github.com/haraldrudell/parl => ../../parl

replace github.com/haraldrudell/parl/mains => ../mains

require (
	github.com/haraldrudell/parl v0.4.104
	github.com/haraldrudell/parl/mains v0.4.104
	gopkg.in/yaml.v3 v3.0.1
)

require (
	golang.org/x/exp v0.0.0-20230713183714-613f0c0eb8a1 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/text v0.11.0 // indirect
)
