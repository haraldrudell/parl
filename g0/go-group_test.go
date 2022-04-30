/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"sync"
	"testing"

	"github.com/haraldrudell/parl"
)

func TestGoCreatorDo_Add(t *testing.T) {
	ctx := context.Background()

	gc := NewGoGroup(ctx)
	goer := gc.Add(parl.EcSharedChan, parl.ExCancelOnExit)

	var waitForRegister sync.WaitGroup
	waitForRegister.Add(1)
	go func(g0 parl.Go) {
		g0.Register() // give thread information to goer
		waitForRegister.Done()
		// TODO
	}(goer.Go())

	waitForRegister.Wait()
	gd := goer.(*GoerDo)
	var otherIDs []parl.ThreadID
	for key := range gd.otherIDs {
		otherIDs = append(otherIDs, key)
	}
	t.Logf("parentID: %s threadID: %s otherIDs: %v "+
		"add:\n%s create:\n%s func:\n%s",
		gd.parentID, gd.threadID, otherIDs,
		gd.addLocation.Short(),
		gd.createLocation.Short(),
		gd.funcLocation.Short(),
	)

	//t.Fail()
}
