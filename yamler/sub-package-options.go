/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package yamler

import (
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pflags"
)

// SubPackageOptions returns a [pflags.OptionData] slice of options to main
//   - return value is appended to the [pflags.OptionData] list in main
//   - newYamlValue returns a pflags.YamlValue referencing main’s y YamlData
//   - filedPointerValues is a list of references to main’s y YamlData
func SubPackageOptions(
	newYamlValue func(fieldPointerValue interface{}) (yamlValue *pflags.YamlValue),
	fieldPointerValues []interface{},
	subPackageOptionData []pflags.OptionData,
) (optionDataList []pflags.OptionData) {

	// verify main and subPackage integrity
	if len(fieldPointerValues) != len(subPackageOptionData) {
		panic(perrors.ErrorfPF("length mismatch: main.fieldPointerValues: %d subPackageOptionData: %d",
			len(fieldPointerValues), len(subPackageOptionData),
		))
	}

	// link subPackageOptionData to main’s yaml
	for i := 0; i < len(subPackageOptionData); i++ {

		// get the reference to the field of main’s y YamlData object for this subPackageOptionData option
		var yReference = fieldPointerValues[i]
		if yReference == nil {
			continue // this option is not in yaml
		}

		// update subPackageOptionData so that [ApplyYaml] can update
		// this option value with values read from yaml
		subPackageOptionData[i].Y = newYamlValue(yReference)
	}

	return
}
