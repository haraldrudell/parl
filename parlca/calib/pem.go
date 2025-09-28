/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package calib

// CertLocalhost KeyLocalhost is 2,048 bit RSA where key and certificate can be parsed using
// [github.com/haraldrudell/parl/parlca.ParsePem] in 850 µs
//
//	cert, _, _, err = parlca.ParsePem([]byte(calib.CertLocalhost))
//
// Certificate:
//
//	    Version: 3 (0x2)
//	    Serial Number:
//	        a5:1c:da:a3:13:5e:49:e1:82:38:b5:ed:e1:49:98:99
//	Signature Algorithm: sha256WithRSAEncryption
//	    Issuer: C=US, CN=c66ca-250927
//	    Validity
//	        Not Before: Sep 28 00:00:00 2025 GMT
//	        Not After : Sep 27 23:59:59 2035 GMT
//	    Subject: C=US, CN=c66
//	    X509v3 extensions:
//	            Digital Signature
//	            TLS Web Server Authentication
//	            CA:FALSE
//	            DNS:localhost, IP Address:127.0.0.1, IP Address:0:0:0:0:0:0:0:1
var (
	CertLocalhost = `
Generated on 2025-09-27 17:46:15-07:00 by github.com/haraldrudell/parl
parl: (c) 2018-present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
sha256 fingerprint: 00ebe089
sha1 fingerprint: a59ff502
-----BEGIN CERTIFICATE-----
MIIDUDCCAjigAwIBAgIRAPFlgHTmv0VijCDEx4kfUeAwDQYJKoZIhvcNAQELBQAw
JDELMAkGA1UEBhMCVVMxFTATBgNVBAMTDGM2NmNhLTI1MDkyNzAeFw0yNTA5Mjgw
MDAwMDBaFw0zNTA5MjcyMzU5NTlaMBsxCzAJBgNVBAYTAlVTMQwwCgYDVQQDEwNj
NjYwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDTPhTVyfNEW649t0ts
v5nfW1T9iCJ1W5ibmN8zYY9OBjto71KBmtTqvhPUkBsU7QZIYTUUv5GZV2X9gAXx
XTds8y2BsyqaZ8ZUj/Q+riBG7d4/0u3Qwgoi3O0TcgLXsSLqQ72DWKrWgf72BnNW
F0Kb6DaJlmzw9RQzHwlHqckZuA0WnzoWScn2n3TZbmn/70BotxFUslopHuJ3WcS4
QYvMqMsL35RqWuSwdMeeJfFKDOtVuL9peIsZMWhJ87dq3GJI2MnEd4cG8Td6gOgP
gsvNZlXe0cnFu2mCf4V2l/icKIxNEV1Unh6I5DAi9JdRjBzb6AQI1DrDTErMFn2I
7fM1AgMBAAGjgYUwgYIwDgYDVR0PAQH/BAQDAgeAMBMGA1UdJQQMMAoGCCsGAQUF
BwMBMAwGA1UdEwEB/wQCMAAwHwYDVR0jBBgwFoAUsxSFzakDkur0AZx3i7l1QIUL
wIYwLAYDVR0RBCUwI4IJbG9jYWxob3N0hwR/AAABhxAAAAAAAAAAAAAAAAAAAAAB
MA0GCSqGSIb3DQEBCwUAA4IBAQCSe3sWmxFLCyPKGaOO3PvRsEnoVe+89spRNQIu
hipW2G7wu379nQMePP4yn3pUClhBQBhmjf6VDxBJS8Gq3ooxqWaREdZY5aTDwd8A
6G4zoht2B/v+5WKl2hkT3zq0X5l3B3p8DdfawLC6es0iB3JwEKMBn4IzFnBo+gOU
A3Jrp1xTDSLGGZZS24IWQkzjB5IiDt0DzEBRzQdJVrcpv55iruerl7hAHhTr2inQ
PDc7rQvDnBDd93gydWGAN00otwGnWyOe4R/6MN7yDzYVI5bltECCf7Ye7QuzfDjU
GA9XS9RuHiJ5xBQuqtyqD0pAU2DvFZ2dlE99iemwrlvKaM46
-----END CERTIFICATE-----
`
	KeyLocalhost = `Generated on 2025-09-27 17:46:15-07:00 by github.com/haraldrudell/parl
parl: (c) 2018-present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
sha256 fingerprint: 3464c89b
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDTPhTVyfNEW649
t0tsv5nfW1T9iCJ1W5ibmN8zYY9OBjto71KBmtTqvhPUkBsU7QZIYTUUv5GZV2X9
gAXxXTds8y2BsyqaZ8ZUj/Q+riBG7d4/0u3Qwgoi3O0TcgLXsSLqQ72DWKrWgf72
BnNWF0Kb6DaJlmzw9RQzHwlHqckZuA0WnzoWScn2n3TZbmn/70BotxFUslopHuJ3
WcS4QYvMqMsL35RqWuSwdMeeJfFKDOtVuL9peIsZMWhJ87dq3GJI2MnEd4cG8Td6
gOgPgsvNZlXe0cnFu2mCf4V2l/icKIxNEV1Unh6I5DAi9JdRjBzb6AQI1DrDTErM
Fn2I7fM1AgMBAAECggEAG5twc6RKA7QLqIss81BzFqrEB8Jj5nyLDELFYNyuMt9D
omosmT2X9/nRc6VFtM1pVcpGeqiyqZyveprhq/KnzLhXSS0WM0m+tMx/ejWdiEeM
FUFGzVKeqoG/BhyPXIsM6hriSKks4o3ouVSMfELb9K47em3LXQ5ajnfO6au52X9k
hbGiw3Hc6bFqzbAuCTbAhCTwPw+BAUDMLym/fhnB/IO6msqjcxYajXWQjlrVp4Mf
t7pHFqqUKoX/9PwW9Ggw2uJoA+HOgvNkBAEV2kYpVuKIBm0SeVtye6tBxD1ckcsK
ygu3haaN6FS3z//9trQbWQm+IYORWfeIQflxMFOLQQKBgQDnhrKGc4ddzSI76KtX
Urxes5ELJVgUK5MSE0QLhTlPiphCKBj6ja3stb92dGm0xVQjcyawJhfNXA2uLgmk
OrTf/exwlSwIAzwTunMP26fUvSKBmb7KqTbzXGjKFrrVAtBh6Ry22UdMuKz+UOeF
p1Sq8RUGWuosYeVi3n3TC9NSEQKBgQDpkn2H0OO1UF1vfCKi+CMUEuWho1wQlfyD
P5ehtgdOZIFcyazjovF8OMEnFOQUCctsuk9h8CWQ16VuZoQgBotcTSg4HmmzxqWo
ec/MkKFwrkrbFfdSfTYB6BtKvNCmsBoCPy1NlKclDtCFdWaYaf2EtpNK0L57Xrcm
xQu768Lq5QKBgBLeUlxMvAJz/k89lgEm1/0ryy1KXNQ//NtjQI9jyxjlZaU2mVqK
A1ugMDtaH2dBEatV7hg68oOk9eR1EgoVUrpSeltufMkmYlYFQu1O7G2VAGPpgLgJ
dFX++PdHRPCpKwxfsBxitsHU5xxOwZ+N1IOd5CXBcQYu8D/PfBegBhfRAoGAPN5/
NkC6xsqNvvrzr9LImXefPjNrT6s2piGRa4QbfVN13u9zzdLt6biEpaGtGoe+6rPW
8if6MjfwlcfDkPEDqmc1HwLV+xK+oxwzihT67XmOam/cBzQ4OeD6E80G9xmXfZRn
QvuFX4Pv1YfV18xvVAGcevfRXCc/xau+NhfnsP0CgYEAiFXPyjCebvr8DZrEBa4z
PKn0V8KcgCSJ7vdB+ARfTlk2iSCEfwKYBcIFMpXnmCrs/rNnu2r9o+r2wLR6KiJ2
fpIdz6MIYgo5rE4HzC09iHVZ28WGSFGqN3xBsgidZ2Ak55GgrNOCzyFDDRlfE8Lb
pcO+9hOQFJvwV4adsZW3FI0=
-----END PRIVATE KEY-----
`
)
