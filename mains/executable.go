/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package mains contains functions for implementing a service or command-line utility
package mains

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/errorglue"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pflags"
	"github.com/haraldrudell/parl/plogger"
	"github.com/haraldrudell/parl/pos"
	"github.com/haraldrudell/parl/pruntime"
	"github.com/haraldrudell/parl/pstrings"
)

const (
	rfcTimeFormat = "2006-01-02 15:04:05-07:00"
	usageHeader   = "Usage:"
	optionsSyntax = "[options…]"
	helpHelp      = "\x20\x20-help -h --help\n\x20\x20\tShows this help"
	timeHeader    = "time: %s"
	hostHeader    = "host: %s"
	defaultOK     = "completed successfully"
	NoOK          = "-"
)

const (
	// NoArguments besides switches, zero trailing arguments is allowed
	NoArguments = 1 << iota
	// OneArgument besides switches, exactly one trailing arguments is allowed
	OneArgument
	// ManyArguments besides switches, one or more trailing arguments is allowed
	ManyArguments
)

// ArgumentSpec bitfield for 0, 1, many arguments following command-line switches
type ArgumentSpec uint32

// Executable constant strings that describes an executable
// advisable static values include Program Version Comment Description Copyright License Arguments
// like:
//
//	var ex = mains.Executable{
//	  Program:     "getip",
//	  Version:     "0.0.1",
//	  Comment:     "first version",
//	  Description: "finds ip address for hostname",
//	  Copyright:   "© 2020-present Harald Rudell <harald.rudell@gmail.com> (http://www.haraldrudell.com)",
//	  License:     "All rights reserved",
//	  Arguments:   mains.NoArguments | mains.OneArgument,
//	}
type Executable struct {

	// fields typically statically assigned in main

	Program        string       // “gonet”
	Version        string       // “0.0.1”
	Comment        string       // [ banner text after program and version] “options parsing” changes in last version
	Description    string       // [Description part of usage] “configures firewall and routing”
	Copyright      string       // “© 2020…”
	License        string       // “ISC License”
	OKtext         string       // “Completed successfully”
	ArgumentsUsage string       // usage help text for arguments after options
	Arguments      ArgumentSpec // eg. mains.NoArguments

	// fields below popualted by .Init()

	Launch       time.Time // process start time
	LaunchString string    // Launch as printable rfc 3339 time string
	Host         string    // short hostname, ie. no dots “mymac”

	// additional fields

	err      error    // errors addded by AddErr and other methods
	ArgCount int      // number of post-options strings during parse
	Arg      string   // if one post-options string and that is allowed, this is the string
	Args     []string // any post-options strings if allowed
	// errors are printed with stack traces, associated values and errors
	//	- panics are always printed long
	//	- if errors long or more than 1 error, the first error is repeated last as a one-liner
	IsLongErrors bool
	// adds a code location to errors if not IsLongErrors
	IsErrorLocation bool
	// optionsWereParsed signals that parsing completed without panic
	optionsWereParsed atomic.Bool
}

// Init initializes a created [mains.Executable] value
//   - the value should have relevant fields populates such as exeuctable name and more
//   - — Program Version Comment Copyright License Arguments
//   - populates launch time and sets silence if first os.Args argument is “-silent.”
//   - Init supports function chaining like:
//
// typical code:
//
//	ex.Init().
//	  PrintBannerAndParseOptions(optionData).
//	  LongErrors(options.Debug, options.Verbosity != "").
//	  ConfigureLog()
//	applyYaml(options.YamlFile, options.YamlKey, applyYaml, optionData)
//	…
func (x *Executable) Init() (ex2 *Executable) {
	ex2 = x
	var now = ProcessStartTime()
	x.Launch = now
	x.LaunchString = now.Format(rfcTimeFormat)
	x.Host = pos.ShortHostname()
	if len(os.Args) > 1 && os.Args[1] == SilentString {
		parl.SetSilent(true)
	}
	return
}

