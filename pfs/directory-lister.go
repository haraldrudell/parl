/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

type DirectoryLister struct {
	Path       string // may begin with ., may be .
	Abs        string
	Results    chan *EntryResult
	sdChan     chan struct{}
	isShutdown parl.AtomicBool
	output     chan *EntryResult
	pinger     chan struct{}
	readEnd    bool
	dirFile    *os.File
	no         int32
	name       string
}

const (
	chanSizeDefault = 5
)

var instanceNo int32

func NewDirStream(path string, chanSize int) (dir *DirectoryLister) {
	if chanSize < 1 {
		chanSize = chanSizeDefault
	}
	cleaned := filepath.Clean(path)
	abs, err := filepath.Abs(cleaned)
	if err != nil {
		panic(perrors.Errorf("filepath.Abs(%q) '%w'", path, err))
	}
	no := atomic.AddInt32(&instanceNo, 1)
	dir = &DirectoryLister{Path: cleaned, Abs: abs, Results: make(chan *EntryResult),
		sdChan: make(chan struct{}),
		output: make(chan *EntryResult, chanSize),
		pinger: make(chan struct{}, 1),
		no:     no, name: fmt.Sprintf("dir%d", no)}
	go dir.reader()
	go dir.forwarder()
	return
}

func (dir *DirectoryLister) Shutdown() {
	if dir.isShutdown.Set() {
		close(dir.sdChan)
	}
}

func (dir *DirectoryLister) reader() {
	name := dir.name + "reader"
	output := dir.output
	defer close(output)
	defer parl.Recover(name, nil, func(err error) { output <- GetErrorResult(err) })

	pinger := dir.pinger
	for {
		if n := cap(output) - len(output); n > 0 {
			dir.read(n)
			if dir.readEnd {
				break
			}
		}
		select {
		case <-dir.sdChan:
		case _, ok := <-pinger:
			if ok {
				continue
			}
		}
		break
	}
}

func (dir *DirectoryLister) read(n int) {

	// os.Open of directory
	d := dir.dirFile
	var err error
	if d == nil {
		if d, err = os.Open(dir.Path); err != nil {
			dir.flagError(err)
			return
		}
		dir.dirFile = d
	}

	// read directory
	var entries []fs.DirEntry // interface
	if entries, err = d.ReadDir(n); err != nil {
		if err != io.EOF {
			dir.flagError(err)
		}
		dir.close()
		return
	}
	for _, entry := range entries {
		var fileInfo fs.FileInfo // interface
		fileInfo, err = entry.Info()
		if err != nil {
			dir.flagError(err)
			dir.close()
			return
		}
		var stat *syscall.Stat_t
		if s, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
			stat = s
		}
		dir.output <- &EntryResult{DLEntry: GetEntry(dir.Path, dir.Abs, entry, fileInfo, stat)}
	}
}

func (dir *DirectoryLister) forwarder() {
	name := dir.name + "forwarder"
	defer close(dir.Results)
	defer parl.Recover(name, nil, func(err error) { dir.Results <- GetErrorResult(err) })
	pinger := dir.pinger
	defer close(pinger)

	in := dir.output
	for {
		result, ok := <-in
		if !ok {
			break
		}
		if len(pinger) == 0 {
			pinger <- struct{}{}
		}
		dir.Results <- result
	}
}

func (dir *DirectoryLister) close() {
	if d := dir.dirFile; d != nil {
		dir.dirFile = nil
		if err := d.Close(); err != nil {
			dir.flagError(err)
		}
	}
	dir.end()
}

func (dir *DirectoryLister) flagError(err error) {
	parl.Debug("flagError: %v\n", err)
	dir.Results <- GetErrorResult(err)
	dir.end()
}

func (dir *DirectoryLister) end() {
	if !dir.readEnd {
		dir.readEnd = true
	}
}
