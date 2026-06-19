// Package secret stores and retrieves credential material (private keys and
// login token bundles) outside the plaintext config file. The primary backend
// is the OS keychain; a 0600 file backend is used as a fallback on headless
// systems. It also validates that a private key is the ES512/P-521 key the
// gr4vy API requires.
package secret

import (
	"errors"
	"os"
)

// Kind identifies the type of secret stored for a profile.
type Kind string

const (
	KindKey   Kind = "key"   // PEM-encoded private key
	KindLogin Kind = "login" // JSON session bundle (access/refresh tokens)
)

// Backend selection.
const (
	BackendAuto    = "auto"
	BackendKeyring = "keyring"
	BackendFile    = "file"
)

// EnvSecretBackend overrides the backend selection.
const EnvSecretBackend = "GR4VY_SECRET_BACKEND"

// ErrNotFound is returned by Store.Get when no secret exists.
var ErrNotFound = errors.New("secret not found")

// Store persists per-profile secrets keyed by Kind.
type Store interface {
	Get(profile string, kind Kind) (string, error)
	Set(profile string, kind Kind, value string) error
	Delete(profile string, kind Kind) error
	// Backend reports the concrete backend name ("keyring" or "file").
	Backend() string
}

// Open returns a Store for the requested backend. The empty string and
// "auto" select the keychain when available, otherwise the file backend.
func Open(backend string) (Store, error) {
	if backend == "" {
		backend = os.Getenv(EnvSecretBackend)
	}
	switch backend {
	case BackendKeyring:
		return newKeyringStore(), nil
	case BackendFile:
		return newFileStore()
	case "", BackendAuto:
		if keyringAvailable() {
			return newKeyringStore(), nil
		}
		return newFileStore()
	default:
		return nil, errors.New("unknown secret backend: " + backend)
	}
}

// account composes the keychain account / file name for a (profile, kind).
func account(profile string, kind Kind) string {
	return string(kind) + ":" + profile
}
