/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"
	"strings"
	"sync"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

type WriteCloserToChanLine struct {
	lock        sync.Mutex
	s           string
	ch          parl.NBChan[string]
	withNewline bool
}

func NewWriteCloserToChanLine(withNewline ...bool) (writeCloser io.WriteCloser) {
	var withNewline0 bool
	if len(withNewline) > 0 {
		withNewline0 = withNewline[0]
	}
	return &WriteCloserToChanLine{withNewline: withNewline0}
}

func (wc *WriteCloserToChanLine) Write(p []byte) (n int, err error) {

	// check for closed write stream
	if wc.ch.DidClose() {
		err = perrors.ErrorfPF(ErrFileAlreadyClosed.Error())
		return
	}

	wc.lock.Lock()
	defer wc.lock.Unlock()

	// write data to string
	s := wc.s + string(p)
	n = len(p)

	// write buffer line-by-line to channel
	for {
		index := strings.Index(s, "\n")
		if index == -1 {
			break // no more newlines
		}
		var i int
		if wc.withNewline {
			i = index + 1
		} else {
			i = index
		}
		wc.ch.Send(s[:i])
		s = s[index+1:]
	}
	wc.s = s // store remaining string characters

	return
}

func (wc *WriteCloserToChanLine) Close() (err error) {
	wc.lock.Lock()
	defer wc.lock.Unlock()

	if wc.s != "" {
		wc.ch.Send(wc.s)
	}

	wc.ch.Close()
	return
}

func (wc *WriteCloserToChanLine) Ch() (readCh <-chan string) {
	return wc.ch.Ch()
}
