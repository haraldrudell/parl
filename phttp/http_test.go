/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phttp

import (
	"fmt"
	"net/http"
	"net/netip"
	"sync/atomic"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pnet"
)

func TestNewHttp(t *testing.T) {
	var nearSocketInvalid0 = netip.AddrPort{}
	var expAddr1 = ":http"
	var nearSocket0 = netip.MustParseAddrPort("1.2.3.4:0")
	var expAddr2 = "1.2.3.4:http"
	var nearSocket6 = netip.MustParseAddrPort("[::1]:1024")
	var expAddr3 = "[::1]:1024"

	// HandleFunc() Listen() SendErr() Shutdown() WaitForUp()
	var pnetHttp *Http

	// nearSocket invalid 0
	pnetHttp = NewHttp(nearSocketInvalid0, pnet.NetworkDefault)
	// addr: ":http" network: tcp
	t.Logf("addr: %q network: %s", pnetHttp.Server.Addr, pnetHttp.Network)
	if pnetHttp.Network != pnet.NetworkTCP {
		t.Errorf("New1 bad network %s exp %s", pnetHttp.Network, pnet.NetworkTCP)
	}
	if pnetHttp.Server.Addr != expAddr1 {
		t.Errorf("New1 bad addr %q exp %q", pnetHttp.Server.Addr, expAddr1)
	}

	// port 0
	pnetHttp = NewHttp(nearSocket0, pnet.NetworkTCP)
	if pnetHttp.Network != pnet.NetworkTCP {
		t.Errorf("New bad network %s exp %s", pnetHttp.Network, pnet.NetworkTCP)
	}
	if pnetHttp.Server.Addr != expAddr2 {
		t.Errorf("New1 bad addr %q exp %q", pnetHttp.Server.Addr, expAddr1)
	}

	// IPv6
	pnetHttp = NewHttp(nearSocket6, pnet.NetworkTCP)
	if pnetHttp.Network != pnet.NetworkTCP {
		t.Errorf("New bad network %s exp %s", pnetHttp.Network, pnet.NetworkTCP)
	}
	if pnetHttp.Server.Addr != expAddr3 {
		t.Errorf("New1 bad addr %q exp %q", pnetHttp.Server.Addr, expAddr1)
	}
}

func TestHttp(t *testing.T) {
	var nearSocket netip.AddrPort
	var network pnet.Network
	// "/" matches everything
	var URIPattern = "/"

	var protocol = "http://"
	var handler *sHandler
	var near, respS string
	var resp *http.Response
	var err error
	var statusCode int
	var goResult = parl.NewGoResult()

	// HandleFunc() Listen() SendErr() Shutdown() WaitForUp()
	var pnetHttp *Http = NewHttp(nearSocket, network)
	handler = newShandler()
	pnetHttp.HandleFunc(URIPattern, handler.Handle)
	defer pnetHttp.Shutdown()

	t.Log("invoking Listen")
	go errChListener(pnetHttp.Listen(), goResult)

	t.Log("waiting for ListenAwaitable")
	<-pnetHttp.ListenAwaitable.Ch()
	if !pnetHttp.Near.IsValid() {
		t.Fatalf("FATAL: pnetHttp.Near invalid")
	}
	near = pnetHttp.Near.String()
	t.Logf("Near: %s", near)

	t.Log("issuing http.GET")
	resp, err = http.Get(protocol + near)
	if resp != nil {
		statusCode = resp.StatusCode
		respS = fmt.Sprintf("status code: %d", statusCode)
	} else {
		respS = "resp nil"
	}

	t.Logf("%s err: %s", respS, perrors.Short(err))
	if err != nil {
		t.Errorf("http.Get err %s", perrors.Short(err))
	}
	// status code should be 204
	if resp != nil {
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("http.Get status code %d exp %d", resp.StatusCode, http.StatusNoContent)
		}
		if e := resp.Body.Close(); e != nil {
			panic(e)
		}
	}

	// handle count should be 1
	if c := int(handler.Rqs.Load()); c != 1 {
		t.Errorf("bad handle count: %d exp 1", c)
	}

	t.Logf("Shutting down server")
	pnetHttp.Shutdown()

	// wait for error reader to exit
	goResult.ReceiveError(parl.NoErrp)

	if !pnetHttp.EndListenAwaitable.IsClosed() {
		t.Error("EndListenAwaitable not closed")
	}
}

// errChListener is goroutine consuming http error channel
func errChListener(errorSource parl.ErrorSource, g parl.GoResult) {
	var err error
	defer g.SendError(&err)

	var endCh = errorSource.EndCh()
	for {
		select {
		case <-errorSource.WaitCh():
			var err, hasValue = errorSource.Error()
			if !hasValue {
				continue
			}
			parl.Log(perrors.Long(err))
			panic(err)
		case <-endCh:
			return
		}
	}
}

// sHandler counts incoming requests
type sHandler struct{ Rqs atomic.Uint64 }

// newShandler returns a request counter
func newShandler() (s *sHandler) { return &sHandler{} }

// Handle is the http-server handler function
func (s *sHandler) Handle(w http.ResponseWriter, r *http.Request) {
	s.Rqs.Add(1)
	w.WriteHeader(http.StatusNoContent)
}
