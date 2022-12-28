/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pslices"
	"github.com/haraldrudell/parl/set"
)

type RateType parl.RateType

func (rt RateType) String() (s string) {
	return rateSet.StringT(rt)
}

var rateSet = set.NewSet(pslices.ConvertSliceToInterface[
	set.Element[RateType],
	parl.Element[RateType],
]([]set.Element[RateType]{
	{RateType(parl.ValueRate), "value rate"},
	{RateType(parl.ValueMaxRate), "value max rate"},
	{RateType(parl.RunningRate), "running inc rate"},
	{RateType(parl.RunningMaxRate), "runninc max inc rate"},
	{RateType(parl.RunningMaxDecRate), "running max dec rate"},
}))
