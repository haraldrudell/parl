/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"
	"slices"
	"testing"
	"time"

	"github.com/haraldrudell/parl/perrors"
)

func TestThreadSafeRwc(t *testing.T) {
	//t.Error("logging on")
	const (
		resultLength = 10
		// timeout when debugging
		//timeout = 0
		// timeout when testing
		timeout = time.Millisecond
	)
	var (
		byte1 = []byte{1}
	)

	var (
		result *readResult
		err    error
		n      int
	)

	// Write Read Close
	var reader = NewThreadSafeRwc().(*ThreadSafeRwc)

	// Read should await data
	result = newReadResult(resultLength)
	go result.read(reader)
	<-result.isReady
	if timeout > 0 {
		<-time.NewTimer(timeout).C
	}
	if result.isExit.Load() {
		t.Fatal("Read not blocking")
	}
	// Read should return data from susbequent Write
	n, err = reader.Write(byte1)
	_ = n
	_ = err
	if !result.awaitCh(timeout) {
		t.Fatal("Read timeout awaiting Write")
	}
	if !slices.Equal(result.p, byte1) {
		t.Errorf("Read %v exp %v", result.p, byte1)
	}

	// Read should return io.EOF on Close
	result = newReadResult(resultLength)
	go result.read(reader)
	<-result.isReady
	err = reader.Close()
	_ = err
	if !result.awaitCh(timeout) {
		t.Fatal("Read timeout awaiting Close")
	}
	if result.err != io.EOF {
		t.Errorf("Read with Close err not io.EOF: %s", perrors.Short(result.err))
	}
}
