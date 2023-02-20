/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package watchfs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices"
)

type eventStore struct {
	events *pslices.ThreadSafeSlice[*WatchEvent]
	t      *testing.T
}

func (e *eventStore) eventFn(watchEvent *WatchEvent) {
	e.events.Append(watchEvent)
}

func (e *eventStore) errFn(err error) {
	e.t.Errorf("FAIL Watcher err: " + perrors.Long(err))
	e.t.FailNow()
}

func TestNewDirNewFile(t *testing.T) {
	//t.Fail()
	baseFilename := "a.txt"
	filter := WatchOpAll
	directory := t.TempDir()
	baseDir2 := "dir2"
	filePerm := os.FileMode(0600)
	dirPerm := os.FileMode(0700)
	shortTime := 100 * time.Millisecond
	t.Logf("Directory: %s", directory)

	var events = *pslices.NewThreadSafeSlice[*WatchEvent]()
	var store = eventStore{events: &events, t: t}

	var err error
	var list []string
	var op string
	var abs string

	// create watcher watching the temporary directory
	watcher := NewWatcher(directory, filter, nil, store.eventFn, store.errFn)
	defer watcher.Shutdown()

	// check state after create watcher
	list = watcher.List()
	t.Logf("List before dir2: %d[%v]", len(list), strings.Join(list, ",\x20"))
	if len(list) != 1 {
		t.Errorf("FAIL List length before dir2: %d exp %d", len(list), 1)
	}
	if len(list) > 0 {
		if list[0] != directory {
			t.Errorf("FAIL list is\n%q exp:\n%q",
				list[0], directory,
			)
		}
	}

	// create dir2
	t.Logf("mkdir dir2")
	dir2 := filepath.Join(directory, baseDir2)
	if err = os.Mkdir(dir2, dirPerm); err != nil {
		err = perrors.Errorf("os.Mkdir: %w", err)
		t.Error(perrors.Short(err))
		return
	}

	// check state after dir2
	t.Log("Sleep…")
	time.Sleep(shortTime)
	list = watcher.List()
	t.Logf("List after dir2: %d[%v]", len(list), strings.Join(list, ",\x20"))
	var events1 = events.SliceClone()
	t.Logf("event count: %d", len(events1))
	for i, ep := range events1 {
		t.Logf("%d: %s", i+1, ep.String())
	}
	if len(list) != 2 {
		t.Errorf("FAIL List length after dir2: %d exp %d", len(list), 2)
	}
	if len(events1) > 0 {
		op = events1[0].Op
		abs = events1[0].AbsName
	} else {
		op = ""
		abs = ""
	}
	if len(events1) != 1 || op != Create.String() || abs != dir2 {
		t.Errorf("FAIL Event bad after dir2: len: %d—%d op: %s—%s abs:\n%q\n%q",
			len(events1), 1,
			op, Create.String(),
			abs, dir2,
		)
	}
	events.Clear()

	// create file2 in dir2
	t.Log("create file2 in dir2")
	file2 := filepath.Join(dir2, baseFilename)
	if err = os.WriteFile(file2, nil, filePerm); err != nil {
		err = perrors.Errorf("os.WriteFile: %w", err)
		t.Error(perrors.Short(err))
		return
	}

	// check state after file2
	t.Log("Sleep…")
	time.Sleep(shortTime)
	list = watcher.List()
	t.Logf("List after file2: %d[%v]", len(list), strings.Join(list, ",\x20"))
	events1 = events.SliceClone()
	t.Logf("event count: %d", len(events1))
	for i, ep := range events1 {
		t.Logf("%d: %s", i+1, ep.String())
	}
	if len(list) != 2 {
		t.Errorf("FAIL List length after file2: %d exp %d", len(list), 2)
	}
	if len(events1) > 0 {
		op = events1[0].Op
		abs = events1[0].AbsName
	} else {
		op = ""
		abs = ""
	}
	if len(events1) != 1 || op != Create.String() || abs != file2 {
		t.Errorf("FAIL Event bad after file2: len: %d—%d op: %s—%s abs:\n%q\n%q",
			len(events1), 1,
			op, Create.String(),
			abs, file2,
		)
	}
	events.Clear()

	// create file in dir
	t.Log("create file in dir")
	file := filepath.Join(directory, baseFilename)
	if err = os.WriteFile(file, nil, filePerm); err != nil {
		err = perrors.Errorf("os.WriteFile: %w", err)
		t.Error(perrors.Short(err))
		return
	}

	// check state after file
	t.Log("Sleep…")
	time.Sleep(shortTime)
	list = watcher.List()
	t.Logf("List after file: %d[%v]", len(list), strings.Join(list, ",\x20"))
	events1 = events.SliceClone()
	t.Logf("event count: %d", len(events1))
	for i, ep := range events1 {
		t.Logf("%d: %s", i+1, ep.String())
	}
	if len(list) != 2 {
		t.Errorf("FAIL List length after file: %d exp %d", len(list), 2)
	}
	if len(events1) > 0 {
		op = events1[0].Op
		abs = events1[0].AbsName
	} else {
		op = ""
		abs = ""
	}
	if len(events1) != 1 || op != Create.String() || abs != file {
		t.Errorf("FAIL Event bad after file: len: %d—%d op: %s—%s abs:\n%q\n%q",
			len(events1), 1,
			op, Create.String(),
			abs, file,
		)
	}
	events.Clear()

	// remove dir2
	t.Log("remove dir2")
	if err = os.RemoveAll(dir2); err != nil {
		err = perrors.Errorf("os.RemoveAll: %w", err)
		t.Error(perrors.Short(err))
		return
	}

	// check state after remove dir2
	t.Log("Sleep…")
	time.Sleep(shortTime)
	list = watcher.List()
	t.Logf("List after remove dir2: %d[%v]", len(list), strings.Join(list, ",\x20"))
	t.Logf("event count: %d", events.Length())
	for i, ep := range events.SliceClone() {
		t.Logf("%d: %s", i+1, ep.String())
	}
	if len(list) != 1 {
		t.Errorf("FAIL List length after remove dir2: %d exp %d", len(list), 1)
	}
	// 220506 on macOS 12.3.1 github.com/fsnotify/fsnotify v1.5.4
	// there is some race condition producing 2 or 3 events unpredictably
	if events.Length() < 2 {
		t.Errorf("FAIL to few events after remove dir2: %d exp >=%d", len(list), 2)
	}
	if events.Length() > 0 {
		event0, _ := events.Get(0)
		op = event0.Op
		abs = event0.AbsName
	} else {
		op = ""
		abs = ""
	}
	if op != Remove.String() || abs != file2 {
		t.Errorf("FAIL Event bad after remove dir2: op: %s—%s abs:\n%q\n%q",
			op, Create.String(),
			abs, file2,
		)
	}
	events.Clear()
}

