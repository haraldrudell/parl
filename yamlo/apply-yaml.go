/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package yamlo

import (
	"strings"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pflags"
	"github.com/haraldrudell/parl/yamler"
)

const (
	// key name from options or a default ‘options’
	defaultTopKey = "options"
)

// ApplyYaml updates effective options with values read from a yaml file
//   - program is app name “date” used to construct yaml file name
//   - yamlFile is a specified file, or for empty string, a scan for files:
//   - — filename: [program]-[hostname].yaml [program].yaml
//   - — Directories: ~/apps .. /etc
//   - — if a specified file is missing, that is error
//   - — if no default file exists, or file was empty,
//     no yaml options are loaded
//   - yamlDictionaryKey is the key read form the top-level dictionary
//     in yaml, empty string is default “options:”
//   - genericYaml is a wrapper of unknown types
//   - optionData is the list of options, containing pointers to effective
//     option values
//   - The top entry in the yaml file must be a dictionary
//   - The value for yamlDictionaryKey must be a dictionary
//   - the remainder of the yamlDictionaryKey is read when it matches
//     the YamData struct
//   - -verbose=yamlo.ApplyYaml “github.com/haraldrudell/parl/yamlo.ApplyYaml”
func ApplyYaml(
	program, yamlFile, yamlDictionaryKey string,
	genericYaml yamler.GenericYaml,
	optionData []pflags.OptionData,
) (err error) {
	if genericYaml == nil {
		panic(perrors.NewPF("genericYaml cannot be nil"))
	}

	// read text from the yaml file
	var yamlText []byte
	parl.Debug("Arguments: yamlFile: %q yamlKey: %q", yamlFile, yamlDictionaryKey)
	if yamlDictionaryKey == "" {
		yamlDictionaryKey = defaultTopKey
	}
	// the filename actually read
	var filename string
	if filename, yamlText, err = FindFile(yamlFile, program); err != nil {
		return // yaml read failure return
	} else if filename == "" || len(yamlText) == 0 {
		parl.Debug("no yaml file")
		return // no default file existed, or file was empty: noop
	}
	parl.Debug("filename: %q top-level key: %q bytes: %q", filename, yamlDictionaryKey, string(yamlText))

	// unmarshal yaml into genericYaml value pointer
	//	- ie. main.y effective yaml values
	var hasData bool
	if hasData, err = genericYaml.Unmarshal(yamlText, yamlDictionaryKey); perrors.IsPF(&err, "filename: %q: %w", filename, err) {
		return // error during unmarshaling return
	} else if !hasData {
		parl.Debug("has no data")
		return // yaml contained no data, eg. no “options:” key
	}

	// get a map of the field references that is appearing in yaml
	var yamlVisistedReferences map[any]string
	if yamlVisistedReferences, err = genericYaml.VisitedReferencesMap(yamlText, yamlDictionaryKey); err != nil {
		return // unmarshal failure return
	}
	yamlText = nil
	if parl.IsThisDebug() {
		parl.Log("yamlVisitedKeys: %v\n", yamlVisistedReferences)
		var oMap = make(map[string]any)
		for _, o := range optionData {
			var yamlFieldReference = o.Y
			if yamlFieldReference == nil {
				continue
			}
			oMap[o.Name] = o.Y
		}
		parl.Log("option-references: %v\n", oMap)
		parl.Log("yaml-y: %s", genericYaml.YDump())
		var effectiveValueMap, _ = pflags.OptionValues(optionData)
		parl.Log("option-values: %v\n", effectiveValueMap)
	}

	// get map of visited options
	var visitedOptions = pflags.NewVisitedOptions().Map()
	if parl.IsThisDebug() {
		var opts []string
		for k := range visitedOptions {
			opts = append(opts, k)
		}
		parl.Log("visitedOptions: %s", strings.Join(opts, "\x20"))
	}

	// iterate over options and apply yaml changes
	//	- ignore if no yaml key
	//	- ignore if yamlVisitedKeys exists and do not have the option
	for _, optionData := range optionData {
		if visitedOptions[optionData.Name] || // was specified on command line, overrides yaml
			optionData.Y == nil || // does not have yaml value
			yamlVisistedReferences[optionData.Y] == "" { // was not visted by yaml
			continue
		}
		if err = optionData.ApplyYaml(); err != nil {
			return
		}
	}

	if parl.IsThisDebug() {
		var effectiveValueMap, _ = pflags.OptionValues(optionData)
		parl.Log("resulting option-values: %v\n", effectiveValueMap)
	}

	return
}
