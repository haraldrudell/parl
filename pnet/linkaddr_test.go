/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

func TestLinkAddr_Interface(t *testing.T) {
	var a = &LinkAddr{Name: "a b"}

	var netInterface *net.Interface
	var isNoSuchInterface bool
	var err error

	netInterface, isNoSuchInterface, err = a.Interface()

	if err == nil {
		t.Fatal("LinkAddr.Interface no error")
	}
	if !isNoSuchInterface {
		t.Errorf("isNoSuchInterface false: err: %q", perrors.Short(err))
	}
	if netInterface != nil {
		t.Error("netInterface not nil")
	}

	// err: *errorglue.errorStack *fmt.wrapError *net.OpError *errors.errorString
	//t.Logf("err: %s", errorglue.DumpChain(err))
}
