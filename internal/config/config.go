// Package config loads, persists, and resolves the gr4vy CLI configuration:
// a TOML file holding one or more named profiles plus an active-profile
// pointer. Only non-secret data lives here — private keys and login tokens are
// handled by the secret package and are never written to this file.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/pelletier/go-toml/v2"
)

// Environment values for a profile's server selection.
const (
	EnvSandbox    = "sandbox"
	EnvProduction = "production"
)

// Authentication methods for a profile.
const (
	AuthKey   = "key"   // sign ES512 JWTs from a private key
	AuthLogin = "login" // email/password session via /auth/sessions
)

// Key reference backends, describing where a profile's private key lives.
const (
	KeyRefStore = "store" // in the secret store (OS keychain or file fallback)
	KeyRefFile  = "file"  // a user-managed PEM file at KeyPath
	KeyRefEnv   = "env"   // a named environment variable (KeyEnv)
)

// Profile is a single named gr4vy configuration. It carries only non-secret
// data; credential material is resolved separately via the secret package.
type Profile struct {
	ID                string `toml:"id"`
	Environment       string `toml:"environment"`
	MerchantAccountID string `toml:"merchant_account_id,omitempty"`
	AuthMethod        string `toml:"auth_method,omitempty"`

	// Key auth — references only, never the PEM bytes.
	KeyRef  string `toml:"key_ref,omitempty"`
	KeyPath string `toml:"key_path,omitempty"`
	KeyEnv  string `toml:"key_env,omitempty"`

	// Login auth — non-secret metadata only.
	Email    string `toml:"email,omitempty"`
	AuthHost string `toml:"auth_host,omitempty"`

	// Defaults for the `token` command.
	DefaultScopes []string `toml:"default_scopes,omitempty"`
	TokenTTL      string   `toml:"token_ttl,omitempty"`
}

// Config is the full on-disk configuration.
type Config struct {
	ActiveProfile string             `toml:"active_profile,omitempty"`
	Profiles      map[string]Profile `toml:"profiles,omitempty"`
}

// Names returns the profile names in sorted order.
func (c *Config) Names() []string {
	names := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Lookup returns the named profile and whether it exists.
func (c *Config) Lookup(name string) (Profile, bool) {
	p, ok := c.Profiles[name]
	return p, ok
}

// Set stores (or replaces) a profile by name.
func (c *Config) Set(name string, p Profile) {
	if c.Profiles == nil {
		c.Profiles = map[string]Profile{}
	}
	c.Profiles[name] = p
}

// Remove deletes a profile and clears the active pointer if it referenced it.
func (c *Config) Remove(name string) {
	delete(c.Profiles, name)
	if c.ActiveProfile == name {
		c.ActiveProfile = ""
	}
}

// DefaultPath resolves the config file path following XDG conventions:
//
//	$GR4VY_CONFIG > $XDG_CONFIG_HOME/gr4vy/config.toml > ~/.config/gr4vy/config.toml
func DefaultPath() (string, error) {
	if v := os.Getenv(EnvConfig); v != "" {
		return v, nil
	}
	if base := os.Getenv("XDG_CONFIG_HOME"); base != "" {
		return filepath.Join(base, "gr4vy", "config.toml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "gr4vy", "config.toml"), nil
}

// Load reads the config file at path. A missing file is not an error: it
// returns an empty Config so commands can guide first-time setup.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{Profiles: map[string]Profile{}}, nil
		}
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var c Config
	if err := toml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	if c.Profiles == nil {
		c.Profiles = map[string]Profile{}
	}
	return &c, nil
}

// Save writes the config to path, creating parent directories (0700) and the
// file (0600) since it may hold semi-sensitive references.
func Save(path string, c *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir %s: %w", dir, err)
	}
	data, err := toml.Marshal(c)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}
	return nil
}
