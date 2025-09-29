/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/x509"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pnet"
	"github.com/haraldrudell/parl/pos"
)

const (
	NoAppName   = ""
	DotLocalDir = ""
)

// ReadOrCreateCredentials does write-once server-credentials caching
//   - appName: the app’s name used for naming files and directories
//   - appName [NoAppName] "": use “test”
//   - dirname: directory for pem format pki files
//   - dirname [DotLocalDir] "": directory is: ~/.local/[appName]
//   - algo: algorithm for keys and certificates
//   - canonicalName: unique name for certificate authority
//   - canonicalName [DefaultCAName] "": CA certificate subject is “[hostname]ca-[date6]” “hostca-251231”
//   - log: printing filenames being written for certificates and keys
//   - ipsAndDomains: the IP literals and domain names for which the certificate is valid
//   - ipsAndDomains missing: IPs are “::1” “127.0.0.1” domains are “localhost”
func ReadOrCreateCredentials(
	appName,
	dirName string,
	algo x509.PublicKeyAlgorithm,
	canonicalName string,
	log parl.PrintfFunc,
	ipsAndDomains ...parl.AnyCount[pnet.Address],
) (
	cert parl.Certificate,
	key parl.PrivateKey,
	err error,
) {
	if appName == "" {
		appName = defaultAppName
	}
	if dirName == DotLocalDir {
		var appd *pos.AppDirectory = pos.NewAppDir(appName)
		appd.EnsureDir()
		dirName = appd.Directory()
	}
	// hostname “c66”
	var host = pos.ShortHostname()
	var certFile = filepath.Join(dirName, host+"-cert.pem")
	var keyFile = filepath.Join(dirName, host+"-key.pem")
	var isFileNotFound bool
	if cert, key, isFileNotFound, err = ReadCredentials(certFile, keyFile); err != nil {
		return
	} else if !isFileNotFound {
		return
	}
	// must create credentials

	// create ca and credentials
	var caCertDER parl.CertificateDer
	var caKey parl.PrivateKey
	if cert, key, caCertDER, caKey, err = CreateCredentials(algo, canonicalName, ipsAndDomains...); err != nil {
		return
	}

	// save to storage
	var writes = []struct {
		filename string
		data     []byte
	}{{
		certFile, cert.DER(),
	}, {
		keyFile, key.DERe(),
	}, {
		filepath.Join(dirName, appName+"-ca.pem"), caCertDER,
	}, {
		filepath.Join(dirName, appName+"-ca-key.pem"), caKey.DERe(),
	}}
	for _, toWrite := range writes {
		log("Writing: %q", toWrite.filename)
		if err = os.WriteFile(toWrite.filename, toWrite.data, fileModeUr); perrors.IsPF(&err, "os.WriteFile %w", err) {
			return
		}
	}

	return
}

const (
	defaultAppName = "test"
	// file mode for created credentials: r-- --- ---
	fileModeUr fs.FileMode = 0400
)
