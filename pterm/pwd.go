/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pterm

import (
	"fmt"
	"os"
	"strings"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/term"
)

type PasswordStdio struct {
	Prompt string
	Input  *os.File
	Output *os.File
}

func NewPassword(prompt string) (ps *PasswordStdio) {
	return &PasswordStdio{Prompt: prompt}
}

func (ps *PasswordStdio) HasPassword() (hasPassword bool) {
	return true
}

func (ps *PasswordStdio) Password() (password string) {
	var bytes []byte
	var err error
	ps.printf("%s: ", ps.Prompt)
	if bytes, err = term.ReadPassword(ps.fd()); err != nil {
		panic(perrors.Errorf("ReadPassword: '%w'", err))
	}
	password = strings.TrimSpace(string(bytes))
	return
}

func (ps *PasswordStdio) printf(format string, a ...interface{}) {
	if ps.Output == nil {
		parl.Console(format, a...)
	} else if _, err := fmt.Fprintf(ps.Output, format, a...); err != nil {
		panic(perrors.Errorf("fmt.Fprintf: '%w'", err))
	}
}

func (ps *PasswordStdio) fd() (no int) {
	// get reader
	var input *os.File
	if ps.Input != nil {
		input = ps.Input
	} else {
		input = os.Stdin
	}
	fdp := input.Fd()
	if fdp == ^(uintptr(0)) {
		panic(perrors.New(perrors.PackFunc() + ": invalid file descriptor"))
	}
	return int(fdp)
}
