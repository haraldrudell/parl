/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "testing"

func TestCloseChannel(t *testing.T) {
	var value = 3
	var doDrain = true

	var ch chan int
	var err, errp error
	var n int
	var isNilChannel, isCloseOfClosedChannel bool

	// close of nil channel should return isNilChannel true
	ch = nil
	isNilChannel, isCloseOfClosedChannel, n, err = CloseChannel(ch, &errp)
	if !isNilChannel {
		t.Error("isNilChannel false")
	}
	_ = err
	_ = n
	_ = isCloseOfClosedChannel

	// n should return number of items when draining
	ch = make(chan int, 1)
	ch <- value
	isNilChannel, isCloseOfClosedChannel, n, err = CloseChannel(ch, &errp, doDrain)
	if n != 1 {
		t.Errorf("n bad %d exp %d", n, 1)
	}
	_ = isNilChannel
	_ = err
	_ = isCloseOfClosedChannel

	// close of closed channel should set isCloseOfClosedChannel, err, errp
	ch = make(chan int)
	close(ch)
	isNilChannel, isCloseOfClosedChannel, n, err = CloseChannel(ch, &errp)
	if !isCloseOfClosedChannel {
		t.Error("isCloseOfClosedChannel false")
	}
	if err == nil {
		t.Error("isCloseOfClosedChannel err nil")
	}
	if errp == nil {
		t.Error("isCloseOfClosedChannel errp nil")
	}
	_ = isNilChannel
	_ = n
}
