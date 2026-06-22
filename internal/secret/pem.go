package secret

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"
)

// NormalizePEM accepts either a raw PEM string or a base64-encoded PEM (handy
// for CI env vars that can't carry newlines) and returns the decoded PEM. If
// the input is neither, it is returned unchanged so validation can report it.
func NormalizePEM(in string) string {
	s := strings.TrimSpace(in)
	if strings.Contains(s, "-----BEGIN") {
		return s
	}
	compact := strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\r', '\t', ' ':
			return -1
		}
		return r
	}, s)
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding,
	} {
		if decoded, err := enc.DecodeString(compact); err == nil && strings.Contains(string(decoded), "-----BEGIN") {
			return strings.TrimSpace(string(decoded))
		}
	}
	return s
}

// ValidatePrivateKeyPEM checks that pemData is a PKCS#8-encoded ECDSA key on
// the P-521 curve — the only key the gr4vy API accepts (ES512). It accepts a
// base64-encoded PEM as well. Validating at import time surfaces a clear error
// instead of an opaque 401 on first use.
func ValidatePrivateKeyPEM(pemData string) error {
	block, _ := pem.Decode([]byte(NormalizePEM(pemData)))
	if block == nil {
		return fmt.Errorf("not a valid PEM block")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("not a PKCS#8 private key: %w", err)
	}
	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return fmt.Errorf("private key is not ECDSA (gr4vy requires ES512/P-521)")
	}
	if ecKey.Curve != elliptic.P521() {
		return fmt.Errorf("unsupported curve %q: gr4vy requires P-521 (ES512)", ecKey.Curve.Params().Name)
	}
	return nil
}
