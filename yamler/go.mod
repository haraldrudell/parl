module github.com/haraldrudell/parl/yamler

go 1.21

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/haraldrudell/parl v0.4.132
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa
	gopkg.in/yaml.v3 v3.0.1
)

require golang.org/x/text v0.14.0 // indirect
