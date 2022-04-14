/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// zpushparl pushes parl and sub-packages to github.com/haraldrudell/parl.
//  cd zbuild
//  go run .
package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/mains"
	"github.com/haraldrudell/parl/perrors"
)

const (
	alwaysErrorLocation = true
)

var exe = mains.Executable{
	Program:   "pushparl",
	Version:   "0.0.1",
	Comment:   "First version",
	Copyright: "© 2022 Harald Rudell <harald.rudell@gmail.com> (https://github.com/haraldrudell)",
	Arguments: mains.NoArguments,
}

var options = &struct {
	*mains.BaseOptionsType
}{BaseOptionsType: &mains.BaseOptions}

var optionData = mains.BaseOptionData(exe.Program, mains.YamlNo)

func main() {
	defer exe.Recover()
	exe.Init().
		PrintBannerAndParseOptions(optionData).
		ConfigureLog().
		LongErrors(options.Debug, alwaysErrorLocation)
	ctx := context.Background()
	_ = ctx

	var err error
	for {
		var out string

		// make sure there are no pending git changes
		if out, err = run([]string{"git", "status", "--porcelain"}); err != nil {
			break
		}
		if out != "" {
			err = perrors.New("git has unsaved changes")
			break
		}

		// switch to master
		if out, err = run([]string{"git", "checkout", "master"}); err != nil {
			break
		}

		// test
		if out, err = run([]string{"go", "test", "./..."}); err != nil {
			break
		}
	}
	if err != nil {
		exe.AddErr(err)
	}
	exe.Exit()
}

func run(cmd []string) (out string, err error) {

	cmdString := strings.Join(cmd, "\x20")
	parl.Out(cmdString)

	// get the parl project directory
	var dir string
	if dir, err = os.Getwd(); err != nil {
		err = perrors.Errorf("os.Getwd: %w", err)
		return
	}
	dir = filepath.Join(dir, "..")

	// go run main.go dir/main.go
	// byts has a terminating newline
	// if dir is not specified, it is that of the parent process, ie. the goprogramming directory
	// output:
	// named files must all be in one directory; have . and dir
	var execCmd *exec.Cmd
	execCmd = exec.Command(cmd[0], cmd[1:]...)
	execCmd.Dir = dir
	var byts []byte

	byts, err = execCmd.CombinedOutput()
	out = string(byts)
	if err != nil {
		err = perrors.Errorf("Command: %q out: %q err: %v", cmdString, out, err)
		return
	}

	return
}