// LongErrors sets if errors are printed with stack trace and values. LongErrors
// supports functional chaining:
//
//	exe.Init().
//	  …
//	  LongErrors(options.Debug, options.Verbosity != "").
//	  ConfigureLog()…
//
// isLongErrors prints full stack traces, related errors and error data in string
// lists and string maps.
//
// isErrorLocation appends the innermost location to the error message when isLongErrors
// is not set:
//
//	error-message at error116.(*csTypeName).FuncName-chainstring_test.go:26
func (x *Executable) LongErrors(isLongErrors bool, isErrorLocation bool) *Executable {
	parl.Debug("exe.LongErrors long: %t location: %t", isLongErrors, isErrorLocation)
	x.IsLongErrors = isLongErrors
	x.IsErrorLocation = isErrorLocation
	return x
}

// PrintBannerAndParseOptions prints greeting like:
//
//	parl 0.1.0 parlca https server/client udp server
//
// It then parses options described by []OptionData stroing the values at OptionData.P.
// If options fail to parse, a proper message is printed to stderr and the process exits
// with status code 2. PrintBannerAndParseOptions supports functional chaining like:
//
//	exe.Init().
//	  PrintBannerAndParseOptions(…).
//	  LongErrors(…
//
// Options and yaml is configured likeso:
//
//	var options = &struct {
//	  noStdin bool
//	  *mains.BaseOptionsType
//	}{BaseOptionsType: &mains.BaseOptions}
//	var optionData = append(mains.BaseOptionData(exe.Program, mains.YamlYes), []mains.OptionData{
//	  {P: &options.noStdin, Name: "no-stdin", Value: false, Usage: "Service: do not use standard input", Y: mains.NewYamlValue(&y, &y.NoStdin)},
//	}...)
//	type YamlData struct {
//	  NoStdin bool // nostdin: true
//	}
//	var y YamlData
func (x *Executable) PrintBannerAndParseOptions(optionsList []pflags.OptionData) (ex1 *Executable) {
	ex1 = x

	// print program name and populated details
	var banner = pstrings.FilteredJoin([]string{
		pstrings.FilteredJoinWithHeading([]string{
			"", x.Program,
			"version", x.Version,
			"comment", x.Comment,
		}, "\x20"),
		x.Copyright,
		fmt.Sprintf(timeHeader, x.LaunchString),
		fmt.Sprintf(hostHeader, x.Host),
	}, "\n")
	if len(banner) != 0 {
		parl.Info(banner)
	}

	pflags.NewArgParser(optionsList, x.usage).Parse()
	if BaseOptions.Version {
		os.Exit(0)
	}

	// parse arguments
	args := flag.Args() // command-line arguments not part of flags
	count := len(args)
	argsOk :=
		count == 0 && (x.Arguments&NoArguments != 0) ||
			count == 1 && (x.Arguments&OneArgument != 0) ||
			count > 0 && (x.Arguments&ManyArguments != 0)
	if !argsOk {
		if count == 0 {
			if x.Arguments&ManyArguments != 0 {
				parl.Log("There must be one or more arguments")
			} else {
				parl.Log("There must be one argument")
			}
		} else {
			for i, v := range args {
				args[i] = fmt.Sprintf("%q", v)
			}
			parl.Log("Unknown parameters: %s\n", strings.Join(args, "\x20"))
		}
		x.usage()
		pos.Exit(pos.StatusCodeUsage, nil)
	}
	x.ArgCount = count
	if count == 1 && (x.Arguments&OneArgument != 0) {
		x.Arg = args[0]
	}
	if count > 0 && (x.Arguments&ManyArguments != 0) {
		x.Args = args
	}

	x.optionsWereParsed.Store(true)

	return
}

