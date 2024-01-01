/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pterm

import (
	"bytes"
	"slices"
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

func TestPassword_Password(t *testing.T) {
	var h = termReadPasswordHook
	defer func() { termReadPasswordHook = h }()
	var prompt = "hello"
	var pwd = []byte("1234")

	var input = newFakeTerm([]byte(pwd))
	termReadPasswordHook = input.ReadPassword
	var output bytes.Buffer
	var err error
	var pwdAct []byte

	// HasPassword() Password()
	var password *Password = NewPassword(prompt)

	// HasPassword should be true
	if !password.HasPassword() {
		t.Error("HasPassword false")
	}

	// Password should return password
	password.Output = &output
	pwdAct, err = password.Password()
	if err != nil {
		t.Errorf("Password err: %s", perrors.Short(err))
	}
	if !slices.Equal(pwdAct, pwd) {
		t.Errorf("password: %q exp %q", pwdAct, pwd)
	}
}

// fakeTerm provides hook function for [term.ReadPassword]
type fakeTerm struct{ input []byte }

// newFakeTerm returns hook function object for [term.ReadPassword]
func newFakeTerm(input []byte) (t *fakeTerm) { return &fakeTerm{input: input} }

// fake [term.ReadPassword]
func (t *fakeTerm) ReadPassword(fd int) (line []byte, err error) {
	line = slices.Clone(t.input)
	return
}
