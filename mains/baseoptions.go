/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"fmt"

	"github.com/haraldrudell/parl/pos"
)

const (
	YamlNo      YamlOption = false
	YamlYes     YamlOption = true
	DebugOption            = "-debug"
)

// helpOptions are additional options implemented by the flag package
//   - [mains.ArgParser] invokes [flag.Parse]
//   - [flag.FlagSet] parseOne method have these as string constants in the code
var helpOptions = []string{"h", "help"}

// mains.YamlNo mains.YamlYes
type YamlOption bool

type BaseOptionsType = struct {
	YamlFile  string
	YamlKey   string
	Verbosity string
	Debug     bool
	Silent    bool
}

var BaseOptions BaseOptionsType

// BaseOptionData returns basic options for mains
//   - verbose debug silent
//   - if yaml == YamlYes: yamlFile yamlKey
func BaseOptionData(program string, yaml YamlOption) (od []OptionData) {
	od = []OptionData{
		{P: &BaseOptions.Verbosity, Name: "verbose", Value: "", Usage: "Regular expression for selective debug, eg. main.main: https://github.com/google/re2/wiki/Syntax"},
		{P: &BaseOptions.Debug, Name: DebugOption[1:], Value: false, Usage: "Global debug printing with code locations and long stack traces"},
		{P: &BaseOptions.Silent, Name: silentOption, Value: false, Usage: "Suppresses banner and informational output. Must be first option"},
	}
	if yaml == YamlYes {
		od = append(od, []OptionData{
			{P: &BaseOptions.YamlFile, Name: "yamlFile", Value: "", Usage: fmt.Sprintf("Use specific file other than %s.yaml %[1]s-%s.yaml in ~/apps .. /etc", program, pos.ShortHostname())},
			{P: &BaseOptions.YamlKey, Name: "yamlKey", Value: "", Usage: "Other dictionary key than 'options:'"},
		}...)
	}
	return
}

// returns implicit help options: "h" "help"
func HelpOptions() (optionNames []string) {
	return helpOptions
}
