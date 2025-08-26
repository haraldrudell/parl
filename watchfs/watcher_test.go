/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package watchfs

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices"
	"github.com/haraldrudell/parl/pstrings"
)

func TestWatcher(t *testing.T) {

	// Shutdown() Watch()
	//var watcher *Watcher = NewWatcher(WatchOpAll, NoIgnores)
}

// test watching single directory
//   - Phase 1: watch single directory
//   - Phase 2: add subdirectory
//   - Phase 3: add file in subdirectory
//   - Phase 4: add file in directory
//   - Phase 5: rmdir subdirectory: removes two entries
func TestWatcherDirectory(t *testing.T) {
	//t.Fail()
	// Phase1: create initially watched directory1
	var directory1 = t.TempDir()
	var directory1AbsEval = func() (absEval string) {
		var err error
		if absEval, err = filepath.Abs(directory1); err != nil {
			panic(err)
		} else if absEval, err = filepath.EvalSymlinks(absEval); err != nil {
			panic(err)
		}
		return
	}()
	// Phase1: expected result of List1
	var list1exp = []string{directory1AbsEval}
	// Phase2 create “directory/dir2”
	var dir2Name = "dir2"
	// Phase2 create “directory/dir2”
	var dir2 = filepath.Join(directory1AbsEval, dir2Name)
	var dir2Perm = os.FileMode(0700)
	// Phase2: expected result of List2
	var list2exp = []string{directory1AbsEval, dir2}
	slices.Sort(list2exp)
	var events2exp = []simpleEvent{{
		AbsName: dir2,
		Op:      Create.String(),
	}}
	// Phase3: create dir2/a.txt
	var base3 = "file3.txt"
	// Phase3: dir2/a.txt
	var file3 = filepath.Join(dir2, base3)
	var file3Perm = os.FileMode(0600)
	var noData []byte
	var events3exp = []simpleEvent{{
		AbsName: file3,
		Op:      Create.String(),
	}}
	// Phase4: dir/a.txt
	var file4 = filepath.Join(directory1AbsEval, "file4.txt")
	var events4exp = []simpleEvent{{
		AbsName: file4,
		Op:      Create.String(),
	}}
	// Phase 5 events: file removed prior to directory
	var events5RemoveDir2 = simpleEvent{
		AbsName: dir2,
		Op:      Remove.String(),
	}
	var events5exp = []simpleEvent{{
		AbsName: file3,
		Op:      Remove.String(),
	},
		events5RemoveDir2,
	}
	// watcher filter
	var filter = WatchOpAll
	var shortTime = time.Millisecond
	var simpleAct []simpleEvent

	// events is a thread-safe slice updfated in real-time
	// with watcher events
	var events = *pslices.NewThreadSafeSlice[*WatchEvent]()
	// store has eventFunc and ErrFn
	var store = newEventStore(&events, t)

	// Watch() Add() List() Shutdown()
	var watcher *Watcher
	var err error
	var listAct []string
	var eventsAct []*WatchEvent
	// no ignore name-filter
	var ignores *regexp.Regexp

	// create watcher watching the temporary directory
	watcher = NewWatcher(filter, ignores, store.eventFunc, store)
	err = watcher.Watch(directory1)
	if err != nil {
		t.Errorf("Watch err: %s", perrors.Short(err))
	}
	defer watcher.Shutdown()

	// Phase1: List should return directory
	t.Logf("Phase1 directory: %s", directory1)
	listAct = watcher.List()
	// List before dir2: 1[/private/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestNewDirNewFile1558088706/001]
	t.Logf("List1 directory only: %d[%v]", len(listAct), strings.Join(listAct, ",\x20"))
	if !slices.Equal(listAct, list1exp) {
		t.Errorf("List1 BAD\n%v exp\n%v", listAct, list1exp)
	}

	// Phase 2: create dir2
	t.Logf("Phase2: mkdir dir2")
	if err = os.Mkdir(dir2, dir2Perm); err != nil {
		panic(err)
	}
	// sleep allows for the watcher to
	//	- receive the CREATE event
	//	- add a new watcher for the created directory
	t.Log("Sleep…")
	time.Sleep(shortTime)

	// Phase2: List should return directory, dir2
	listAct = watcher.List()
	slices.Sort(listAct)
	// List director, dir2: 2[
	// "/private/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestNewDirNewFile1936666749/001"
	// "/private/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestNewDirNewFile1936666749/001/dir2"]
	t.Logf("List2: directory, dir2: %d[%s]", len(listAct), pstrings.QuoteList(listAct))
	if !slices.Equal(listAct, list2exp) {
		t.Errorf("List2 BAD\n%v exp\n%v", listAct, list2exp)
	}

	// Phase2: events should be single CREATE
	eventsAct = events.SliceClone()
	// events2 count: 1
	t.Logf("events2 count: %d", len(eventsAct))
	for i, ep := range eventsAct {
		// 1: 231215_06:52:13-08 uuid: df75
		// CREATE
		// /private/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestNewDirNewFile1942365152/001/dir2
		t.Logf("%d: %s", i+1, ep.String())
	}
	simpleAct = simpleSlice(eventsAct)
	if !slices.Equal(simpleAct, events2exp) {
		t.Errorf("Events2 BAD after dir2: %v exp %v", simpleAct, events2exp)
	}
	events.Clear()

	// Phase3: create file2 in dir2
	t.Log("Phase3: create file2 in dir2")
	if err = os.WriteFile(file3, noData, file3Perm); err != nil {
		panic(err)
	}
	t.Log("Sleep…")
	time.Sleep(shortTime)

	// Phase3: List should still return dir, dir2
	listAct = watcher.List()
	slices.Sort(listAct)
	// List3 after file2: 2[
	// /private/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestWatcherDirectory4278016608/001
	// /private/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestWatcherDirectory4278016608/001/dir2]
	t.Logf("List3 after file2: %d[%v]", len(listAct), strings.Join(listAct, "\x20"))
	if !slices.Equal(listAct, list2exp) {
		t.Errorf("List3 BAD\n%v exp\n%v", listAct, list2exp)
	}

	// Phase 3 events should return single Create dir2/a.txt
	eventsAct = events.SliceClone()
	t.Logf("events3 count: %d", len(eventsAct))
	for i, ep := range eventsAct {
		t.Logf("%d: %s", i+1, ep.String())
	}
	simpleAct = simpleSlice(eventsAct)
	if !slices.Equal(simpleAct, events3exp) {
		t.Errorf("Events3 BAD after dir2/a.txt: %v exp %v", simpleAct, events3exp)
	}
	events.Clear()

	// Phase4: create file in dir
	t.Log("Phase 4: create file in dir")
	if err = os.WriteFile(file4, noData, file3Perm); err != nil {
		panic(err)
	}
	t.Log("Sleep…")
	time.Sleep(shortTime)

	// Phase4: List should return directiry1, dir2
	listAct = watcher.List()
	slices.Sort(listAct)
	t.Logf("List4 after file4: %d[%v]", len(listAct), strings.Join(listAct, ",\x20"))
	if !slices.Equal(listAct, list2exp) {
		t.Errorf("List3 BAD\n%v exp\n%v", listAct, list2exp)
	}

	// Phase 4: events should be 1 Create file4
	eventsAct = events.SliceClone()
	t.Logf("event4 count: %d", len(eventsAct))
	for i, ep := range eventsAct {
		t.Logf("%d: %s", i+1, ep.String())
	}
	simpleAct = simpleSlice(eventsAct)
	if !slices.Equal(simpleAct, events4exp) {
		t.Errorf("Events4 BAD after dir/file4: %v exp %v", simpleAct, events4exp)
	}
	events.Clear()

	// Phase5: remove dir2
	t.Log("Phase 5: remove dir2")
	if err = os.RemoveAll(dir2); err != nil {
		panic(err)
	}
	t.Log("Sleep…")
	time.Sleep(shortTime)

	// Phase 5: List should return directory1 only
	listAct = watcher.List()
	// List5 after remove dir2: 1[
	// /private/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestWatcherDirectory3111656516/001]
	t.Logf("List5 after remove dir2: %d[%v]", len(listAct), strings.Join(listAct, ",\x20"))
	if !slices.Equal(listAct, list1exp) {
		t.Errorf("List5 BAD\n%v exp\n%v", listAct, list1exp)
	}

	// Phase 5: events: should be Remove file3, Remove dir2
	eventsAct = events.SliceClone()
	// event5 count: 2
	t.Logf("event5 count: %d", len(eventsAct))
	for i, ep := range eventsAct {
		t.Logf("%d: %s", i+1, ep.String())
	}
	simpleAct = simpleSlice(eventsAct)
	// 220506 on macOS 12.3.1 github.com/fsnotify/fsnotify v1.5.4
	//	- there is some race condition producing 2 or 3 events unpredictably
	//	- duplicate Remove events for dir2
	var didRemoveDir2 bool
	for i := 0; i < len(simpleAct); i++ {
		if simpleAct[i] != events5RemoveDir2 {
			continue // not Remove dir2
		}
		if !didRemoveDir2 {
			didRemoveDir2 = true
			continue // 1 is allowed
		}
		t.Logf("FIX DUPLICATE EVENTS macOS fsnotify v1.5.4+")
		simpleAct = append(simpleAct[:i], simpleAct[i+1:]...)
		i--
	}
	if !slices.Equal(simpleAct, events5exp) {
		t.Errorf("Events5 BAD after rmdir dir2: %v exp %v", simpleAct, events5exp)
	}
	events.Clear()
}

