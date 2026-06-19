package secret

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// fileStore persists secrets as 0600 files under a 0700 directory. It is the
// fallback when no OS keychain is available (headless boxes, CI).
type fileStore struct {
	dir string
}

func newFileStore() (*fileStore, error) {
	dir, err := defaultSecretsDir()
	if err != nil {
		return nil, err
	}
	return &fileStore{dir: dir}, nil
}

func defaultSecretsDir() (string, error) {
	if base := os.Getenv("XDG_CONFIG_HOME"); base != "" {
		return filepath.Join(base, "gr4vy", "secrets"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "gr4vy", "secrets"), nil
}

func (s *fileStore) path(profile string, kind Kind) string {
	return filepath.Join(s.dir, safeName(profile)+"."+string(kind))
}

func (s *fileStore) Get(profile string, kind Kind) (string, error) {
	data, err := os.ReadFile(s.path(profile, kind))
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNotFound
		}
		return "", err
	}
	return string(data), nil
}

func (s *fileStore) Set(profile string, kind Kind, value string) error {
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return fmt.Errorf("create secrets dir %s: %w", s.dir, err)
	}
	if err := os.WriteFile(s.path(profile, kind), []byte(value), 0o600); err != nil {
		return fmt.Errorf("write secret: %w", err)
	}
	return nil
}

func (s *fileStore) Delete(profile string, kind Kind) error {
	err := os.Remove(s.path(profile, kind))
	if os.IsNotExist(err) {
		return ErrNotFound
	}
	return err
}

func (s *fileStore) Backend() string { return BackendFile }

// safeName makes a profile name safe to use as a filename component.
func safeName(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			return r
		default:
			return '_'
		}
	}, s)
}
