/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"fmt"

	"github.com/haraldrudell/parl/pflags"
	"github.com/haraldrudell/parl/pos"
)

const (
	// as second argument to [BaseOptionData], indicates that yaml options -yamlFile -yamlKey should not be present
	YamlNo YamlOption = false
	// as second argument to [BaseOptionData], indicates that yaml options -yamlFile -yamlKey should be present
	YamlYes YamlOption = true
	// indicates silent: no banner. Must be first option on command-line ‘-silent’
	SilentString = "-" + silentOption
	silentOption = "silent"
)

// type for second argument to [BaseOptionData]
//   - mains.YamlNo mains.YamlYes
type YamlOption bool

// BaseOptionsType is the type that holds mains’ effective option values
type BaseOptionsType = struct {
	YamlFile, YamlKey, Verbosity string
	Debug, Silent, Version       bool
}

// BaseOptions is the value that holds mains’ effective option values
var BaseOptions BaseOptionsType

// BaseOptionData returns basic options for mains
//   - verbose debug silent version
//   - if yaml == YamlYes: yamlFile yamlKey
func BaseOptionData(program string, yaml YamlOption) (optionData []pflags.OptionData) {

	var nonYamlOptions = []pflags.OptionData{
		{P: &BaseOptions.Version, Name: "version", Value: false, Usage: "displays version"},
		{P: &BaseOptions.Verbosity, Name: "verbose", Value: "", Usage: "Regular expression for selective debug, eg. main.main: https://github.com/google/re2/wiki/Syntax"},
		{P: &BaseOptions.Debug, Name: pflags.DebugOptionName, Value: false, Usage: "Global debug printing with code locations and long stack traces"},
		{P: &BaseOptions.Silent, Name: silentOption, Value: false, Usage: "Suppresses banner and informational output. Must be first option"},
	}
	optionData = nonYamlOptions

	if yaml == YamlYes {
		var yamlOptions = []pflags.OptionData{
			{P: &BaseOptions.YamlFile, Name: "yamlFile", Value: "", Usage: fmt.Sprintf("Use specific file other than %s.yaml %[1]s-%s.yaml in ~/apps .. /etc", program, pos.ShortHostname())},
			{P: &BaseOptions.YamlKey, Name: "yamlKey", Value: "", Usage: "Other dictionary key than ‘options:’"},
		}
		optionData = append(optionData, yamlOptions...)
	}

	return
}
