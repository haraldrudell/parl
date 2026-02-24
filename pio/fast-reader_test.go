/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
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

func TestNewFastReader(t *testing.T) {
	//t.Error("logging on")
	const (
		resultLength = 10
		// timeout when debugging
		//timeout = 0
		// timeout when testing
		//	- 200 ms accounts for any thread being suspended
		timeout = 200 * time.Millisecond
	)
	var (
		byte1 = []byte{1}
	)

	var (
		result *readResult
	)

	// Read
	var reader *FastReader
	// Close CloseWithError Write
	var dataSource *io.PipeWriter
	var reset = func() {
		var pipeReader io.Reader
		pipeReader, dataSource = io.Pipe()
		reader = NewFastReader(pipeReader).(*FastReader)
	}

	// Read should await data
	reset()
	result = newReadResult(resultLength)
	go result.read(reader)
	<-result.isReady
	if timeout > 0 {
		<-time.NewTimer(timeout).C
	}
	if result.isExit.Load() {
		t.Fatal("Read not blocking")
	}
	// byte becoming available should unblock Read
	dataSource.Write(byte1)
	if !result.awaitCh(timeout) {
		t.Fatal("Read timeout awaiting Write")
	}
	if !slices.Equal(result.p, byte1) {
		t.Errorf("Read %v exp %v", result.p, byte1)
	}
	// EOF should be returned
	result = newReadResult(resultLength)
	go result.read(reader)
	<-result.isReady
	dataSource.Close()
	if !result.awaitCh(timeout) {
		// TODO 260224 test failed during pushparl
		//	- Read timeout awaiting Close
		t.Fatal("Read timeout awaiting Close")
	}
	if result.err != io.EOF {
		t.Errorf("Read with Close err not io.EOF: %s", perrors.Short(result.err))
	}
}

func TestNewFastReader2Writes(t *testing.T) {
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
		byte2 = []byte{2}
		exp12 = append(slices.Clone(byte1), byte2...)
	)

	var (
		result *readResult
	)

	// Read
	var reader *FastReader
	// Close CloseWithError Write
	var dataSource *io.PipeWriter
	var reset = func() {
		var pipeReader io.Reader
		pipeReader, dataSource = io.Pipe()
		reader = NewFastReader(pipeReader).(*FastReader)
	}

	// two writes should be merged into one Read
	reset()
	dataSource.Write(byte1)
	dataSource.Write(byte2)
	// await fastReaderDrainThread saving the bytes
	//	- it is holding at Read
	//	- must save any read bytes to bufferList
	<-time.NewTimer(time.Millisecond).C
	if length := reader.Length(); length != len(exp12) {
		t.Fatalf("fastReaderDrainThread hung %d exp %d", length, len(exp12))
	}
	result = newReadResult(resultLength)
	go result.read(reader)
	if !result.awaitCh(timeout) {
		t.Fatal("Read timeout when data available")
	}
	if !slices.Equal(result.p, exp12) {
		t.Errorf("Read %v exp %v", result.p, exp12)
	}
}

func TestNewFastReaderOffloading(t *testing.T) {
	//t.Error("logging on")
	const (
		resultLength = 1
		// timeout when debugging
		//timeout = 0
		// timeout when testing
		timeout = time.Millisecond
	)
	var (
		byte1  = []byte{1}
		byte2  = []byte{2}
		byte12 = append(slices.Clone(byte1), byte2...)
	)

	var (
		result *readResult
		buffer []byte
	)

	// Read
	var reader *FastReader
	// Close CloseWithError Write
	var dataSource *io.PipeWriter
	var reset = func() {
		var pipeReader io.Reader
		pipeReader, dataSource = io.Pipe()
		reader = NewFastReader(pipeReader).(*FastReader)
	}

	// a waiting Reader should get its buffer filled
	reset()
	// reads 1 byte
	result = newReadResult(resultLength)
	go result.read(reader)
	<-result.isReady
	// readerP is now active
	//	- writing 2 bytes should provide the first byte to Reader
	dataSource.Write(byte12)
	if !result.awaitCh(timeout) {
		t.Fatal("Read hung awaiting Write")
	}
	if !slices.Equal(result.p, byte1) {
		t.Errorf("Read readerEvent %v exp %v", result.p, byte1)
	}
	buffer = reader.Buffer()
	if !slices.Equal(buffer, byte2) {
		t.Errorf("buffer bad: %v exp %v", buffer, byte2)
	}
}
