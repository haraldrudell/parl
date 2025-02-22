/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/mains/malib"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/perrors/errorglue"
	"github.com/haraldrudell/parl/pflags"
	"github.com/haraldrudell/parl/plog"
	"github.com/haraldrudell/parl/pos"
	"github.com/haraldrudell/parl/pruntime"
	"github.com/haraldrudell/parl/pstrings"
)

const (
	// if [Executable.OKtext] is assigned NoOK there ir no successful message on app exit
	NoOK = "-"
	// error location is not appended to errors printed without stack trace
	//	- second argument to [Executable.LongErrors]
	NoErrorLocationTrue = false
	// always output error stack traces
	//	- first argument to [Executable.LongErrors]
	AlwaysStackTrace = true
	// stack traces are not output for errors
	//	- stack traces are printed for panic
	//	- first argument to [Executable.LongErrors]
	NoStackTrace = false
)

const (
	rfcTimeFormat = "2006-01-02 15:04:05-07:00"
	usageHeader   = "Usage:"
	optionsSyntax = "[options…]"
	helpHelp      = "\x20\x20-help -h --help\n\x20\x20\tShows this help"
	timeHeader    = "time: %s"
	hostHeader    = "host: %s"
	defaultOK     = "completed successfully"
	// count (Recover or EarlyPanic) and doPanicFrames
	//	- must have at least one of the two, so use 1
	doPanicFrames = 1
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

	// errors addded by AddErr and other methods
	//	- because an error added may have associated errors,
	//		err must be a slice, to distinguish indivdual error adds
	//	- that slice must be thread-safe
	err      malib.ErrStore
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
	// a specific status code to use on exit
	statusCode parl.Atomic64[int]
}

// Executable is an error sink
var _ parl.ErrorSink = &Executable{}

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
	var now = malib.ProcessStartTime()
	x.Launch = now
	x.LaunchString = now.Format(rfcTimeFormat)
	x.Host = pos.ShortHostname()
	if len(os.Args) > 1 && os.Args[1] == SilentString {
		parl.SetSilent(true)
	}
	return
}

// LongErrors configures error output
// sets if errors are printed with stack trace and values. LongErrors
//   - isLongErrors true: prints full stack traces, related errors and error data in string
//     lists and string maps
//   - isLongErrors false: prints one-liner error messages
//   - isErrorLocation missing: error location is output for errors printed without stack trace
//   - isErrorLocation NoLocation: error location is not output
//   - functional chaining
//
// Usage:
//
//	exe.Init().
//	…
//	LongErrors(options.Debug, options.Verbosity != "").
//	ConfigureLog()…
func (x *Executable) LongErrors(isLongErrors bool, isErrorLocation ...ErrLoc) (x2 *Executable) {
	x2 = x

	parl.Debug("exe.LongErrors long: %t location: %t", isLongErrors, isErrorLocation)
	x.IsLongErrors = isLongErrors
	x.IsErrorLocation = len(isErrorLocation) == 0 ||
		isErrorLocation[0] != NoLocation

	return
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
	ex1 = x

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

	return
}

// errEarlyPanicError is a non-nil error value between EarlyPanic and Recover
var errEarlyPanicError = errors.New("mains stored a panic")

// EarlyPanic is a deferred function that recovers panic prior to all other deferred functions
//   - EarlyPanic ensures that a panic is displayed immediately and
//     that the panic is stored as the first occurring error
//   - the first deferred function should be [Executable.Recover]
//   - the last deferred function should be [Executable.EarlyPanic]
//   - EarlyPanic avoids confusion from other deferred functions quietly hanging
//     or storing a first error that is actually subsequent to the panic
func (x *Executable) EarlyPanic(errp ...*error) {

	// print and store any panic returned by recover()
	if !x.processPanicValue(recover()) {
		return // there was no panic
	}
	// a panic was recovered, printed and stored

	// since the panic was stored and printed:
	//	- an error must be returned to maintain the error condition
	//	- — panic cannot be invoked again because it would cancel other deferred functions
	//	- — if the just stored error value is used, this may be modified and lead to error duplication
	//	- therefore, a new simple error is returned
	//	- — once at [Executable.Recover] the simple error can be discarded
	//	- — since EarlyPanic just recovered a panic, EarlyPanic itself would produce a a panic stack,
	//		why a stack trace added to the simple error now would make the error appear as a panic
	//	- — therefore, the simple error has no stack

	// ensure any non-nil errp points to a non-nil error
	if len(errp) > 0 {
		if errp0 := errp[0]; errp0 != nil && *errp0 == nil {
			*errp0 = errEarlyPanicError
		}
	}
}

