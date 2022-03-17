/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License

Demonstrates usage of the github.com/haraldrudell/parl go package
parl provides:
— handling of command-line options and defaults loaded from yaml files
— managing long-running goroutine tasks with status and panic handling
— predictable parallel programing mechanics
— generic interfaces to operating systems, file systems, SQLite and time
— parl can be used to implement Linux and macOS services

execute on-the-fly:
go run ./cmd/parl
*/

// parl.go demonstrate usage of the parl package, a go library for command-line utilities and concurrency
package main

import (
	"errors"
	"strings"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/mains"
)

// exe describes the executable
var exe = mains.Executable{
	Program:     "parl",
	Version:     "0.1.0",
	Comment:     "parlca https server/client udp server",
	Description: "demonstrates the parl package: github.com/haraldrudell/parl",
	Copyright:   "© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://github.com/haraldrudell)",
	Arguments:   mains.NoArguments,
}

// options contains option values parsed from command line arguments
var options = &struct {
	error   bool
	panic   bool
	noStdin bool
	*mains.BaseOptionsType
}{BaseOptionsType: &mains.BaseOptions}

// optionData describes available command-line options
var optionData = append(mains.BaseOptionData(exe.Program, mains.YamlYes), []mains.OptionData{
	{P: &options.panic, Name: "panic", Value: false, Usage: "parl has one error then panics"},
	{P: &options.error, Name: "error", Value: false, Usage: "parl exits with error"},
	{P: &options.noStdin, Name: "no-stdin", Value: false, Usage: "Service: do not use standard input", Y: mains.NewYamlValue(&y, &y.NoStdin)},
}...)

// YamlData describes the top-level options key of a yaml options file
type YamlData struct {
	NoStdin bool // nostdin: true
}

var y YamlData

func main() {

	// Start command-line executable
	defer exe.Recover()
	exe.Init().
		PrintBannerAndParseOptions(optionData).
		LongErrors(options.Debug, options.Verbosity != "").
		ConfigureLog().
		ApplyYaml(options.YamlFile, options.YamlKey, applyYaml, optionData)
	ctx := parl.NewContext()

	if options.error {
		exe.AddErr(parl.Errorf("Single error")).Exit()
	}
	if options.panic {
		exe.AddErr(parl.Errorf("First error"))
		panic(errors.New("panic error"))
	}

	parl.Info(strings.Join([]string{
		"",
		"Logging:",
		"Parl.Console is for interactivity",
		"Parl.Out is for executable usable output",
		"Parl.Log always prints",
		"Parl.Info can be silenced with -silent",
		"Parl.Debug only prints with -debug or -verbose",
		"— parl logging uses thread-safe log.Output",
		"\n",
	}, "\n"))

	parl.Log("parl.Log always prints, gets location with -debug")
	parl.Info("parl.Info output can be silenced with -silent or get location with -debug")
	parl.Debug("parl.Debug only prints with -debug or with package filter -verbose")
	parl.D("parl.D is used for temporary debug printing, not to be checked in")

	parl.Info(strings.Join([]string{
		"",
		"Try:",
		"parl -help",
		"parl -silent",
		"parl -error",
		"parl -panic",
		"parl -debug",
		"parl -debug -errors",
		"parl -verbose main.main",
		"\n",
	}, "\n"))

	// Demonstrate ad-hoc goroutine
	parl.Info("Launching 1-second goroutine…")
	ch := make(chan *string) // result type
	che := make(chan error)
	go func() {
		defer parl.Recover("parl-goroutine1", nil, func(err error) { che <- err })
		time.Sleep(time.Second)
		result := "success"
		ch <- &result
	}()
	select {
	case value := <-ch: // goroutine success
		parl.Info("ad-hoc goroutine result: %s\n", *value)
	case err := <-che: // failure from goroutine
		exe.AddErr(err)
		exe.Exit()
	case <-ctx.Done(): // application terminated
		exe.Exit()
	}

	// Demonstrate sub-task
	//
}

func applyYaml(bytes []byte, unmarshal mains.UnmarshalFunc, yamlDictionaryKey string) (hasData bool, err error) {
	parl.Debug("applyYaml: bytes: %d key: %q\nbytes: %q\n", len(bytes), yamlDictionaryKey, string(bytes))

	yamlContentObject := map[string]*YamlData{}                 // need map[string] because yaml top level is dictionary
	if err = unmarshal(bytes, &yamlContentObject); err != nil { // populate mapp
		err = parl.Errorf("applyYaml unmarshal: '%w'", err)
		return
	}
	parl.Debug("applyYaml: object: %+v\n", yamlContentObject) // map[options:0xc0000723c0]
	yamlDataPointer := yamlContentObject[yamlDictionaryKey]   // pick out the options dictionary value
	if hasData = yamlDataPointer != nil; !hasData {
		parl.Debug("applyYaml: yaml key had no data")
		return
	}
	y = *yamlDataPointer
	parl.Debug("applyYaml: data: %+v\n", y)
	return
}
