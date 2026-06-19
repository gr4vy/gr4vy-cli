package secret

import (
	"errors"

	keyring "github.com/zalando/go-keyring"
)

// keyringService is the service name under which all gr4vy secrets are stored
// in the OS keychain.
const keyringService = "gr4vy-cli"

type keyringStore struct{}

func newKeyringStore() *keyringStore { return &keyringStore{} }

func (s *keyringStore) Get(profile string, kind Kind) (string, error) {
	v, err := keyring.Get(keyringService, account(profile, kind))
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", err
	}
	return v, nil
}

func (s *keyringStore) Set(profile string, kind Kind, value string) error {
	return keyring.Set(keyringService, account(profile, kind), value)
}

func (s *keyringStore) Delete(profile string, kind Kind) error {
	err := keyring.Delete(keyringService, account(profile, kind))
	if errors.Is(err, keyring.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *keyringStore) Backend() string { return BackendKeyring }

// keyringAvailable probes whether a working keychain backend is present. A
// "not found" result means the service responded (available); any other error
// means the backend is unusable (e.g. no Secret Service on a headless Linux box).
func keyringAvailable() bool {
	_, err := keyring.Get(keyringService, "__availability_probe__")
	return err == nil || errors.Is(err, keyring.ErrNotFound)
}
