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

	"github.com/google/uuid"
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/ptime"
)

// hyphen is separator in self-signed ca name
const hyphen = "-"

// EnsureTemplate ensures that cert can be signed
//   - use as [x509.CreateCertificate] template argument
//   - [x509.Certificate.SerialNumber] non-nil.
//     Default: uuid 128-bit number for 160-bit field.
//     The certificate is uniquely identified by its
//     serial number alone, a 39-digit decimal number.
//     uuid to ensure certificates signed by equally named
//     but different certificate authorities are still unique
//   - [x509.Certificate.Subject.CommonName] is non-empty.
//     Default short hostname
//   - — “c66” from “c66.example.com”
//   - [x509.Certificate.Subject.Country] is non-empty.
//     Default “US”
//   - [x509.Certificate.NotBefore] is non-zero.
//     Default: today in UTC time zone at 0:0:0
//   - [x509.Certificate.NotAfter] is non-zero.
//     Default: NotBefore + 10 years - one minute
//   - [x509.Certificate.BasicConstraintsValid] true
//   - certificate uniqueness is commonly (macOS):
//   - — certificate serial number
//   - — certificate issuer name, ie. certificate authority name
//   - — certificate type: X.509 Version 3
func EnsureTemplate(cert *x509.Certificate) {

	// serial number uuid
	if cert.SerialNumber == nil {
		cert.SerialNumber = uuidSerialNumber()
	}

	// subject
	if len(cert.Subject.Country) == 0 {
		cert.Subject.Country = []string{DefaultCountry}
	}
	if cert.Subject.CommonName == "" {
		cert.Subject.CommonName = shortHostname()
	}

	if cert.NotBefore.IsZero() {
		var nowUTC = time.Now().UTC()
		var year, month, day = nowUTC.Date()
		cert.NotBefore = time.Date(year, month, day, 0, 0, 0, 0, nowUTC.Location())
	}
	if cert.NotAfter.IsZero() {
		var notBeforeUTC = cert.NotBefore.UTC()
		var year, month, day = notBeforeUTC.Date()
		cert.NotAfter = time.Date(year+notAfterYears, month, day, 0, 0, -1, 0, notBeforeUTC.Location())
	}
	cert.BasicConstraintsValid = true
}

// EnsureSelfSigned ensures that cert can be used as
// self-signed certificate authority
//   - use as [x509.CreateCertificate] both template and parent arguments
//   - [x509.Certificate.Issuer.CommonName] is non-empty.
//     Default short hostname + “ca” + 6-digit local-time date.
//     If only one certificate authority is generated per host and day,
//     its name alone is a unique identifier.
//   - — “c66ca-241231”
//   - [x509.Certificate.Issuer.Country] is non-empty.
//     Default “US”
//   - [x509.Certificate.Subject] is non-empty.
//     Default is issuer
//   - [x509.Certificate.IsCA] true
//   - [x509.Certificate.KeyUsage] includes [x509.KeyUsageCertSign] and [x509.KeyUsageCRLSign]
//   - additionally: [EnsureTemplate] values
func EnsureSelfSigned(cert *x509.Certificate) {

	// ensure issuer common name
	if cert.Issuer.CommonName == "" {
		// get short hostname “c66” from “c66.example.com”
		var hostname = shortHostname()
		// date as 241231, local time zone
		var date6 = time.Now().Format(ptime.Date6)
		// “c66ca-241231”
		cert.Issuer.CommonName = hostname + caSubjectSuffix + hyphen + date6
	}

	// ensure issuer country
	if len(cert.Issuer.Country) == 0 {
		cert.Issuer.Country = []string{DefaultCountry}
	}

	// ensure subject, default issuer
	if len(cert.Subject.Country) == 0 {
		cert.Subject = cert.Issuer
	}
	cert.IsCA = true
	cert.KeyUsage |= x509.KeyUsageCertSign | x509.KeyUsageCRLSign

	EnsureTemplate(cert)
}

// EnsureServer ensures cert can be signed and
// used as a server certificate
//   - enables use as template argument to [x509.CreateCertificate]
//   - use is X.509 and X.509 v3 server authentication
func EnsureServer(cert *x509.Certificate) {
	EnsureTemplate(cert)
	cert.KeyUsage |= x509.KeyUsageDigitalSignature
	cert.ExtKeyUsage = append(cert.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
}

// EnsureClient ensures cert can be signed and
// used as a client certificate
//   - enables use as template argument to [x509.CreateCertificate]
//   - use is X.509 and X.509 v3 client authentication
func EnsureClient(cert *x509.Certificate) {
	EnsureTemplate(cert)
	cert.KeyUsage |= x509.KeyUsageDigitalSignature
	cert.ExtKeyUsage = append(cert.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
}

// uuidSerialNumber returns a certificate serial number based on uuid
//   - 39 decimal digits
//   - serialNumber: a certificate serial number different from
//     all other serial numbers
//   - the benefit of using a uuid serial number is that
//     certificate stores will never consider two certificates
//     the same, even when their certificate authorities have the same name
//   - certificate uniqueness is commonly (macOS):
//   - — certificate serial number
//   - — certificate issuer name, ie. certificate authority name
//   - — certificate type: X.509 Version 3
//   - panic on troubles: will not happen
func uuidSerialNumber() (serialNumber *big.Int) {
	// [big.Int] is an umlimited arbtitrary sized integer implemented as binary array
	serialNumber = &big.Int{}

	// b is the 16 binary bytes of a new 128-bit uuid
	var b []byte
	var err error
	if b, err = uuid.New().MarshalBinary(); err != nil {
		panic(perrors.ErrorfPF("uuid.MarshalBinary %w", err))
	}

	serialNumber.SetBytes(b)

	return
}

// shortHostname returns hostname as single space-free word
//   - a-z 0-9 hyphen 7-bit US ASCII, 1 to 63 characters long
//   - “c66” from “c66.example.com”
//   - panic on error
func shortHostname() (hostname string) {
	var err error
	if hostname, err = os.Hostname(); err != nil {
		panic(perrors.Errorf("os.Hostname: “%w”", err))
	} else if index := strings.Index(hostname, "."); index != -1 {
		hostname = hostname[:index]
	}
	if hostname == "" {
		panic(parl.NilError("hostname"))
	}
	return
}
