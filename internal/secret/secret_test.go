package secret

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"strings"
	"testing"
)

func TestFileStoreRoundTrip(t *testing.T) {
	s := &fileStore{dir: t.TempDir()}

	if _, err := s.Get("default", KindKey); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := s.Set("default", KindKey, "secret-value"); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get("default", KindKey)
	if err != nil {
		t.Fatal(err)
	}
	if got != "secret-value" {
		t.Fatalf("got %q", got)
	}
	if err := s.Delete("default", KindKey); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get("default", KindKey); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func pkcs8PEM(t *testing.T, curve elliptic.Curve) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

func TestValidatePrivateKeyPEM(t *testing.T) {
	if err := ValidatePrivateKeyPEM(pkcs8PEM(t, elliptic.P521())); err != nil {
		t.Errorf("valid P-521 key rejected: %v", err)
	}
	if err := ValidatePrivateKeyPEM(pkcs8PEM(t, elliptic.P256())); err == nil {
		t.Error("P-256 key should be rejected")
	}
	if err := ValidatePrivateKeyPEM("not a pem"); err == nil {
		t.Error("garbage should be rejected")
	}
}

func TestNormalizePEM(t *testing.T) {
	pemStr := pkcs8PEM(t, elliptic.P521())

	// Raw PEM passes through unchanged (trimmed).
	if got := NormalizePEM(pemStr); got != strings.TrimSpace(pemStr) {
		t.Error("raw PEM should pass through unchanged")
	}

	// Base64-encoded PEM decodes back to the PEM.
	b64 := base64.StdEncoding.EncodeToString([]byte(pemStr))
	if got := NormalizePEM(b64); got != strings.TrimSpace(pemStr) {
		t.Errorf("base64 PEM did not decode back; got %.40q...", got)
	}
	// ...and validates as a key.
	if err := ValidatePrivateKeyPEM(b64); err != nil {
		t.Errorf("base64-encoded P-521 key rejected: %v", err)
	}

	// Base64 with embedded newlines (e.g. `base64` wrapping) still works.
	wrapped := strings.Join([]string{b64[:40], b64[40:]}, "\n")
	if err := ValidatePrivateKeyPEM(wrapped); err != nil {
		t.Errorf("wrapped base64 key rejected: %v", err)
	}

	// Non-PEM, non-base64 input is returned unchanged.
	if got := NormalizePEM("not a key"); got != "not a key" {
		t.Errorf("garbage should pass through; got %q", got)
	}
}