func TestNewFile(t *testing.T) {
	//t.Fail()
	baseFilename := "a.txt"
	filter := WatchOpAll
	directory := t.TempDir()
	filePerm := os.FileMode(0600)
	shortTime := 100 * time.Millisecond
	t.Logf("Directory: %s", directory)

	var events = pslices.NewThreadSafeSlice[*WatchEvent]()
	var store = eventStore{events: events, t: t}
	var err error
	var list []string
	var op string
	var abs string

	// create watcher
	watcher := NewWatcher(directory, filter, nil, store.eventFn, store.errFn)
	defer watcher.Shutdown()

	// create a file
	filename := filepath.Join(directory, baseFilename)
	if err = os.WriteFile(filename, nil, filePerm); err != nil {
		err = perrors.Errorf("os.WriteFile: %w", err)
		t.Error(perrors.Short(err))
		return
	}

	// check after file
	t.Log("Sleep…")
	time.Sleep(shortTime)
	list = watcher.List()
	t.Logf("List after file: %d[%v]", len(list), strings.Join(list, ",\x20"))
	t.Logf("event count: %d", events.Length())
	var rangeCh = pslices.NewRangeCh(events)
	for ep := range rangeCh.Ch() {

		// 220505_23:24:59-07 uuid: 098c CREATE abs: /var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestNewFile2950680646/001/a.txt
		t.Log(ep.String())
	}

	if events.Length() == 0 {
		t.Error("no events")
	} else if events.Length() > 1 {
		t.Error("More than one event")
	}
	if event0, hasValue := events.Get(0); hasValue {
		op = event0.Op
		abs = event0.AbsName
	} else {
		op = ""
		abs = ""
	}
	if events.Length() != 1 || op != Create.String() || abs != filename {
		t.Errorf("FAIL Event bad after file: len: %d—%d op: %s—%s abs:\n%q\n%q",
			events.Length(), 1,
			op, Create.String(),
			abs, filename,
		)
	}
}

// allBits: 31 opSTring: CREATE|REMOVE|WRITE|RENAME|CHMOD|xe0
//t.Logf("allBits: %d opSTring: %s", allBits, Op(255).String())
// t.FailNow()

/*
	watcher: &{
		Now:2022-03-15 17:55:20.624849 -0700 PDT m=+0.001181376
		ID:1
		Err:<nil>
		dir0:/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestWatch4194545761/001
		cleanDir:/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestWatch4194545761/001
		abs:/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestWatch4194545761/001
		events:0x140000242a0
		errChan:0x14000024300
		Watcher:{Events:0x140000243c0 Errors:0x14000024420 done:0x14000024480 kq:7 mu:{state:0 sema:0} watches:map[] externalWatches:map[] dirFlags:map[] paths:map[] fileExists:map[] isClosed:false}
		filter:0
		shutCh:0x14000024360
		ctx:0x140000180a8
		shutdownLock:{done:0 m:{state:0 sema:0}}
		isShutdown:{value:0}
	}
*/
//t.Logf("watcher: %+v", w)
