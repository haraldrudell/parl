/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import "github.com/haraldrudell/parl/pflags"

// OptionsList merges a lists of options to a single list
func OptionsList(optionsLists ...[]pflags.OptionData) (singleList []pflags.OptionData) {

	// get size to allocate and allocate
	var size int
	for _, optionsList := range optionsLists {
		size += len(optionsList)
	}
	singleList = make([]pflags.OptionData, size)

	// merge lists
	var tempList = singleList
	for _, optionsList := range optionsLists {
		copy(tempList, optionsList)
		tempList = tempList[len(optionsList):]
	}

	return
}
