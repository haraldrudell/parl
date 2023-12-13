/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package pflags provides declarative options and a string-slice option type.
package pflags

import (
	"flag"
	"os"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pstrings"
)

const (
	// the os.Args value of debug option: “-debug”
	DebugOption = "-" + DebugOptionName
	// the name of debug option “debug”
	DebugOptionName = "debug"
)

const (
	offFlagFalse = false
	offFlagTrue  = true
)

// ArgParser invokes [flag.Parse] with off-flags support.
// pflags package compared to flags package:
//   - ability to use declarative options
//   - ability to read default option values from yaml configuration files
//   - support for multiple-strings option value
//   - support for unary off-flags, “-no-flag” options
//   - map or list of visited options
type ArgParser struct {
	optionsList []OptionData
	usage       func()
}

// NewArgParser returns an options-parser with off-flags support
func NewArgParser(optionsList []OptionData, usage func()) (argParser *ArgParser) {
	return &ArgParser{
		optionsList: optionsList,
		usage:       usage,
	}
}

// Parse invokes [flag.Parse] after providing optionsList and usage to flag package
//   - -no-flagname flags are inverted before and after
func (a *ArgParser) Parse() {

	// options have not been parsed yet, so verbose state cannot be determined
	//	- if first option is “-debug”, it’s debug
	if len(os.Args) > 1 && os.Args[1] == DebugOption {
		var _, defaultsMap = OptionValues(a.optionsList)
		parl.Log("option defaults: %v", defaultsMap)
		parl.Log("os.args[1:]: %s", pstrings.QuoteList(os.Args[1:]))
		defer func() {
			var effectiveValueMap, _ = OptionValues(a.optionsList)
			parl.Log("resulting option values: %v", effectiveValueMap)
		}()
	}

	flag.Usage = a.usage

	// booleanOffList is a list of pointers to the effective values of off-flags
	//	- on return from [flag.Parse], these values are inverted
	var booleanOffList []*bool
	defer a.parseEnd(&booleanOffList)

	// provide optionData list to flag package
	omLen := len(a.optionsList)
	for i := 0; i < omLen; i++ {
		option := &a.optionsList[i]

		// flag package does not support -no-flagname off-flags. pflags does implement off-flags
		//	- an off-flag is a boolean flag with default value true
		//	- — any time the flag occurs on the command-line its value is set to false
		//	- — off-flags typically have names with leading “no”
		//	- — [OptionData.Name] is like “no-stdin”
		//	- — on command-line is provided like “-no-stdin”
		//	- the only option-type allowed by [flags.Parse] to not have an argument is a boolean flag
		//	- — therefore, off-flags must be plain boolean flags
		//	- but boolean flags have default value false, and are set to true on occurrence
		//	- — therefore, identify all boolean flags with default value true
		//	- — prior to invoking [flags.Parse], set their default value to false
		//	- — on occurrence, [flags.Parse] will set their value to true
		//	- — on return from [flags.Parse], invert the off-flags’ effective values
		//	- — if an off-flag did not occur, [flags.Parse] set its effective value to false, and result is true
		//	- — if an off-flag did occur, [flags.Parse] set its effective value to true, and result is false

		// is this option’s effective value type bool?
		if boolp, ok := option.P.(*bool); ok {
			// is the default value for this boolean flag true?
			if value, ok := option.Value.(bool); ok && value {

				// this is a flag of type bool with default value true
				//	- this is used for off-flags -no-flagname
				//	- retain a pointer to the off-flag option’s effective value
				booleanOffList = append(booleanOffList, boolp)

				// in a copy of optionData, set off-option default value to false
				var o = *option
				o.Value = offFlagFalse // have default value false
				option = &o            // use the copy
			}
		}
		option.AddOption()
	}

	// flag.Parse uses os.Args[1:]
	flag.Parse()
}

// iterate over the -no-flagname off-flag options
//   - if the value is false, the flag was not provided, result should be true
//   - if the value is true, the flag was provided, the result should be false
func (a *ArgParser) parseEnd(booleanOffList *[]*bool) {
	for _, boolp := range *booleanOffList {
		// invert effective value
		if *boolp {
			*boolp = offFlagFalse
		} else {
			*boolp = offFlagTrue
		}
	}
}
