/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlfs

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	name := "a.txt"
	dir := t.TempDir()
	wg := sync.WaitGroup{}
	//var err error

	w := NewWatch(dir, 0, nil)

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
	//t.FailNow()

	var errs []error
	wg.Add(1)
	go func() {
		for {
			err, ok := <-w.Errors()
			if !ok {
				break
			}
			errs = append(errs, err)
		}
		wg.Done()
	}()

	var events []*WatchEvent
	wg.Add(1)
	go func() {
		for {
			we, ok := <-w.Events()
			if !ok {
				break
			}
			events = append(events, we)
		}
		wg.Done()
	}()

	filename := filepath.Join(dir, name)

	// /var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestWatch4233947822/001/a.txt
	//t.Log(filename)
	//t.FailNow()

	if osFile, err := os.Create(filename); err != nil {
		t.Logf("os.Create: %s", err.Error())
		t.FailNow()
	} else if err = osFile.Close(); err != nil {
		t.Logf("file Close: %s", err.Error())
		t.FailNow()
	}

	time.Sleep(100 * time.Millisecond)
	w.Shutdown()
	wg.Wait()

	if len(errs) > 0 {
		ss := make([]string, len(errs))
		for i, e := range errs {
			ss[i] = e.Error()
		}
		t.Logf("%d errors: %s", len(errs), strings.Join(ss, ", "))
		t.Fail()
	}

	if len(events) == 0 {
		t.Log("no events")
		t.FailNow()
	}

	if len(events) > 1 {
		t.Logf("More than one event: %d", len(events))
		t.Fail()
	}

	/*
		event:
			2022-03-15 20:11:01.277245 -0700 PDT m=+0.001918293
			uuid: 00000000-0000-0000-0000-000000000000
			base: a.txt
			event: fsnotify.Event{
				Name:"/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestWatch3300123412/001/a.txt",
				Op:0x1
			}
	*/
	t.Logf("event: %s", events[0])
	t.Fail()
}

/*
func TestClose(t *testing.T) {
	ch := make(chan struct{})
	close(ch)
	// close(ch) // panic: close of closed channel
	//ch = nil
	//close(ch) // panic: close of nil channel
}
*/