// processPanicValue prints and stores any non-nil recover() values
//   - the value with its type
//   - stack trace for any error that has stack
//   - recovery stack trace
func (x *Executable) processPanicValue(panicValue any) (wasPanic bool) {

	// ensure -debug honored if panic or error before options parsing
	if !x.optionsWereParsed.Load() {
		for _, option := range os.Args {
			if option == pflags.DebugOption { // -debug
				parl.SetDebug(true)
			}
		}
	}

	// no ongoing panic is noop
	if wasPanic = panicValue != nil; !wasPanic {
		return
	}
	// a panic was recovered in panicValue

	// determine if panicValue implements error
	var err, recoverValueIsError = panicValue.(error)
	// the innermost error if panicValue is error
	var error0 error

	// debug print panicValue
	isDebug := parl.IsThisDebug()
	if isDebug {
		hasStack := false
		var valueString string
		var error0type string
		if !recoverValueIsError {
			valueString = parl.Sprintf(" “%+v”", panicValue)
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
		// panic printout:
		//	- the packFunc for this function
		//	- the type returned by recover()
		//	- innermost error type if panicValue implements error
		//	- true if panicValue was an error with a stack trace
		//	- the value returned by recover():
		//	- — if recover() did not return an error: printed using “%+v”
		//	- — if recover() returned an error with stack: the error with stack and
		//		the innermost error printed using “%+v”
		//	- — if recover() returned an error without stack: the message and
		//		the innermost error printed using “%+v”
		parl.Debug("%s: panic with -debug: recover-value type: %T%s hasStack: %t%s",
			pruntime.NewCodeLocation(0).PackFunc(),
			panicValue, error0type, hasStack, valueString)
	}

	// print recovery stack trace
	var stackString string
	if isDebug {
		stackString = " recovery stack trace:\n\n" + pruntime.DebugStack(0) + "\n\n"
	}
	var programString string
	if x.Program != "" {
		programString = "\x20" + x.Program
	}
	parl.Log("\n\nProgram%s Recovered a main-thread panic:%s", programString, stackString)

	// store recovery value as error
	var prepend string
	var postpend string
	if !recoverValueIsError {
		err = perrors.Errorf("panic: non-error value: %T %[1]v", panicValue)
	} else {
		prepend = "panic: “"
		postpend = "”"
		if isDebug {
			// put error0 type name in error message
			postpend += parl.Sprintf(" type: %T", error0)
		}
	}
	// always add a stack trace after panic
	//	- must contain one frame after panic
	err = perrors.Stackn(err, doPanicFrames)
	err = perrors.Errorf("main-thread %s%w%s", prepend, err, postpend)
	x.AddError(err)

	return
}

// AddError extended with immediate printing of first error
//   - if err is the first error, it is immediately printed
//   - subsequent errors are appended to x.err
//   - err nil: ignored
func (x *Executable) AddError(err error) {

	// debug printing
	if parl.IsThisDebug() {
		packFunc := pruntime.PackFunc()
		var errS string
		if err != nil {
			errS = "\x27" + err.Error() + "\x27"
		} else {
			errS = "nil"
		}
		parl.Debug("\n%s(error: %s)\n%[1]s invocation:\n%[3]s", packFunc, errS, pruntime.Invocation(0))
		plog.GetLog(os.Stderr).Output(0, "") // newline after debug location. No location appended to this printout
	}

	// if AddErr with no error, do nothing
	if err == nil {
		return // no error do nothing return
	}

	// if the first error, immediately print it
	if x.err.Count() == 0 {

		// print and store the first error
		if x.printErr(err, malib.CheckForPanic(err)) {
			x.err.IsFirstLong.Store(true)
		}
	}

	// append subsequent error
	x.err.Add(err)
}

// a subt making [Executable] implement [parl.ErrorSink]
func (x *Executable) EndErrors() {}

// SetStatusCode allos to set the status code to use on exit
//   - deferrable, thread-safe
func (x *Executable) SetStatusCode(statusCode *int) { x.statusCode.Store(*statusCode) }

// Recover terminates the process, typically invoked as the first deferred function in main.main
//   - errp: optional pointer that may point to a non-nil error
//   - Recover does not return, instead it invokes [os.Exit]
//   - Recover recovers any ongoing panic using recover()
//   - successful exit is:
//   - — no ongoing panic
//   - — no non-nil errors pointed to by errp
//   - — no errors previously stored by [Executable.AddError] or [Executable.EarlyPanic]
//   - — no non-zero status code previously stored by [Executable.SetStatusCode]
//   - on successful exit, a timestamped success message is printed followed by exit status zero:
//   - — “gtee completed successfully at 240524 17:26:50-07”
//   - on failure exit:
//   - — non-panic errors are printed as one-liners with code location
//   - — panics are printed with stack trace and error chain
//   - — the last printed line is a timestamped exit message:
//   - — “240524 21:32:08-07 gtee: exit status 1”
//   - — exit status code is any non-zero status code provided to [Executable.SetStatusCode] or 1
//
// Usage:
//
//	func main() {
//	  var err error
//	  defer Recover(&err)
//	  …
func (x *Executable) Recover(errp ...*error) {

	// process remaining error data
	//	— get error from *errp and store in x.err
	if len(errp) > 0 {
		if errp0 := errp[0]; errp0 != nil {
			if err := *errp0; err != nil && err != errEarlyPanicError {
				x.AddError(err)
			}
		}
	}
	// recover any ongoing panic
	x.processPanicValue(recover())

	// exit status code, zero on success:
	//	- errCount zero and
	//	- no non-zero status code set by [Executable.SetStatusCode]
	var statusCode = x.statusCode.Load()
	// the number of errors and panics
	var errCount = x.err.Count()
	// if not success, ensure non-zero status code
	//	- status code 1 is default
	if statusCode == 0 && errCount > 0 {
		statusCode = pos.StatusCodeErr
	}

	// printouts when IsDebug
	if errCount == 0 {
		parl.Debug("\nexe.Exit: no error")
	} else {
		parl.Debug("\nexe.Exit: err: %T '%[1]v'", x.err.GetN())
	}
	parl.Debug("\nexe.Exit invocation:\n%s\n", pruntime.NewStack(0))
	if parl.IsThisDebug() { // add newline during debug without location
		plog.GetLog(os.Stderr).Output(0, "") // newline after debug location. No location appended to this printout
	}

	// successfull exit
	if statusCode == 0 {
		// print timestamped success message if not suppressed
		//	- “gtee completed successfully at 240524 17:26:50-07”
		if x.OKtext != NoOK {
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
		// exit status zero
		pos.OsExit(statusCode)
	}

	// err0 is the first occuring error if any
	var err0 error

	// print all errors except the very first
	//	- if first error value was printed long, it does not have to be printed
	//	- otherwise, print its associated errors
	//	- subsequent errors printed in full
	for i, err := range x.err.Get() {
		if i == 0 {
			err0 = err
			// if first error was printed long, ignore it alltogether
			if x.err.IsFirstLong.Load() {
				continue
			}
		} else {
			// print a subsequent error
			if x.printErr(err, malib.CheckForPanic(err)) {
				continue // printed long means it’s complete
			}
		}

		// now recursively print all associated errors
		x.printAssociated(strconv.Itoa(i+1), err)
	}

	// just before exit, print the one-liner message of the first occurring error again
	if x.IsLongErrors || errCount > 1 {
		fmt.Fprintln(os.Stderr, err0)
	}

	// timestamped staus-code message
	//	- “240524 21:32:08-07 gtee: exit status 1”
	parl.Log(parl.ShortSpace() + "\x20" + x.Program + ": exit status " + strconv.Itoa(statusCode)) // outputs "060102 15:04:05Z07 " without newline to stderr

	// exit with non-zero status code
	pos.Exit(statusCode, nil)
}

// printAssociated recursively prints associated error values
func (x *Executable) printAssociated(i string, err error) {
	var associatedErrors = errorglue.ErrorList(err)
	// 1 means no associated errors
	if len(associatedErrors) < 1 {
		return
	}
	for j, e := range associatedErrors[1:] {
		var label = parl.Sprintf("error#%s-associated#%d", i, j+1)
		parl.Log(label)
		if x.printErr(err, malib.CheckForPanic(err)) {
			continue // printed long means it’s complete
		}
		x.printAssociated(label, e)
	}
}

// printErr prints error to stderr
//   - err is printed using perrors.Long or Short
//   - panicString non-empty: with stack trace, append this panic description
//   - long with stack trace if panicString non-empty or x.IsLongErrors
//   - short has location if x.IsErrorLocation
func (x *Executable) printErr(err error, panicString ...string) (printedLong bool) {
	var s string

	// get panic string
	if len(panicString) > 0 {
		s = panicString[0]
	}

	// print the error
	if x.IsLongErrors || s != "" {
		printedLong = true
		s += perrors.Long(err) + "\n— — —"
	} else if x.IsErrorLocation {
		s = perrors.Short(err)
	} else if err != nil {
		s = err.Error()
	}
	parl.Log(s)
	return
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
