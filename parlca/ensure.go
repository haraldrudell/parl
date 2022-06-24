/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/x509"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/haraldrudell/parl/perrors"
)

func EnsureTemplate(cert *x509.Certificate) {
	if cert.SerialNumber == nil {
		cert.SerialNumber = big.NewInt(1)
	}
	if len(cert.Subject.Country) == 0 {
		cert.Subject.Country = []string{DefaultCountry}
	}
	if cert.Subject.CommonName == "" {
		if host, err := os.Hostname(); err != nil {
			panic(perrors.Errorf("os.Hostname: '%w'", err))
		} else {
			if index := strings.Index(host, "."); index != -1 {
				host = host[:index]
			}
			cert.Subject.CommonName = host
		}
	}
	if cert.NotBefore.IsZero() {
		nowUTC := time.Now().UTC()
		year, month, day := nowUTC.Date()
		cert.NotBefore = time.Date(year, month, day, 0, 0, 0, 0, nowUTC.Location())
	}
	if cert.NotAfter.IsZero() {
		notBeforeUTC := cert.NotBefore.UTC()
		year, month, day := notBeforeUTC.Date()
		cert.NotAfter = time.Date(year+notAfterYears, month, day, 0, 0, -1, 0, notBeforeUTC.Location())
	}
	cert.BasicConstraintsValid = true
}

func EnsureSelfSigned(cert *x509.Certificate) {
	if cert.Issuer.CommonName == "" {
		if host, err := os.Hostname(); err != nil {
			panic(perrors.Errorf("os.Hostname: '%w'", err))
		} else {
			if index := strings.Index(host, "."); index != -1 {
				host = host[:index]
			}
			cert.Issuer.CommonName = host + caSubjectSuffix
		}
	}
	if len(cert.Issuer.Country) == 0 {
		cert.Issuer.Country = []string{DefaultCountry}
	}
	if len(cert.Subject.Country) == 0 {
		cert.Subject = cert.Issuer
	}
	cert.IsCA = true
	cert.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	EnsureTemplate(cert)
}

func EnsureServer(cert *x509.Certificate) {
	EnsureTemplate(cert)
	cert.KeyUsage = x509.KeyUsageDigitalSignature
	cert.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
}

func EnsureClient(cert *x509.Certificate) {
	EnsureTemplate(cert)
	cert.KeyUsage = x509.KeyUsageDigitalSignature
	cert.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
}
