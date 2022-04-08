/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/parlp"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/plog"
	"github.com/haraldrudell/parl/pos"
	"github.com/haraldrudell/parl/pruntime"
	"github.com/haraldrudell/parl/pstrings"
	"gopkg.in/yaml.v2"
)

const (
	rfcTimeFormat = "2006-01-02 15:04:05-07:00"
	usageHeader   = "Usage:"
	optionsSyntax = "[options…]"
	helpHelp      = "\x20\x20-help -h --help\n\x20\x20\tShows this help"
	timeHeader    = "time: %s"
	hostHeader    = "host: %s"
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
	now := parlp.ProcessStartTime()
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

// ParentDir gets absolute path of executable parent directory
func ParentDir() (dir string) {
	if dir = ExecDir(); dir != "" {
		dir = filepath.Dir(dir)
	}
	return
}

// ExecDir gets abolute path to directory where executable is located
func ExecDir() (dir string) {
	if len(os.Args) == 0 {
		return
	}
	dir = Abs(filepath.Dir(os.Args[0]))
	return
}

// Abs ensures a file system path is fully qualified.
// Abs is single-return-value and panics on troubles
func Abs(dir string) (out string) {
	var err error
	if out, err = filepath.Abs(dir); err != nil {
		panic(parl.Errorf("filepath.Abs: '%w'", err))
	}
	return
}

// UserHomeDir gets file system path of user’s home directory
func UserHomeDir() (dir string) {
	var err error
	if dir, err = os.UserHomeDir(); err != nil {
		panic(parl.Errorf("os.UserHomeDir: '%w'", err))
	}
	return
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
		ex.License,
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

/*
ApplyYaml adds options read from a yaml file.

File: The filename searched comes from Executable.Program or the yamlFile argument.
If yamlFile is set exactly this file is read. If yamlFile is not set, A search is executed in the
directories ~/apps .. and /etc.
The filename searched is [program]-[hostname].yaml and [program].yaml.

Content: If the file does not eixst, no action is taken. If the file is empty,
no action is taken.

The top entry in the yaml file is expected to be a dictionary of dictionaries.
The key searched for in the
top level dictionary is the yamlKey argument or “options” if not set.

thunk needs to in otuside of the library for typoing reasons and is
implemented similar to the below. y is the variable receiving parsed yaml data. The om list
is then svcaned for its Y pointers to copy yaml settings to options.
 func applyYaml(bytes []byte, unmarshal mains.UnmarshalFunc, yamlDictionaryKey string) (hasData bool, err error) {
   yamlContentObject := map[string]*YamlData{} // need map[string] because yaml top level is dictionary
   if err = unmarshal(bytes, &yamlContentObject); err != nil {
     return
   }
   yamlDataPointer := yamlContentObject[yamlDictionaryKey] // pick out the options dictionary value
   if hasData = yamlDataPointer != nil; !hasData {
     return
   }
   y = *yamlDataPointer
   return
 }
*/
func (ex *Executable) ApplyYaml(yamlFile, yamlKey string, thunk UnmarshalThunk, om []OptionData) {
	if thunk == nil {
		panic(parl.New("mains.Executable.ApplyYaml: thunk cannot be nil"))
	}
	parl.Debug("Arguments: yamlFile: %q yamlKey: %q", yamlFile, yamlKey)
	filename, byts := FindFile(yamlFile, ex.Program)
	if filename == "" || len(byts) == 0 {
		parl.Debug("ex.ApplyYaml: no yaml file")
		return
	}
	yamlDictionaryKey := GetTopLevelKey(yamlKey) // key name from option or a default
	parl.Debug("filename: %q top-level key: %q bytes: %q", filename, yamlDictionaryKey, string(byts))

	// try to obtain the list of defined keys in the options dictionary
	var yamlVisitedKeys map[string]bool
	yco := map[string]map[string]interface{}{} // a dictionary of dictionaries with unknown content
	parl.Debug("ex.ApplyYaml: first yaml.Unmarshal")
	if yaml.Unmarshal(byts, &yco) == nil {
		yamlVisitedKeys = map[string]bool{}
		if optionsMap := yco[yamlDictionaryKey]; optionsMap != nil {
			for key := range optionsMap {
				yamlVisitedKeys[strings.ToLower(key)] = true
			}
		}
	}
	parl.Debug("ex.ApplyYaml: yamlVisitedKeys: %v\n", yamlVisitedKeys)

	hasData, err := thunk(byts, yaml.Unmarshal, yamlDictionaryKey)
	if err != nil {
		ex.AddErr(parl.Errorf("ex.ApplyYaml thunk: filename: %q: %w", filename, err)).Exit()
	} else if !hasData {
		return
	}

	// iterate over options
	// ignore if no yaml key
	// ignore if yamlVisitedKeys exists and do not have the option
	visitedOptions := map[string]bool{}
	flag.Visit(func(fp *flag.Flag) { visitedOptions[fp.Name] = true })
	for _, optionData := range om {
		if visitedOptions[optionData.Name] || // was specified on command line
			optionData.Y == nil { // does not have yaml value
			continue
		}
		if yamlVisitedKeys != nil { // we have a map of the yaml keys present
			if !yamlVisitedKeys[optionData.Y.Name] {
				continue // this key was not present in yaml
			}
		} else if pstrings.IsDefaultValue(optionData.Y.Pointer) {
			continue // no visited information,, so ignore default values
		}
		if err := optionData.ApplyYaml(); err != nil {
			ex.AddErr(err)
			ex.Exit()
		}
	}
}

func (ex *Executable) usage() {
	writer := flag.CommandLine.Output()
	fmt.Fprintln(
		writer,
		pstrings.FilteredJoin([]string{
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
func (ex *Executable) Recover() {
	if e := recover(); e != nil {
		hasStack := false
		if err, ok := e.(error); ok {
			hasStack = perrors.HasStack(err)
		}
		parl.Debug("exe.Recover: executable %s panic: %T hasStack: %t '%+[2]v'", ex.Program, e, hasStack)
		fmt.Println("Unhandled panic invoked exe.Recover: stack:")
		debug.PrintStack()
		var err error
		var ok bool
		if err, ok = e.(error); !ok {
			err = fmt.Errorf("non-error value: %T %[1]v", e)
		}
		err = parl.Errorf("Unhandled panic to exe.Recover: '%w'", err)
		ex.AddErr(err)
		ex.Exit()
	}
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
