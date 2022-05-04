/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/goid"
	"github.com/haraldrudell/parl/pruntime"
	"github.com/haraldrudell/parl/ptime"
)

const (
	gcAddFrames = 1
)

type GoGroup struct {
	waiterr          // Add() IsExit() Wait() Ch() send()
	lock             sync.Mutex
	m                map[parl.GoIndex]*GoerDo // behind lock
	goerIndex                                 // goIndex()
	cancelAndContext                          // Cancel() Context()
}

func NewGoGroup(ctx context.Context) (goCreator parl.GoGroup) {
	return &GoGroup{
		waiterr:          waiterr{wg: &parl.WaitGroup{}},
		m:                map[parl.GoIndex]*GoerDo{},
		cancelAndContext: *newCancelAndContext(ctx),
	}
}

func (gc *GoGroup) Add(conduit parl.ErrorConduit, exitAction parl.ExitAction) (goer parl.Goer) {
	gc.add(1)
	index := gc.goIndex()
	goer = NewGoer(
		conduit,
		exitAction,
		index,
		gc,
		goid.GoID(),
		pruntime.NewCodeLocation(gcAddFrames),
	)
	goerDo := goer.(*GoerDo)
	gc.addGoer(goerDo, index)

	return
}

func (gc *GoGroup) WaitPeriod(duration ...time.Duration) {

	// is GoCreator already done?
	if gc.IsExit() {
		return
	}

	// get duration
	var d time.Duration
	if len(duration) > 0 {
		d = duration[0]
	}
	if d < time.Second {
		d = time.Second
	}

	// channel indicating Wait complete
	waitCh := make(chan struct{})
	go func() {
		parl.Recover(parl.Annotation(), nil, parl.Infallible)
		gc.Wait()
		close(waitCh)
	}()

	// ticker for period status prints
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	parl.Console(gc.String())
	for keepGoing := true; keepGoing; {
		select {
		case <-waitCh:
			keepGoing = false
		case <-ticker.C:
		}

		parl.Console(gc.String())
	}
}

func (gc *GoGroup) String() (s string) {
	timeStamp := ptime.Short()
	goIndex := gc.getGoerList()

	goList := make([]parl.GoIndex, len(goIndex))
	i := 0
	for key := range goIndex {
		goList[i] = key
		i++
	}
	sort.Slice(goList, func(i, j int) bool { return goList[i] < goList[j] })

	adds, dones := gc.wg.Counters()
	s = parl.Sprintf("%s %d(%d)", timeStamp, adds-dones, adds)

	if len(goIndex) == 0 {
		return s + "\x20None"
	}
	if len(goIndex) == 1 {
		return s + "\x20" + (goIndex[goList[0]]).String()
	}

	sList := []string{"\n" + s + ":"}
	for _, index := range goList {
		sList = append(sList, goIndex[index].String())
	}
	return strings.Join(sList, "\n")
}

func (gc *GoGroup) exitAction(err error, exitAction parl.ExitAction, index parl.GoIndex) {
	gc.wg.Done()
	if gc.deleteGoer(index) == 0 {
		gc.waiterr.close()
	}

	if exitAction == parl.ExIgnoreExit ||
		err == nil && exitAction == parl.ExCancelOnFailure {
		return
	}

	gc.Cancel()
}

func (gc *GoGroup) addGoer(goer *GoerDo, index parl.GoIndex) {
	gc.lock.Lock()
	defer gc.lock.Unlock()

	gc.m[index] = goer
}

func (gc *GoGroup) deleteGoer(index parl.GoIndex) (remaining int) {
	gc.lock.Lock()
	defer gc.lock.Unlock()

	delete(gc.m, index)
	return len(gc.m)
}

func (gc *GoGroup) getGoerList() (goIndex map[parl.GoIndex]*GoerDo) {
	gc.lock.Lock()
	defer gc.lock.Unlock()

	goIndex = map[parl.GoIndex]*GoerDo{}
	for index, goer := range gc.m {
		goIndex[index] = goer
	}

	return
}