// ConfigureLog configures the default log such as parl.Log parl.Out parl.D
// for silent, debug and regExp.
// Settings come from BaseOptions.Silent and BaseOptions.Debug.
//
// ConfigureLog supports functional chaining like:
//
//	exe.Init().
//	  …
//	  ConfigureLog().
//	  ApplyYaml(…)
func (x *Executable) ConfigureLog() (ex1 *Executable) {
	if BaseOptions.Silent {
		parl.SetSilent(true)
	}
	if BaseOptions.Debug {
		parl.SetDebug(true)
	}
	if BaseOptions.Verbosity != "" {
		if err := parl.SetRegexp(BaseOptions.Verbosity); err != nil {
			pos.Exit(pos.StatusCodeUsage, err)
		}
	}
	parl.Debug("exe.ConfigureLog silent: %t debug: %t verbosity: %q\n",
		BaseOptions.Silent, BaseOptions.Debug, BaseOptions.Verbosity)
	return x
}

// Recover function to be used in main.main:
//
//	func main() {
//	  defer Recover()
//	  …
//
// On panic, the function prints to stderr: "Unhandled panic invoked exe.Recover: stack:"
// followed by a stack trace. It then adds an error to mains.Executable and terminates
// the process with status code 1
func (x *Executable) Recover(errp ...*error) {

	// get error from *errp and store in ex.err
	if len(errp) > 0 {
		if errp0 := errp[0]; errp0 != nil {
			if err := *errp0; err != nil {
				x.AddErr(err)
			}
		}
	}

	// ensure -debug honored if panic before options parsing
	if !x.optionsWereParsed.Load() {
		for _, option := range os.Args {
			if option == DebugOption { // -debug
				parl.SetDebug(true)
			}
		}
	}

	// check for panic
	if v := recover(); v != nil {

		// determine if v is error
		err, recoverValueIsError := v.(error)
		var error0 error

		// debug print
		isDebug := parl.IsThisDebug()
		if isDebug {
			hasStack := false
			var valueString string
			var error0type string
			if !recoverValueIsError {
				valueString = parl.Sprintf(" '%+v'", v)
			} else {
				error0 = perrors.Error0(err)
				error0type = parl.Sprintf(" panic error type: %T", error0)
				hasStack = perrors.HasStack(err)
				error0value := fmt.Sprintf("error0: %+v", error0)
				if hasStack {
					valueString = parl.Sprintf(" error-value:\n\n%s\n\n%s\n\n", perrors.Long(err), error0value)
				} else {
					valueString = err.Error() + "\n" + error0value
				}
			}
			parl.Debug("%s: panic with -debug: recover-value type: %T%s hasStack: %t%s",
				pruntime.NewCodeLocation(0).PackFunc(),
				v, error0type, hasStack, valueString)
		}

		// print panic message and invocation stack
		var stackString string
		if isDebug {
			stackString = " recovery stack trace:\n\n" + pruntime.DebugStack(0) + "\n\n"
		}
		var programString string
		if x.Program != "" {
			programString = "\x20" + x.Program
		}
		parl.Log("\n\nProgram%s Recovered a Main-Thread panic:%s", programString, stackString)

		// store recovery value as error
		var prepend string
		var postpend string
		if !recoverValueIsError {
			err = perrors.Errorf("panic: non-error value: %T %[1]v", v)
		} else {
			prepend = "panic: \x27"
			postpend = "\x27"
			if isDebug {
				// put error0 type name in error message
				postpend += parl.Sprintf(" type: %T", error0)
			}
		}
		err = perrors.Errorf("main-thread %s%w%s", prepend, err, postpend)
		x.AddErr(err)
	}

	// ex.err now contains program result

	// print completed successfully
	if x.err == nil && x.OKtext != NoOK {
		var program string
		var completedSuccessfully string
		now := "at " + parl.ShortSpace() // time now second precision
		if x.OKtext != "" {
			completedSuccessfully = x.OKtext // custom "Completed successfully"
		} else {
			program = x.Program
			completedSuccessfully = defaultOK // "<executable> completed successfully
		}
		sList := []string{program, completedSuccessfully, now}
		parl.Log(pstrings.FilteredJoin(sList)) // to stderr
	}

	// will print any errors
	x.Exit()
}

