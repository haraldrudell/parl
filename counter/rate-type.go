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
	{ValueV: RateType(parl.ValueRate), Name: "value rate"},
	{ValueV: RateType(parl.ValueMaxRate), Name: "value max rate"},
	{ValueV: RateType(parl.RunningRate), Name: "running inc rate"},
	{ValueV: RateType(parl.RunningMaxRate), Name: "runninc max inc rate"},
	{ValueV: RateType(parl.RunningMaxDecRate), Name: "running max dec rate"},
}))
