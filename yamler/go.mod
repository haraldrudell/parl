module github.com/haraldrudell/parl/yamler

go 1.21

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/haraldrudell/parl v0.4.134
	golang.org/x/exp v0.0.0-20231127185646-65229373498e
	gopkg.in/yaml.v3 v3.0.1
)

require golang.org/x/text v0.14.0 // indirect
