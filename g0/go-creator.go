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
	"sync/atomic"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/goid"
	"github.com/haraldrudell/parl/pruntime"
	"github.com/haraldrudell/parl/ptime"
)

const (
	gcAddFrames = 1
)

type GoCreatorDo struct {
	sharedChan parl.NBChan[parl.GoError]
	wg         parl.WaitGroup
	lock       sync.Mutex
	m          map[parl.GoIndex]*GoerDo // behind lock
	index      uint64                   // atomic
	ctx        parl.CancelContext
}

func NewGoCreator(ctx context.Context) (goCreator parl.GoCreator) {
	return &GoCreatorDo{
		m:   map[parl.GoIndex]*GoerDo{},
		ctx: parl.NewCancelContext(ctx),
	}
}

func (gc *GoCreatorDo) Add(conduit parl.ErrorConduit, exitAction parl.ExitAction) (goer parl.Goer) {
	gc.wg.Add(1)
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

func (gc *GoCreatorDo) Warnings() (ch <-chan parl.GoError) {
	return gc.sharedChan.Ch()
}

func (gc *GoCreatorDo) Wait() {
	gc.wg.Wait()
}

func (gc *GoCreatorDo) WaitPeriod(duration ...time.Duration) {

	// is GoCreator already done?
	if gc.wg.IsZero() {
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
		gc.wg.Wait()
		close(waitCh)
	}()

	// ticker for period status prints
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	parl.Console(gc.List())
	for keepGoing := true; keepGoing; {
		select {
		case <-waitCh:
			keepGoing = false
		case <-ticker.C:
		}

		parl.Console(gc.List())
	}
}

func (gc *GoCreatorDo) IsExit() (isExit bool) {
	return gc.wg.IsZero()
}

func (gc *GoCreatorDo) List() (s string) {
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

func (gc *GoCreatorDo) errorReceiver(err parl.GoError) {

	// emit the error
	gc.sharedChan.Send(err)
}

func (gc *GoCreatorDo) exitAction(err error, exitAction parl.ExitAction, index parl.GoIndex) {
	gc.wg.Done()
	gc.deleteGoer(index)

	if exitAction == parl.ExIgnoreExit ||
		err == nil && exitAction == parl.ExCancelOnFailure {
		return
	}

	gc.ctx.Cancel()
}

func (gc *GoCreatorDo) addGoer(goer *GoerDo, index parl.GoIndex) {
	gc.lock.Lock()
	defer gc.lock.Unlock()

	gc.m[index] = goer
}

func (gc *GoCreatorDo) deleteGoer(index parl.GoIndex) {
	gc.lock.Lock()
	defer gc.lock.Unlock()

	delete(gc.m, index)
}

func (gc *GoCreatorDo) getGoerList() (goIndex map[parl.GoIndex]*GoerDo) {
	gc.lock.Lock()
	defer gc.lock.Unlock()

	goIndex = map[parl.GoIndex]*GoerDo{}
	for index, goer := range gc.m {
		goIndex[index] = goer
	}

	return
}

func (gc *GoCreatorDo) goIndex() (goIndex parl.GoIndex) {
	return parl.GoIndex(atomic.AddUint64(&gc.index, 1))
}
