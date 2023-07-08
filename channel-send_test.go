/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

func TestChannelSend(t *testing.T) {
	var errCh = make(chan error, 1)
	var err0 error

	var didSend, isNilChannel, isClosedChannel bool
	var err error

	// successful send
	didSend, isNilChannel, isClosedChannel, err = ChannelSend(errCh, err0)
	if !didSend {
		t.Error("didSend false")
	}
	if isNilChannel {
		t.Error("isNilChannel true")
	}
	if isClosedChannel {
		t.Error("isClosedChannel true")
	}
	if err != nil {
		t.Errorf("ChannelSend err: %s", perrors.Short(err))
	}
}

func TestChannelSendNil(t *testing.T) {
	var errCh chan error
	var err0 error

	var didSend, isNilChannel, isClosedChannel bool
	var err error

	// nil send
	didSend, isNilChannel, isClosedChannel, err = ChannelSend(errCh, err0)
	if didSend {
		t.Error("didSend true")
	}
	if !isNilChannel {
		t.Error("isNilChannel false")
	}
	if isClosedChannel {
		t.Error("isClosedChannel true")
	}
	if err == nil {
		t.Error("ChannelSend missing err")
	}
}

func TestChannelSendClosed(t *testing.T) {
	var errCh = make(chan error, 1)
	var err0 error

	var didSend, isNilChannel, isClosedChannel bool
	var err error

	// closed send
	close(errCh)
	didSend, isNilChannel, isClosedChannel, err = ChannelSend(errCh, err0)
	if didSend {
		t.Error("didSend true")
	}
	if isNilChannel {
		t.Error("isNilChannel true")
	}
	if !isClosedChannel {
		t.Error("isClosedChannel false")
	}
	if err == nil {
		t.Error("ChannelSend missing err")
	}
}

func TestChannelSendNonBlocking(t *testing.T) {
	var errCh = make(chan error)
	var err0 error

	var didSend, isNilChannel, isClosedChannel bool
	var err error

	// non-blocking send
	didSend, isNilChannel, isClosedChannel, err = ChannelSend(errCh, err0, SendNonBlocking)
	if didSend {
		t.Error("didSend true")
	}
	if isNilChannel {
		t.Error("isNilChannel true")
	}
	if isClosedChannel {
		t.Error("isClosedChannel true")
	}
	if err != nil {
		t.Errorf("ChannelSend err: %s", perrors.Short(err))
	}
}