// tests that initial scan finds subdirectory
func TestWatcherDirScan(t *testing.T) {
	//t.Fail()
	// Phase 1: directory1 is temporary directory
	var directory1 = t.TempDir()
	// absolute symlink-evaled temporary directory
	var directory1AbsEval = func() (absEval string) {
		var err error
		if absEval, err = filepath.Abs(directory1); err != nil {
			panic(err)
		} else if absEval, err = filepath.EvalSymlinks(absEval); err != nil {
			panic(err)
		}
		return
	}()
	var dir2 = filepath.Join(directory1AbsEval, "dir2")
	var dir2Perm = os.FileMode(0700)
	var list1exp = []string{directory1AbsEval, dir2}
	slices.Sort(list1exp)
	// watcher filter
	var filter = WatchOpAll

	// events is a thread-safe slice updfated in real-time
	// with watcher events
	var events = *pslices.NewThreadSafeSlice[*WatchEvent]()
	// store has eventFunc and ErrFn
	var store = newEventStore(&events, t)

	// Watch() Add() List() Shutdown()
	var watcher *Watcher
	var err error
	var listAct []string
	// no ignore name-filter
	var ignores *regexp.Regexp

	// Phase 1: watch directory with subdirectory
	t.Log("Phase 1: watch directory with subdirectory dir2")
	if err = os.Mkdir(dir2, dir2Perm); err != nil {
		panic(err)
	}
	watcher = NewWatcher(filter, ignores, store.eventFunc, store)
	err = watcher.Watch(directory1)
	if err != nil {
		t.Errorf("Watch err: %s", perrors.Short(err))
	}
	defer watcher.Shutdown()

	// Phase1: List should return directory, dir2
	listAct = watcher.List()
	slices.Sort(listAct)
	// List1 directory dir2: 2[
	// /private/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestWatcherDirScan3224474477/001,
	// /private/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestWatcherDirScan3224474477/001/dir2]
	t.Logf("List1 directory dir2: %d[%v]", len(listAct), strings.Join(listAct, ",\x20"))
	if !slices.Equal(listAct, list1exp) {
		t.Errorf("List1 BAD\n%v exp\n%v", listAct, list1exp)
	}
}

