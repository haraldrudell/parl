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
	"github.com/haraldrudell/parl/error116"
	"github.com/haraldrudell/parl/parlos"
	"github.com/haraldrudell/parl/parls"
	"github.com/haraldrudell/parl/runt"
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

// Executable constant strings that describes an executable
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

// Init populate launch time — may throw
func (ex *Executable) Init() *Executable {
	now := time.Now()
	ex.Launch = now
	ex.LaunchString = now.Format(rfcTimeFormat)
	ex.Host = parlos.ShortHostname()
	if len(os.Args) > 1 && os.Args[1] == SilentString {
		parl.SetSilent(true)
	}
	return ex
}

// LongErrors sets if errors are printed with stack trace and values
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

// ExecDir gets directory where executable is located
func ExecDir() (dir string) {
	if len(os.Args) == 0 {
		return
	}
	dir = Abs(filepath.Dir(os.Args[0]))
	return
}

// Abs ensures a file system path is fully qualified
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

// PrintBannerAndParseOptions prints greeting
func (ex *Executable) PrintBannerAndParseOptions(om []OptionData) (ex1 *Executable) {
	ex1 = ex
	// print program name and populated details
	banner := parls.FilteredJoin([]string{
		parls.FilteredJoin([]string{ex.Program, ex.Version, ex.Comment}, "\x20"),
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
		parlos.Exit(parlos.StatusCodeUsage, nil)
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

func (ex *Executable) ConfigureLog() (ex1 *Executable) {
	if BaseOptions.Silent {
		parl.SetSilent(true)
	}
	if BaseOptions.Debug {
		parl.SetDebug(true)
	}
	if BaseOptions.Verbosity != "" {
		if err := parl.SetRegexp(BaseOptions.Verbosity); err != nil {
			parlos.Exit(parlos.StatusCodeUsage, err)
		}
	}
	parl.Debug("exe.ConfigureLog silent: %t debug: %t verbosity: %q\n",
		BaseOptions.Silent, BaseOptions.Debug, BaseOptions.Verbosity)
	return ex
}

func (ex *Executable) ApplyYaml(yamlFile, yamlKey string, thunk UnmarshalThunk, om []OptionData) {
	parl.Debug("exe.ApplyYaml: file: %q", yamlFile)
	filename, bytes := FindFile(yamlFile, ex.Program)
	if filename == "" || len(bytes) == 0 {
		parl.Debug("ApplyYaml: no yaml file")
		return
	}
	yamlDictionaryKey := GetTopLevelKey(yamlKey) // key name from option or a default

	// try to obtain the list of defined keys in the options dictionary
	var yamlVisitedKeys map[string]bool
	yco := map[string]map[string]interface{}{}
	parl.Debug("ApplyYaml: first yaml.Unmarshal")
	if yaml.Unmarshal(bytes, &yco) == nil {
		yamlVisitedKeys = map[string]bool{}
		if optionsMap := yco[yamlDictionaryKey]; optionsMap != nil {
			for key := range optionsMap {
				yamlVisitedKeys[strings.ToLower(key)] = true
			}
		}
	}
	parl.Debug("ApplyYaml: yamlVisitedKeys: %v\n", yamlVisitedKeys)

	hasData, err := thunk(bytes, yaml.Unmarshal, yamlDictionaryKey)
	if err != nil {
		ex.AddErr(parl.Errorf("ApplyYaml thunk: filename: %q: %w", filename, err)).Exit()
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
		if yamlVisitedKeys != nil { // we hve a map of the yaml keys present
			if !yamlVisitedKeys[optionData.Y.Name] {
				continue // this key was not present in yaml
			}
		} else if parls.IsDefaultValue(optionData.Y.Pointer) {
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
		parls.FilteredJoin([]string{
			parls.FilteredJoin([]string{
				ex.Program,
				ex.Description,
			}, "\x20"),
			usageHeader,
			parls.FilteredJoin([]string{
				ex.Program,
				optionsSyntax,
				ex.ArgumentsUsage,
			}, "\x20"),
		}, "\n"))
	flag.PrintDefaults()
	fmt.Fprintln(writer, helpHelp)
}

// Recover function to be deferred in main()
func (ex *Executable) Recover() {
	if e := recover(); e != nil {
		hasStack := false
		if err, ok := e.(error); ok {
			hasStack = error116.HasStack(err)
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
		packFunc := error116.PackFunc()
		var errS string
		if err != nil {
			errS = "\x27" + err.Error() + "\x27"
		} else {
			errS = "nil"
		}
		parl.Debug("%s(error: %s)\n%[1]s invocation: %[3]s", packFunc, errS, runt.Invocation(0))
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
	if errorList, ok := ex.err.(error116.ErrorHasList); ok {
		errorList.Append(err)
		return
	}
	ex.err = error116.AppendError(ex.err, err)
	return
}

// PrintErr prints an error
func (ex *Executable) PrintErr(err error) {
	format := error116.DefaultFormat // format := "%v\n"
	if ex.IsLongErrors {
		format = error116.LongFormat // format = "%+v\n"
	} else if ex.IsErrorLocation {
		format = error116.ShortFormat
	}
	fmt.Fprintln(os.Stderr, error116.ChainString(err, format))
}

// Exit terminate from mains.err: exit 0 or echo to stderr and status code 1
func (ex *Executable) Exit() {
	if ex.err == nil {
		parl.Debug("exe.Exit: no error")
	} else {
		parl.Debug("exe.Exit: err: %T '%[1]v'", ex.err)
	}
	parl.Debug("%s", debug.Stack())
	if ex.err == nil {
		parlos.Exit0()
	}
	errorList, isList := ex.err.(error116.ErrorHasList)
	if isList {
		for _, e := range errorList.ErrorList() {
			ex.PrintErr(e)
		}
	}
	if ex.IsLongErrors || isList {
		fmt.Fprintln(os.Stderr, ex.err)
	}
	parlos.Exit1(nil)
}
