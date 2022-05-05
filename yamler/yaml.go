/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package yamler

import (
	"flag"
	"strings"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/mains"
	"github.com/haraldrudell/parl/pstrings"
	"gopkg.in/yaml.v3"
)

type UnmarshalFunc func(in []byte, out interface{}) (err error) // yaml.Unmarshal
type UnmarshalThunk func(bytes []byte, unmarshal UnmarshalFunc, yamlKey string) (hasDate bool, err error)

var Unmarshal = yaml.Unmarshal

/*
ApplyYaml adds options read from a yaml file.

File: The filename searched comes from Executable.Program or the yamlFile argument.
If yamlFile is set exactly this file is read. If yamlFile is not set, A search is executed in the
directories ~/apps .. and /etc.
The filename searched is [program]-[hostname].yaml and [program].yaml.

Content: If the file does not eixst, no action is taken. If the file is empty,
no action is taken.

The top entry in the yaml file is expected to be a dictionary of dictionaries.
The key searched for in the
top level dictionary is the yamlKey argument or “options” if not set.

thunk needs to in otuside of the library for typoing reasons and is
implemented similar to the below. y is the variable receiving parsed yaml data. The om list
is then svcaned for its Y pointers to copy yaml settings to options.
 func applyYaml(bytes []byte, unmarshal mains.UnmarshalFunc, yamlDictionaryKey string) (hasData bool, err error) {
   yamlContentObject := map[string]*YamlData{} // need map[string] because yaml top level is dictionary
   if err = unmarshal(bytes, &yamlContentObject); err != nil {
     return
   }
   yamlDataPointer := yamlContentObject[yamlDictionaryKey] // pick out the options dictionary value
   if hasData = yamlDataPointer != nil; !hasData {
     return
   }
   y = *yamlDataPointer
   return
 }
*/
func ApplyYaml(ex mains.Executable, yamlFile, yamlKey string, thunk UnmarshalThunk, om []mains.OptionData) {
	if thunk == nil {
		panic(parl.New("yaml.ApplyYaml: thunk cannot be nil"))
	}
	parl.Debug("Arguments: yamlFile: %q yamlKey: %q", yamlFile, yamlKey)
	filename, byts := FindFile(yamlFile, ex.Program)
	if filename == "" || len(byts) == 0 {
		parl.Debug("ex.ApplyYaml: no yaml file")
		return
	}
	yamlDictionaryKey := GetTopLevelKey(yamlKey) // key name from option or a default
	parl.Debug("filename: %q top-level key: %q bytes: %q", filename, yamlDictionaryKey, string(byts))

	// try to obtain the list of defined keys in the options dictionary
	var yamlVisitedKeys map[string]bool
	yco := map[string]map[string]interface{}{} // a dictionary of dictionaries with unknown content
	parl.Debug("ex.ApplyYaml: first yaml.Unmarshal")
	if yaml.Unmarshal(byts, &yco) == nil {
		yamlVisitedKeys = map[string]bool{}
		if optionsMap := yco[yamlDictionaryKey]; optionsMap != nil {
			for key := range optionsMap {
				yamlVisitedKeys[strings.ToLower(key)] = true
			}
		}
	}
	parl.Debug("ex.ApplyYaml: yamlVisitedKeys: %v\n", yamlVisitedKeys)

	hasData, err := thunk(byts, yaml.Unmarshal, yamlDictionaryKey)
	if err != nil {
		ex.AddErr(parl.Errorf("ex.ApplyYaml thunk: filename: %q: %w", filename, err)).Exit()
	} else if !hasData {
		return
	}

	// iterate over options
	// ignore if no yaml key
	// ignore if yamlVisitedKeys exists and do not have the option
	visitedOptions := map[string]bool{}
	flag.Visit(func(fp *flag.Flag) { visitedOptions[fp.Name] = true })
	for _, optionData := range om {
		if visitedOptions[optionData.Name] || // was specified on command line
			optionData.Y == nil { // does not have yaml value
			continue
		}
		if yamlVisitedKeys != nil { // we have a map of the yaml keys present
			if !yamlVisitedKeys[optionData.Y.Name] {
				continue // this key was not present in yaml
			}
		} else if pstrings.IsDefaultValue(optionData.Y.Pointer) {
			continue // no visited information,, so ignore default values
		}
		if err := optionData.ApplyYaml(); err != nil {
			ex.AddErr(err)
			ex.Exit()
		}
	}
}
