// Package auth turns a resolved profile into a bearer-token provider and an
// authenticated gr4vy-go client. Two providers are supported: key-based (sign
// ES512 JWTs locally) and login-based (email/password session with refresh).
package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	gr4vygo "github.com/gr4vy/gr4vy-go"
	"github.com/gr4vy/gr4vy-go/models/components"

	"github.com/gr4vy/gr4vy-cli/internal/config"
	"github.com/gr4vy/gr4vy-cli/internal/secret"
)

// Credential-related environment variables.
const (
	EnvPrivateKey     = "GR4VY_PRIVATE_KEY"      // raw PEM
	EnvPrivateKeyFile = "GR4VY_PRIVATE_KEY_FILE" // path to PEM
	EnvPassword       = "GR4VY_PASSWORD"
	EnvAccessToken    = "GR4VY_ACCESS_TOKEN"
)

// defaultTokenTTL is the lifetime, in seconds, of JWTs minted for API calls.
const defaultTokenTTL = 3600

// TokenProvider yields a valid bearer token for an API call, minting or
// refreshing as needed.
type TokenProvider interface {
	Token(ctx context.Context) (string, error)
	Scopes() []gr4vygo.JWTScope
}

// StaticTokenProvider returns a fixed, pre-generated token (from --token or
// GR4VY_TOKEN). It never refreshes.
type StaticTokenProvider struct{ T string }

func (p StaticTokenProvider) Token(context.Context) (string, error) { return p.T, nil }
func (p StaticTokenProvider) Scopes() []gr4vygo.JWTScope            { return nil }

// KeyTokenProvider signs a fresh ES512 JWT per call from a PEM private key.
type KeyTokenProvider struct {
	PEM       string
	ScopeSet  []gr4vygo.JWTScope
	ExpiresIn int // seconds
}

func (p *KeyTokenProvider) Token(context.Context) (string, error) {
	ttl := p.ExpiresIn
	if ttl <= 0 {
		ttl = defaultTokenTTL
	}
	return gr4vygo.GetToken(p.PEM, p.ScopeSet, ttl)
}

func (p *KeyTokenProvider) Scopes() []gr4vygo.JWTScope {
	if len(p.ScopeSet) == 0 {
		return []gr4vygo.JWTScope{gr4vygo.ReadAll, gr4vygo.WriteAll}
	}
	return p.ScopeSet
}

// SecuritySource adapts a TokenProvider to the gr4vy-go security source, which
// the SDK invokes on every request (so refresh/re-sign is transparent).
func SecuritySource(p TokenProvider) func(context.Context) (components.Security, error) {
	return func(ctx context.Context) (components.Security, error) {
		tok, err := p.Token(ctx)
		if err != nil {
			return components.Security{}, err
		}
		return components.Security{BearerAuth: &tok}, nil
	}
}

// NewClient builds an authenticated gr4vy-go client for the resolved profile.
func NewClient(r config.Resolved, p TokenProvider, timeout time.Duration) (*gr4vygo.Gr4vy, error) {
	if r.Profile.ID == "" {
		return nil, fmt.Errorf("no instance id configured: set --id, GR4VY_ID, or a profile")
	}
	opts := []gr4vygo.SDKOption{
		gr4vygo.WithID(r.Profile.ID),
		gr4vygo.WithServer(serverFor(r.Profile.Environment)),
		gr4vygo.WithSecuritySource(SecuritySource(p)),
	}
	if r.Profile.MerchantAccountID != "" {
		opts = append(opts, gr4vygo.WithMerchantAccountID(r.Profile.MerchantAccountID))
	}
	if timeout > 0 {
		opts = append(opts, gr4vygo.WithTimeout(timeout))
	}
	return gr4vygo.New(opts...), nil
}

// serverFor maps a config environment to a gr4vy-go server name.
func serverFor(environment string) string {
	if environment == config.EnvProduction {
		return gr4vygo.ServerProduction
	}
	return gr4vygo.ServerSandbox
}

// APIBaseURL returns the API host for an instance id and environment, using the
// same templates as the SDK (so e.g. login can hit /auth/sessions there).
func APIBaseURL(id, environment string) string {
	tmpl := gr4vygo.ServerList[serverFor(environment)]
	return strings.ReplaceAll(tmpl, "{id}", id)
}

// BuildProvider selects and constructs the TokenProvider for an API command,
// based on the resolved auth method and available credentials.
func BuildProvider(r config.Resolved, store secret.Store, env config.EnvLookup) (TokenProvider, error) {
	if env == nil {
		env = config.OSEnv
	}
	if r.Token != "" {
		return StaticTokenProvider{T: r.Token}, nil
	}
	switch r.Profile.AuthMethod {
	case config.AuthLogin:
		return NewLoginTokenProvider(r, store, env)
	default: // key
		pem, err := ResolveKeyPEM(r, store, env)
		if err != nil {
			return nil, err
		}
		return &KeyTokenProvider{PEM: pem, ExpiresIn: defaultTokenTTL}, nil
	}
}

// ResolveKeyPEM returns the PEM private key for the resolved profile, honouring
// CI overrides (raw env, file env) before the profile's stored key reference.
// Any source may supply a base64-encoded PEM; it is decoded transparently.
func ResolveKeyPEM(r config.Resolved, store secret.Store, env config.EnvLookup) (string, error) {
	raw, err := resolveRawKeyPEM(r, store, env)
	if err != nil {
		return "", err
	}
	return secret.NormalizePEM(raw), nil
}

func resolveRawKeyPEM(r config.Resolved, store secret.Store, env config.EnvLookup) (string, error) {
	if env == nil {
		env = config.OSEnv
	}
	if v, ok := env(EnvPrivateKey); ok && v != "" {
		return v, nil
	}
	if path, ok := env(EnvPrivateKeyFile); ok && path != "" {
		return readKeyFile(path)
	}

	p := r.Profile
	switch p.KeyRef {
	case config.KeyRefEnv:
		if p.KeyEnv == "" {
			return "", fmt.Errorf("profile %q uses key_ref=env but key_env is not set", r.ProfileName)
		}
		if v, ok := env(p.KeyEnv); ok && v != "" {
			return v, nil
		}
		return "", fmt.Errorf("environment variable %s (key_env for profile %q) is not set", p.KeyEnv, r.ProfileName)
	case config.KeyRefFile:
		if p.KeyPath == "" {
			return "", fmt.Errorf("profile %q uses key_ref=file but key_path is not set", r.ProfileName)
		}
		return readKeyFile(p.KeyPath)
	case config.KeyRefStore, "":
		if store == nil {
			return "", fmt.Errorf("no secret store available")
		}
		v, err := store.Get(r.ProfileName, secret.KindKey)
		if errors.Is(err, secret.ErrNotFound) {
			return "", fmt.Errorf("no private key stored for profile %q; run `gr4vy init` or set %s", r.ProfileName, EnvPrivateKey)
		}
		return v, err
	default:
		return "", fmt.Errorf("unknown key_ref %q for profile %q", p.KeyRef, r.ProfileName)
	}
}

func readKeyFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("read private key %s: %w", path, err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		fmt.Fprintf(os.Stderr, "warning: private key %s is readable by others (%#o); consider chmod 600\n", path, info.Mode().Perm())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read private key %s: %w", path, err)
	}
	return string(data), nil
}