// tests watching a file
//   - Phase 1: file is watched
//   - Phase 2: watched file is removed
func TestWatcherFile(t *testing.T) {
	//t.Fail()
	// Phase 1: directory1 is temporary directory
	var directory1 = t.TempDir()
	// absolute symlink-evaled temporary directory
	var directory1AbsEval = func() (absEval string) {
		var err error
		if absEval, err = filepath.Abs(directory1); err != nil {
			panic(err)
		} else if absEval, err = filepath.EvalSymlinks(absEval); err != nil {
			panic(err)
		}
		return
	}()
	var file1 = filepath.Join(directory1AbsEval, "file1.txt")
	var file1Perm = os.FileMode(0600)
	// Phase1: expected result of List1
	var list1exp = []string{file1}
	// Phase 2: remove directory1
	var list2exp = []string{}
	var events2exp = []simpleEvent{{
		AbsName: file1,
		Op:      Remove.String(),
	}}
	var filter = WatchOpAll
	var noData []byte
	var shortTime = time.Millisecond
	var simpleAct []simpleEvent

	var watcher *Watcher
	var events = pslices.NewThreadSafeSlice[*WatchEvent]()
	var store = eventStore{events: events, t: t}
	var err error
	var listAct []string
	var eventsAct []*WatchEvent
	// no ignore name-filter
	var ignores *regexp.Regexp

	// Phase 1: create file
	t.Logf("Phase 1: file1 in directory: %s", directory1)
	if err = os.WriteFile(file1, noData, file1Perm); err != nil {
		panic(err)
	}
	watcher = NewWatcher(filter, ignores, store.eventFunc, &store)
	err = watcher.Watch(file1)
	if err != nil {
		t.Errorf("Watch err: %s", perrors.Short(err))
	}
	defer watcher.Shutdown()

	// Phase 1: List should return directory1
	listAct = watcher.List()
	// List1 after WriteFile: 1[
	// /private/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestWatcherFile1102180354/001/file1.txt]
	t.Logf("List1 after WriteFile: %d[%v]", len(listAct), strings.Join(listAct, ",\x20"))
	if !slices.Equal(listAct, list1exp) {
		t.Errorf("List1 BAD\n%v exp\n%v", listAct, list1exp)
	}

	// Phase 2: remove parent directory
	if err = os.RemoveAll(directory1); err != nil {
		panic(err)
	}
	t.Log("Sleep…")
	time.Sleep(shortTime)

	// Phase 2: list
	listAct = watcher.List()
	// List1 after WriteFile: 1[
	// /private/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestWatcherFile864689061/001]
	t.Logf("List2 after remove: %d[%v]", len(listAct), strings.Join(listAct, ",\x20"))
	if !slices.Equal(listAct, list2exp) {
		t.Errorf("List2 BAD\n%v exp\n%v", listAct, list2exp)
	}

	// Phase 2: events
	eventsAct = events.SliceClone()
	// events2 count: 1
	t.Logf("events2 count: %d", len(eventsAct))
	var rangeCh = pslices.NewRangeCh(events)
	for ep := range rangeCh.Ch() {
		// 220505_23:24:59-07 uuid: 098c CREATE abs: /var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestNewFile2950680646/001/a.txt
		t.Log(ep.String())
	}
	simpleAct = simpleSlice(eventsAct)
	if !slices.Equal(simpleAct, events2exp) {
		t.Errorf("Events2 BAD after removeAll directory1: %v exp %v", simpleAct, events2exp)
	}
	events.Clear()
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

// eventStore appends watcher events to events slice in real-time
type eventStore struct {
	events *pslices.ThreadSafeSlice[*WatchEvent]
	t      *testing.T
}

// eventStore appends watcher events to events slice in real-time
func newEventStore(events *pslices.ThreadSafeSlice[*WatchEvent], t *testing.T) (store *eventStore) {
	return &eventStore{events: events, t: t}
}

// eventFunc is the event receiver
func (e *eventStore) eventFunc(watchEvent *WatchEvent) {
	e.events.Append(watchEvent)
}

// errFn is the error receiver
func (e *eventStore) AddError(err error) {
	e.t.Fatalf("FAIL Watcher err: %s", perrors.Long(err))
}

var _ WatchEvent

// simpleEvent is a simplified WatchEvent
//   - allows for slices.Equal
type simpleEvent struct {
	BaseName string
	AbsName  string
	Op       string
}

// simpleSlice returns a predictable event list
//   - allows for slices.Equal
func simpleSlice(slice []*WatchEvent) (simple []simpleEvent) {
	simple = make([]simpleEvent, len(slice))
	for i, ep := range slice {
		simple[i] = simpleEvent{
			AbsName: ep.AbsName,
			Op:      ep.Op,
		}
	}
	return
}
