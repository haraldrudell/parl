/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"fmt"

	"github.com/haraldrudell/parl/mains/malib"
	"github.com/haraldrudell/parl/pflags"
	"github.com/haraldrudell/parl/pos"
)

const (
	// indicates silent: no banner. Must be first option on command-line ‘-silent’
	SilentString = "-" + silentOption
	// name of version option
	Version = "version"
)

// BaseOptions is the value that holds mains’ effective option values
//   - -yamlFile -yamlKey -verbose -debug -silent -version -no-yaml
var BaseOptions BaseOptionsType

// BaseOptionData returns basic options for mains
//   - program: used to generate help text
//   - yaml: [YamlNo] means no yaml options
//   - -verbose -debug -silent -version
//   - if yaml == YamlYes: -yamlFile -yamlKey
func BaseOptionData(program string, yaml ...malib.YamlOption) (optionData []pflags.OptionData) {

	var nonYamlOptions = []pflags.OptionData{
		{P: &BaseOptions.Version, Name: Version, Value: false, Usage: "displays version"},
		{P: &BaseOptions.Verbosity, Name: "verbose", Value: "", Usage: verboseOptionHelp},
		{P: &BaseOptions.Debug, Name: pflags.DebugOptionName, Value: false, Usage: "Global debug printing with code locations and long stack traces"},
		{P: &BaseOptions.Silent, Name: silentOption, Value: false, Usage: "Suppresses banner and informational output. Must be first option"},
	}
	optionData = nonYamlOptions

	if len(yaml) == 0 || yaml[0] != malib.YamlNo {
		var yamlOptions = []pflags.OptionData{
			{P: &BaseOptions.YamlFile, Name: "yamlFile", Value: "", Usage: fmt.Sprintf("Use specific file other than %s.yaml %[1]s-%s.yaml in ~/apps .. /etc", program, pos.ShortHostname())},
			{P: &BaseOptions.YamlKey, Name: "yamlKey", Value: "", Usage: "Other dictionary key than ‘options:’"},
			{P: &BaseOptions.DoYaml, Name: "no-yaml", Value: true, Usage: "do not read yaml data"},
		}
		optionData = append(optionData, yamlOptions...)
	}

	return
}

const (
	// name of silent option
	silentOption = "silent"
	// help text for -verbose
	verboseOptionHelp = "Regular expression for selective debug matched against CodeLocation FuncName" +
		"\nmain.main: -verbose=main.main" +
		"\ngithub.com/haraldrudell/parl/mains.(*Executable).Init: -verbose=mains...Executable" +
		"\ngithub.com/haraldrudell/parl/mains.Func: -verbose=mains.Func" +
		"\nper https://github.com/google/re2/wiki/Syntax"
)
