/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"fmt"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// GoErrorDump prints everything about [parl.GoError]
//   - parl.GoError: type: *g0.GoError
//   - err: pnet.InterfaceAddrs netInterface.Addrs route ip+net: invalid network interface at pnet.InterfaceAddrs()-interface.go:30
//   - t: 2023-05-10 15:53:07.885969000-07:00
//   - errContext: GeLocalChan
//   - goroutine: 72_func:g5.(*Netlink).streamReaderThread()-netlink.go:156_cre:g5.(*Netlink).ReaderThread()-netlink.go:79
func GoErrorDump(goError parl.GoError) (s string) {

	// check for nil
	s = fmt.Sprintf("parl.GoError: type: %T", goError)
	if goError == nil {
		return // goError nil returns "parl.GoError: type: <nil>"
	}

	var threadInfo string
	var goroutine = goError.Go()
	if goroutine == nil {
		threadInfo = "nil"
	} else {
		threadInfo = goroutine.ThreadInfo().String()
	}

	// GoError.err t errContext
	s += fmt.Sprintf("\nerr: %s\nt: %s\nerrContext: %s\ngoroutine: %s\nerr trace: %s",
		goError.Error(),
		goError.Time().Format(parl.Rfc3339ns),
		goError.ErrContext(),
		threadInfo,
		perrors.Long(goError.Err()),
	)

	return
}
