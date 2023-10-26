/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"crypto"
	"crypto/tls"
	"io/fs"
	"net"
	"net/http"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pstrings"
)

type Https struct {
	Http
	Cert    parl.CertificateDer
	Private crypto.Signer
}

func NewHttps(host, network string, certDER parl.CertificateDer, signer crypto.Signer) (hp *Https) {
	if host == "" {
		host = httpsAddr
	}
	return &Https{
		Http:    *NewHttp(host, network),
		Cert:    certDER,
		Private: signer,
	}
}

func (hp *Https) Listen() (errCh <-chan error) {
	errCh = hp.SubListen()
	go hp.listenerThread()
	return
}

const (
	httpsAddr = ":https"
	http11    = "http/1.1"
)

func (hp *Https) listenerThread() {
	defer hp.CloseErr()
	defer parl.Recover(parl.Annotation(), nil, hp.SendErr)
	var didReadyWg bool
	defer func() {
		if !didReadyWg {
			hp.ReadyWg.Done()
		}
	}()

	tlsListener, err := hp.TLS() // *tls.listener — need this or it’s file certificates
	if err != nil {
		return
	}
	hp.ReadyWg.Done()
	didReadyWg = true
	hp.IsListening.Store(true)

	srv := &hp.Server
	if err := srv.Serve(tlsListener); err != nil { // blocking until Shutdown or Close
		if err != http.ErrServerClosed {
			hp.SendErr(perrors.Errorf("srv.ServeTLS: '%w'", err))
			return
		}
	}
}

func (hp *Https) TLS() (tlsListener net.Listener, err error) {
	srv := &hp.Server

	// make ServeTLS execute srv.setupHTTP2_ServeTLS go 1.16.6
	badPath := "%"
	if err = srv.ServeTLS(nil, badPath, badPath); err == nil {
		err = perrors.New("missing error from srv.ServeTLS")
		return
	} else {
		if pathError, ok := err.(*fs.PathError); ok {
			if pathError.Path == badPath {
				err = nil
			}
		}
		if err != nil {
			err = perrors.Errorf("srv.ServeTLS: '%w'", err)
			return
		}
	}

	// get *net.TCPListener from srv.ListenAndServeTLS
	var listener net.Listener
	if listener, err = hp.Listener(); err != nil {
		return
	}

	// create a TLS listener from srv.ServeTLS
	var tlsConfig *tls.Config
	if srv.TLSConfig == nil {
		tlsConfig = &tls.Config{}
	} else {
		tlsConfig = srv.TLSConfig.Clone()
	}
	if !pstrings.StrSliceContains(tlsConfig.NextProtos, http11) {
		tlsConfig.NextProtos = append(tlsConfig.NextProtos, http11)
	}
	tlsConfig.Certificates = make([]tls.Certificate, 1)
	tlsCertificate := &tlsConfig.Certificates[0]
	tlsCertificate.Certificate = append(tlsCertificate.Certificate, hp.Cert) // certificate not from file system
	tlsCertificate.PrivateKey = hp.Private                                   // private key not from file system
	tlsListener = tls.NewListener(listener, tlsConfig)
	return
}
