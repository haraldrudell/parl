/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/plog"
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
	defaultOK     = " completed successfully"
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

/*
Executable constant strings that describes an executable
advisable static values include Program Version Comment Description Copyright License Arguments
like:
 var exe = mains.Executable{
   Program:     "getip",
   Version:     "0.0.1",
   Comment:     "first version",
   Description: "finds ip address for hostname",
   Copyright:   "© 2020-present Harald Rudell <harald.rudell@gmail.com> (http://www.haraldrudell.com)",
   License:     "All rights reserved",
   Arguments:   mains.NoArguments | mains.OneArgument,
 }
*/
type Executable struct {
	Program        string // gonet
	Version        string // 0.0.1
	Comment        string // [ banner text after program and version] options parsing (about this version)
	Description    string // [Description part of usage] configures firewall and routing
	Copyright      string // © 2020…
	License        string // ISC License
	OKtext         string // Completed successfully
	ArgumentsUsage string
	Arguments      ArgumentSpec // eg. NoArguments
	// fields below popualted by .Init()
	Launch          time.Time
	LaunchString    string
	Host            string // short hostname, ie. no dots
	err             error
	ArgCount        int
	Arg             string
	Args            []string
	IsLongErrors    bool
	IsErrorLocation bool
}

/*
Init populate launch time and sets silence if first argument is “-silent.”
Init supports function chaining like:
 exe.Init().
   PrintBannerAndParseOptions(optionData).
   LongErrors(options.Debug, options.Verbosity != "").
   ConfigureLog().
   ApplyYaml(options.YamlFile, options.YamlKey, applyYaml, optionData)
   …
*/
func (ex *Executable) Init() *Executable {
	now := ProcessStartTime()
	ex.Launch = now
	ex.LaunchString = now.Format(rfcTimeFormat)
	ex.Host = pos.ShortHostname()
	if len(os.Args) > 1 && os.Args[1] == SilentString {
		parl.SetSilent(true)
	}
	return ex
}

/*
LongErrors sets if errors are printed with stack trace and values. LongErrors
supports functional chaining:
 exe.Init().
   …
   LongErrors(options.Debug, options.Verbosity != "").
   ConfigureLog()…

isLongErrors prints full stack traces, related errors and error data in string
lists and string maps.

isErrorLocation appends the innermost location to the error message when isLongErrors
is not set:
 error-message at error116.(*csTypeName).FuncName-chainstring_test.go:26
*/
func (ex *Executable) LongErrors(isLongErrors bool, isErrorLocation bool) *Executable {
	parl.Debug("exe.LongErrors long: %t location: %t", isLongErrors, isErrorLocation)
	ex.IsLongErrors = isLongErrors
	ex.IsErrorLocation = isErrorLocation
	return ex
}

