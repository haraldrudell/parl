/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package yamlo

import (
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pflags"
)

// AddYamlReferences returns a [pflags.OptionData] slice of options to main
//   - return value is appended to the [pflags.OptionData] list in main
//   - filedPointerValues is a list of references to main’s y YamlData
//   - subPackageOptionData is a list of options declared in subpackage
func AddYamlReferences(
	subPackageOptionData []pflags.OptionData,
	fieldPointerValues []interface{},
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
		subPackageOptionData[i].Y = yReference
	}

	optionDataList = subPackageOptionData

	return
}
