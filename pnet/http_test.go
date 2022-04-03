/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"
	"net/http"
	"sync"
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

func TestNewHttp(t *testing.T) {
	TCPaddress := "127.0.0.1:0"
	network := ""
	hp := NewHttp(TCPaddress, network)
	_ = hp
}

func TestHttpListen(t *testing.T) {
	TCPaddress := "127.0.0.1:0"
	network := ""
	protocol := "http://"
	URIPattern := "/" // "/" matches everything

	t.Log("starting server")
	hp := NewHttp(TCPaddress, network)
	hp.HandleFunc(URIPattern, func(w http.ResponseWriter, req *http.Request) {
		t.Logf("server received request from: %s", req.RemoteAddr)
		w.WriteHeader(http.StatusNoContent)
	})
	errCh := hp.Listen()

	// listen for errors
	var errChWg sync.WaitGroup
	errChWg.Add(1)
	go func() {
		defer errChWg.Done()
		t.Log("Reading errCh")
		for {
			err, ok := <-errCh
			if !ok {
				break // errCh closed
			}
			t.Errorf("errCh: %+v", err)
			panic(err)
		}
	}()

	expectServerError := false
	for once := true; once; once = false {

		t.Logf("waiting for server ready")
		isUp, addr := hp.WaitForUp()
		if !isUp {
			t.Error("Server failed to start")
			expectServerError = true
			break
		}
		t.Logf("server listening at %s", addr.String())

		t.Logf("Sending http request")
		requestURL := protocol + addr.String()
		resp, err := http.Get(requestURL)
		if err != nil {
			t.Errorf("http.Get '%s': err: %+v", requestURL, err)
			expectServerError = true
			break
		}
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("http.Get bad status code '%s': %d expected: %d", requestURL, resp.StatusCode, http.StatusNoContent)
			return
		}
		if err := resp.Body.Close(); err != nil {
			t.Errorf("Body.Close: %+v", err)
			return
		}
		t.Logf("response status code: %d", resp.StatusCode)
	}

	// shutdown server
	t.Logf("Shutting down server")
	hp.Shutdown()

	errChWg.Wait() // wait for errCh to close

	if expectServerError {
		panic(perrors.New("Server provided no error")) // errCh should have paniced
	}

	t.Logf("Completed successfully")
}

func NewHttpTypes() {
	var httpServer http.Server // struct
	// Handler value is usually pointer to ServeMux
	_ = httpServer.Handler // field Handler http.Handler
	var _ http.Handler     // interface, methods: ServeHTTP(ResponseWriter, *Request)

	var httpServeMux http.ServeMux // struct, implements http.Handler
	// cascading handlers
	_ = httpServeMux.Handle // func (*http.ServeMux).Handle(pattern string, handler http.Handler)
	// HandleFunc registers the handler function for the given pattern
	_ = httpServeMux.HandleFunc // func (*http.ServeMux).HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	// returns the handler to use for the given request
	_ = httpServeMux.Handler // func (*http.ServeMux).Handler(r *http.Request) (h http.Handler, pattern string)
	// ServeHTTP dispatches the request
	_ = httpServeMux.ServeHTTP // func (*http.ServeMux).ServeHTTP(w http.ResponseWriter, r *http.Request)

	var tcp net.TCPListener // net.Listen listener type: *net.TCPListener
	_ = tcp                 // has func (*net.TCPListener).Close() error
}