/*
PrintBannerAndParseOptions prints greeting like:
 parl 0.1.0 parlca https server/client udp server
It then parses options described by []OptionData stroing the values at OptionData.P.
If options fail to parse, a proper message is printed to stderr and the process exits
with status code 2. PrintBannerAndParseOptions supports functional chaining like:
 exe.Init().
   PrintBannerAndParseOptions(…).
   LongErrors(…
Options and yaml is configured likeso:
 var options = &struct {
   noStdin bool
   *mains.BaseOptionsType
 }{BaseOptionsType: &mains.BaseOptions}
 var optionData = append(mains.BaseOptionData(exe.Program, mains.YamlYes), []mains.OptionData{
   {P: &options.noStdin, Name: "no-stdin", Value: false, Usage: "Service: do not use standard input", Y: mains.NewYamlValue(&y, &y.NoStdin)},
 }...)
 type YamlData struct {
   NoStdin bool // nostdin: true
 }
 var y YamlData
*/
func (ex *Executable) PrintBannerAndParseOptions(om []OptionData) (ex1 *Executable) {
	ex1 = ex
	// print program name and populated details
	banner := pstrings.FilteredJoin([]string{
		pstrings.FilteredJoin([]string{ex.Program, ex.Version, ex.Comment}, "\x20"),
		ex.Copyright,
		fmt.Sprintf(timeHeader, ex.LaunchString),
		fmt.Sprintf(hostHeader, ex.Host),
	}, "\n")
	if len(banner) != 0 {
		parl.Info(banner)
	}

	// parse options
	flag.Usage = ex.usage
	omLen := len(om)
	for i := 0; i < omLen; i++ {
		(&om[i]).AddOption()
	}
	flag.Parse()

	// parse arguments
	args := flag.Args() // command-line arguments not part of flags
	count := len(args)
	argsOk :=
		count == 0 && (ex.Arguments&NoArguments != 0) ||
			count == 1 && (ex.Arguments&OneArgument != 0) ||
			count > 0 && (ex.Arguments&ManyArguments != 0)
	if !argsOk {
		if count == 0 {
			if ex.Arguments&ManyArguments != 0 {
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
		ex.usage()
		pos.Exit(pos.StatusCodeUsage, nil)
	}
	ex.ArgCount = count
	if count == 1 && (ex.Arguments&OneArgument != 0) {
		ex.Arg = args[0]
	}
	if count > 0 && (ex.Arguments&ManyArguments != 0) {
		ex.Args = args
	}
	return
}

// ConfigureLog configures the default log such as parl.Log parl.Out parl.D
// for silent, debug and regExp.
// Settings come from BaseOptions.Silent and BaseOptions.Debug.
//
// ConfigureLog supports functional chaining like:
//  exe.Init().
//    …
//    ConfigureLog().
//    ApplyYaml(…)
func (ex *Executable) ConfigureLog() (ex1 *Executable) {
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
	return ex
}

func (ex *Executable) usage() {
	writer := flag.CommandLine.Output()
	var license string
	if ex.License != "" {
		license = "License: " + ex.License
	}
	fmt.Fprintln(
		writer,
		pstrings.FilteredJoin([]string{
			license,
			pstrings.FilteredJoin([]string{
				ex.Program,
				ex.Description,
			}, "\x20"),
			usageHeader,
			pstrings.FilteredJoin([]string{
				ex.Program,
				optionsSyntax,
				ex.ArgumentsUsage,
			}, "\x20"),
		}, "\n"))
	flag.PrintDefaults()
	fmt.Fprintln(writer, helpHelp)
}

/*
Recover function to be used in main.main:
  func main() {
    defer Recover()
    …
On panic, the function prints to stderr: "Unhandled panic invoked exe.Recover: stack:"
followed by a stack trace. It then adds an error to mains.Executable and terminates
the process with status code 1
*/
func (ex *Executable) Recover(errp ...*error) {

	// get error from *errp
	if len(errp) > 0 {
		if errp0 := errp[0]; errp0 != nil {
			if err := *errp0; err != nil {
				ex.AddErr(err)
			}
		}
	}

	// check for panic
	if e := recover(); e != nil {
		hasStack := false
		if err, ok := e.(error); ok {
			hasStack = perrors.HasStack(err)
		}
		parl.Debug("exe.Recover: executable %s panic: %T hasStack: %t '%+[2]v'", ex.Program, e, hasStack)
		parl.Log("Unhandled panic invoked exe.Recover: stack:")
		debug.PrintStack()
		var err error
		var ok bool
		if err, ok = e.(error); !ok {
			err = fmt.Errorf("non-error value: %T %[1]v", e)
		}
		err = parl.Errorf("Unhandled panic to exe.Recover: '%w'", err)
		ex.AddErr(err)
	}

	// print completed successfully
	if ex.err == nil && ex.OKtext != NoOK {
		var s string
		if ex.OKtext != "" {
			s = ex.OKtext
		} else {
			s = ex.Program + defaultOK
		}
		parl.Log(s)
	}

	ex.Exit()
}

// AddErr extended with immediate printing of first error
func (ex *Executable) AddErr(err error) (x *Executable) {
	if parl.IsThisDebug() {
		packFunc := perrors.PackFunc()
		var errS string
		if err != nil {
			errS = "\x27" + err.Error() + "\x27"
		} else {
			errS = "nil"
		}
		parl.Debug("\n%s(error: %s)\n%[1]s invocation:\n%[3]s", packFunc, errS, pruntime.Invocation(0))
		plog.GetLog(os.Stderr).Output(0, "") // newline after debug location. No location appended to this printout
	}
	x = ex
	if err == nil {
		return
	}
	if ex.err == nil {
		ex.PrintErr(err)
		ex.err = err
		return
	}
	ex.err = perrors.AppendError(ex.err, err)
	return
}

// PrintErr prints an error
func (ex *Executable) PrintErr(err error) {
	var s string
	if ex.IsLongErrors {
		s = perrors.Long(err) + "\n"
	} else if ex.IsErrorLocation {
		s = perrors.Short(err)
	} else if err != nil {
		s = err.Error()
	}
	parl.Log(s)
}

// Exit terminate from mains.err: exit 0 or echo to stderr and status code 1
func (ex *Executable) Exit() {
	if ex.err == nil {
		parl.Debug("\nexe.Exit: no error")
	} else {
		parl.Debug("\nexe.Exit: err: %T '%[1]v'", ex.err)
	}
	parl.Debug("\nexe.Exit invocation:\n%s\n", debug.Stack())
	if parl.IsThisDebug() { // add newline during debug without location
		plog.GetLog(os.Stderr).Output(0, "") // newline after debug location. No location appended to this printout
	}
	if ex.err == nil {
		pos.Exit0()
	}
	errorList := perrors.ErrorList(ex.err)
	isList := len(errorList) > 1
	if isList {
		for _, e := range errorList[1:] {
			ex.PrintErr(e)
		}
	}
	if ex.IsLongErrors || isList {
		fmt.Fprintln(os.Stderr, ex.err)
	}
	pos.Exit1(nil)
}
