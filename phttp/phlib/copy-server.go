/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phlib

import "net/http"

// CopyServer does shallow copy from serverFrom to serverTo
func CopyServer(serverFrom, serverTo *http.Server) {
	*serverTo = http.Server{
		Addr:                         serverFrom.Addr,
		Handler:                      serverFrom.Handler,
		DisableGeneralOptionsHandler: serverFrom.DisableGeneralOptionsHandler,
		TLSConfig:                    serverFrom.TLSConfig,
		ReadTimeout:                  serverFrom.ReadTimeout,
		ReadHeaderTimeout:            serverFrom.ReadHeaderTimeout,
		WriteTimeout:                 serverFrom.WriteTimeout,
		IdleTimeout:                  serverFrom.IdleTimeout,
		MaxHeaderBytes:               serverFrom.MaxHeaderBytes,
		TLSNextProto:                 serverFrom.TLSNextProto,
		ConnState:                    serverFrom.ConnState,
		ErrorLog:                     serverFrom.ErrorLog,
		BaseContext:                  serverFrom.BaseContext,
		ConnContext:                  serverFrom.ConnContext,
		HTTP2:                        serverFrom.HTTP2,
		Protocols:                    serverFrom.Protocols,
	}
}
