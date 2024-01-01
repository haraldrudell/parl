/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pterm

import (
	"bytes"
	"io"
	"os"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/term"
)

const (
	// default prompt is “password: ”
	DefaultPrompt = ""
)

const (
	// prompt suffix
	promptSuffix = ": "
	// default prompt is “password: ”
	defaultPrompt = "password" + promptSuffix
	badFd         = ^(uintptr(0))
)

var termReadPasswordHook func(fd int) ([]byte, error)

// Password implements reading a password from standard input
type Password struct {
	// password prompt like “password”
	//	- colon-space is appended
	Prompt string
	// Input is where characters are read by [term.ReadPassword]
	//	- must have file descriptor, so [os.File]
	//	- default is [os.Stdin]
	Input *os.File
	// Output is where prompt is written using [io.Write]
	//	- if nil, [parl.Consolew] is used for standard error
	Output io.Writer
}

// NewPassword returns an interactive password reader
//   - prompt is printed with appended colon-space
//   - default prompt is “password: ”
//   - not compatible with an active [StatusTerminal]
func NewPassword(prompt string) (passworder *Password) { return &Password{Prompt: prompt} }

// HasPassword returns true indicating that interactive password input is available
func (p *Password) HasPassword() (hasPassword bool) { return true }

// Password reads a password interactively form the keyboard
func (p *Password) Password() (password []byte, err error) {
	var fd int
	if fd, err = p.inputFd(); err != nil {
		return // bad input file descriptor return
	} else if err = p.print(p.prompt()); err != nil {
		return // prompt output failed return
	}
	var pwdBytes []byte
	if hook := termReadPasswordHook; hook == nil {
		pwdBytes, err = term.ReadPassword(fd)
	} else {
		pwdBytes, err = hook(fd)
	}
	if perrors.IsPF(&err, "ReadPassword %w", err) {
		return // ReadPassword failed return
	}
	password = bytes.TrimSpace(pwdBytes)

	return // success return
}

// prompt returns promt ending with colon-space
func (p *Password) prompt() (prompt string) {
	if p := p.Prompt; p != "" {
		prompt = p + promptSuffix
		return
	}
	prompt = defaultPrompt

	return
}

// print outputs the prompt
func (p *Password) print(s string) (err error) {
	var o = p.Output
	if o == nil {
		parl.Consolew(s)
		return
	}
	var byts = []byte(s)
	var n int
	if n, err = o.Write([]byte(s)); perrors.IsPF(&err, "prompt write %w", err) {
		return
	} else if n != len(byts) {
		err = perrors.NewPF("prompt write bad write")
	}

	return
}

// inputFd returns the file descriptor for input
func (p *Password) inputFd() (fd int, err error) {

	// get reader
	var input *os.File
	if p.Input != nil {
		input = p.Input
	} else {
		input = os.Stdin
	}

	// integer Unix file descriptor referencing the open file
	var fdp = input.Fd()
	if fdp == badFd {
		err = perrors.NewPF("invalid input file descriptor")
		return
	}
	fd = int(fdp)

	return
}
