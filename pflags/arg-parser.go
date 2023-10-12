/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pflags

import (
	"flag"
)

type ArgParser struct {
	optionsList []OptionData
	usage       func()
}

func NewArgParser(optionsList []OptionData, usage func()) (argParser *ArgParser) {
	return &ArgParser{
		optionsList: optionsList,
		usage:       usage,
	}
}

// Parse invokes [flag.Parse] after providing optionsList and usage to flag package
//   - -no-flagname flags are inverted before and after
func (a *ArgParser) Parse() {
	var booleanOffList []*bool

	flag.Usage = a.usage

	// provide optionData list to flag package
	omLen := len(a.optionsList)
	for i := 0; i < omLen; i++ {
		option := &a.optionsList[i]

		// the dumb flag package does not support -no-flagname off-flags.
		//	- the only allowed option to not have an argument is a boolean flag
		//	- so an off-flag, that has default value true, must be a boolean flag with
		//		default value changed to false beforehand
		//	- afterwards, invert its value
		if boolp, ok := option.P.(*bool); ok {
			// this is a boolean flag, is its default value true?
			if value, ok := option.Value.(bool); ok {
				if value {

					// this is a flag of type bool with default value true
					//	- this is used for off flags -no-flagname
					//	- remember the value pointer so it an be inverted later
					booleanOffList = append(booleanOffList, boolp)
					var o = *option
					o.Value = false // have default value false
					option = &o
				}
			}
		}
		option.AddOption()
	}

	flag.Parse()

	// iterate over the -no-flagname
	//	- if the value is false, the flag was not provided, result should be true
	//	- if the value is true, the flag was provided, the result should be false
	for _, boolp := range booleanOffList {
		*boolp = !*boolp
	}
}