// AddErr extended with immediate printing of first error
func (x *Executable) AddErr(err error) {

	// debug printing
	if parl.IsThisDebug() {
		packFunc := perrors.PackFunc()
		var errS string
		if err != nil {
			errS = "\x27" + err.Error() + "\x27"
		} else {
			errS = "nil"
		}
		parl.Debug("\n%s(error: %s)\n%[1]s invocation:\n%[3]s", packFunc, errS, pruntime.Invocation(0))
		plogger.GetLog(os.Stderr).Output(0, "") // newline after debug location. No location appended to this printout
	}

	// if AddErr with no error, do nothing
	if err == nil {
		return // no error do nothing return
	}

	// if the first error, immediately print it
	if x.err == nil {

		// print and store the first error
		x.printErr(err, checkForPanic(err))
		x.err = err
		return // first error stored return
	}

	// append subsequent error
	x.err = perrors.AppendError(x.err, err)
}

// Exit terminate from mains.err: exit 0 or echo to stderr and status code 1
//   - Usually invoked for all app terminations
//   - — either by defer ex.Recover(&err) at beginning fo main
//   - — or rarely by direct invocation in program code
//   - when invoked, errors are expected to be in ex.err from:
//   - — ex.AddErr or
//   - — ex.Recover
//   - Exit does not return from invoking os.Exit
func (x *Executable) Exit(stausCode ...int) {

	// get requested status code
	var statusCode0 int
	if len(stausCode) > 0 {
		statusCode0 = stausCode[0]
	}

	// printouts when IsDebug
	if x.err == nil {
		parl.Debug("\nexe.Exit: no error")
	} else {
		parl.Debug("\nexe.Exit: err: %T '%[1]v'", x.err)
	}
	parl.Debug("\nexe.Exit invocation:\n%s\n", string(pruntime.StackTrace()))
	if parl.IsThisDebug() { // add newline during debug without location
		plogger.GetLog(os.Stderr).Output(0, "") // newline after debug location. No location appended to this printout
	}

	// terminate when there are no errors
	if x.err == nil {
		if statusCode0 != 0 {
			pos.OsExit(statusCode0)
		}
		pos.Exit0()
	}

	// print all errors except the very first
	// the very first error was already printed when it occurred
	errorList := perrors.ErrorList(x.err)
	errorCount := 0
	for _, errorListEntry := range errorList {
		for _, err := range errorglue.ErrorList(errorListEntry) {
			errorCount++
			if errorCount == 1 {
				continue
			}
			x.printErr(err, checkForPanic(err))
		}
	}

	// just before exit, print the one-liner message of the first occurring error again
	if x.IsLongErrors || errorCount > 1 {
		fmt.Fprintln(os.Stderr, x.err)
	}

	// exit 1
	if statusCode0 == 0 {
		statusCode0 = pos.StatusCodeErr
	}
	parl.Log(parl.ShortSpace() + "\x20" + x.Program + ": exit status " + strconv.Itoa(statusCode0)) // outputs "060102 15:04:05Z07 " without newline to stderr
	pos.Exit(statusCode0, nil)                                                                      // os.Exit(1) outputs "exit status 1" to stderr
}

// printErr prints an error
func (x *Executable) printErr(err error, panicString ...string) {
	var s string

	// get panic string
	if len(panicString) > 0 {
		s = panicString[0]
	}

	// print the error
	if x.IsLongErrors || s != "" {
		s += perrors.Long(err) + "\n"
	} else if x.IsErrorLocation {
		s = perrors.Short(err)
	} else if err != nil {
		s = err.Error()
	}
	parl.Log(s)
}

// usage prints options usage
func (x *Executable) usage() {
	writer := flag.CommandLine.Output()
	var license string
	if x.License != "" {
		license = "License: " + x.License
	}
	fmt.Fprintln(
		writer,
		pstrings.FilteredJoin([]string{
			license,
			pstrings.FilteredJoin([]string{
				x.Program,
				x.Description,
			}, "\x20"),
			usageHeader,
			pstrings.FilteredJoin([]string{
				x.Program,
				optionsSyntax,
				x.ArgumentsUsage,
			}, "\x20"),
		}, "\n"))
	flag.PrintDefaults()
	fmt.Fprintln(writer, helpHelp)
}
