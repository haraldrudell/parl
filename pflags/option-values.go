/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pflags

import "fmt"

// OptionValues returns a printable map of current values
//   - %v: map[debug:false …]
func OptionValues(optionData []OptionData) (effectiveValueMap, defaultsMap map[string]string) {
	effectiveValueMap = make(map[string]string, len(optionData))
	defaultsMap = make(map[string]string, len(optionData))
	for _, o := range optionData {
		effectiveValueMap[o.Name] = o.ValueDump()
		defaultsMap[o.Name] = fmt.Sprintf("%v", o.Value)
	}
	return
}
