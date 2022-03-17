/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"io"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/haraldrudell/parl/error116"
)

type SelfSigned struct {
	Reader io.Reader
	CaDER  CertificateDER // der: Distinguished Encoding Rules is a binary format
	KeyPair
}

type DER []byte

//func NewCertificateAuthority()
func NewSelfSigned(canonicalName string) (ca CertificateAuthority) {

	// get private key
	var keyPair KeyPair
	var err error
	if keyPair, err = NewEd25519(); err != nil {
		panic(err)
	}

	//  reate certificate authority
	c := SelfSigned{Reader: rand.Reader, KeyPair: keyPair}

	// sign self-signed ca certificate
	cert := &x509.Certificate{}
	EnsureSelfSigned(cert)
	signer := c.Private()
	if c.CaDER, err = x509.CreateCertificate(rand.Reader, cert, cert, signer.Public(), signer); err != nil {
		panic(error116.Errorf("x509.CreateCertificate ca: '%w'", err))
	}
	ca = &c
	return
}

const (
	/*
		NoPassword       PasswordType = "\tnoPassword"
		GeneratePassword PasswordType = "\tgeneratePassword"
		GenerateOnTheFly Strategy     = iota << 0
		UseFileSystem
		DefaultStrategy = GenerateOnTheFly
	*/
	DefaultCountry  = "US"
	notAfterYears   = 10
	caSubjectSuffix = "ca"
)

func (ca *SelfSigned) Sign(template *x509.Certificate, publicKey crypto.PublicKey) (certDER CertificateDER, err error) {

	// get certificate authority x509.Certificate
	var isValid bool
	var caCert *x509.Certificate
	if isValid, caCert, err = ca.Check(); err != nil {
		return
	} else if !isValid {
		err = error116.New("Self-Signed Certiicate Auhtority not valid")
		return
	}
	caSigner := ca.KeyPair.Private()

	// sign template
	if certDER, err = x509.CreateCertificate(ca.Reader, template, caCert, publicKey, caSigner); err != nil {
		err = error116.Errorf("x509.CreateCertificate: '%w'", err)
		return
	}
	return
}

func (ca *SelfSigned) Check() (isValid bool, cert *x509.Certificate, err error) {
	if !ca.HasKey() || !ca.HasDER() {
		return
	}
	var c x509Certificate
	if c.Certificate, err = x509.ParseCertificate(ca.CaDER); err != nil {
		err = error116.Errorf("x509.ParseCertificate: '%w'", err)
		return
	}
	isValid = c.HasPublic()
	cert = c.Certificate
	return
}

func EnsureTemplate(cert *x509.Certificate) {
	if cert.SerialNumber == nil {
		cert.SerialNumber = big.NewInt(1)
	}
	if len(cert.Subject.Country) == 0 {
		cert.Subject.Country = []string{DefaultCountry}
	}
	if cert.Subject.CommonName == "" {
		if host, err := os.Hostname(); err != nil {
			panic(error116.Errorf("os.Hostname: '%w'", err))
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
			panic(error116.Errorf("os.Hostname: '%w'", err))
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

func (ca *SelfSigned) DER() (bytes CertificateDER) {
	return ca.CaDER
}

func (ca *SelfSigned) SetReader(reader io.Reader) {
	ca.Reader = reader
}

func (ca *SelfSigned) HasDER() (hasDER bool) {
	hasDER = len(ca.CaDER) > 0
	return
}
