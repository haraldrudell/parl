/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"
	"net/http"
)

type Tcp struct {
	net.Addr // interface
	http.ServeMux
	http.Server
}

/*
// NewHttp creates http server host is host:port, default ":http"
func NewTcp(host string) (hp *Http) {
	h := Http{ServeMux: *http.NewServeMux()}
	h.Server.Addr = host
	h.Server.Handler = &h.ServeMux // interface
	return &h
}

const (
	schemeHttp  = "http"
	schemeHttps = "https"
)

func (hp *Http) Run() (errCh <-chan error) {
	errChan := make(chan error)
	listener, err := hp.listen()
	if err != nil {
		panic(err)
	}
	go hp.run(errChan, listener)
	return errChan
}

func (hp *Http) RunTLS() (errCh <-chan error) {
	errChan := make(chan error)
	listener, err := hp.listenTLS()
	if err != nil {
		panic(err)
	}
	go hp.run(errChan, listener)
	return errChan
}

func (hp *Http) run(errCh chan<- error, listener net.Listener) {
	defer close(errCh)
	defer parl.Recover("", func(e error) { errCh <- e })
	if err := hp.Server.Serve(listener); err != nil { // blocking until Shutdown or Close
		if err != http.ErrServerClosed {
			errCh <- err
			return
		}
	}
}

func (hp *Http) listen() (listener net.Listener, err error) {
	srv := hp.Server
	addr := srv.Addr
	if addr == "" {
		addr = ":http"
	}
	hp.mutex.Lock()
	defer hp.mutex.Unlock()
	if hp.Addr != nil {
		err = error116.New("Already listening")
		return
	}
	listener, err = net.Listen("tcp", addr)
	hp.Addr = listener.Addr()
	return
}

func (hp *Http) listenTLS() (listener net.Listener, err error) {
	srv := hp.Server
	addr := srv.Addr
	if addr == "" {
		addr = ":https"
	}
	hp.mutex.Lock()
	defer hp.mutex.Unlock()
	if hp.Addr != nil {
		err = error116.New("Already listening")
		return
	}
	listener, err = net.Listen("tcp", addr)
	hp.Addr = listener.Addr()
	return
}

func (hp *Http) Shutdown() (err error) {
	return hp.Server.Shutdown(context.Background())
}
&*/
