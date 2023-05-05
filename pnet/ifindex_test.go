/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net/netip"
	"strconv"
	"strings"
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

func TestIfIndex_Interface(t *testing.T) {
	var ifIndex = NewIfIndex(1) // pick 1 which is the first interface index

	var name string
	var i4, i6 []netip.Prefix
	var err error

	if name, i4, i6, err = ifIndex.InterfaceAddrs(); err != nil {
		t.Fatalf("ifIndex.Interface err: %s", perrors.Short(err))
	}
	var stringifySlice = func(slic []netip.Prefix) (s string) {
		var sL = make([]string, len(slic))
		for i, p := range slic {
			sL[i] = p.String()
			if p.Addr().Is6() {
				sL[i] += " zone: " + strconv.Quote(p.Addr().Zone())
			}
		}
		return strconv.Itoa(len(slic)) + "[" + strings.Join(sL, "\x20") + "]"
	}
	t.Logf("name: %q IPv4: %s IPv6: %s",
		name,
		stringifySlice(i4),
		stringifySlice(i6),
	)
}
