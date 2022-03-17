/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"strconv"
	"time"

	"github.com/haraldrudell/parl/ev"
	"github.com/haraldrudell/parl/evx"
)

const (
	threeSeconds = "3 seconds passed: type q[enter] to exit"
)

// Timer displays a second counter and help on long-running invocations
func Timer(ctx ev.Callee) {
	var err error
	defer ctx.Result(&err)

	ticker := time.NewTicker(time.Second)
	count := 0
	for {
		select {
		case <-ticker.C:
			count++

			// 3 second alert
			if count == 3 {
				ctx.Send(evx.PrintLine(threeSeconds))
			}

			// second counter
			ctx.Send(evx.StatusText(strconv.Itoa(count)))
			continue
		case <-ctx.Done():
		}
		break
	}
	ticker.Stop()
	ctx.Success()
}
