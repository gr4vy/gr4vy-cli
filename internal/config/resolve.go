package config

import (
	"fmt"
	"os"
)

// Environment variable names recognised by the CLI. Credential-specific
// variables (private key, password, access token, secret backend) are owned by
// the secret and auth packages.
const (
	EnvConfig            = "GR4VY_CONFIG"
	EnvProfile           = "GR4VY_PROFILE"
	EnvID                = "GR4VY_ID"
	EnvServer            = "GR4VY_SERVER"
	EnvEnvironment       = "GR4VY_ENVIRONMENT" // alias for EnvServer
	EnvMerchantAccountID = "GR4VY_MERCHANT_ACCOUNT_ID"
	EnvAuthMethod        = "GR4VY_AUTH_METHOD"
	EnvAuthHost          = "GR4VY_AUTH_HOST"
	EnvEmail             = "GR4VY_EMAIL"
	EnvToken             = "GR4VY_TOKEN"
)

// EnvLookup resolves an environment variable, returning its value and whether
// it was set. It mirrors os.LookupEnv and can be substituted in tests.
type EnvLookup func(key string) (string, bool)

// OSEnv is the default EnvLookup backed by the process environment.
func OSEnv(key string) (string, bool) { return os.LookupEnv(key) }

// Overrides are values supplied on the command line (highest precedence).
type Overrides struct {
	Profile           string
	ID                string
	Environment       string
	MerchantAccountID string
	Token             string
}

// Resolved is the effective configuration for a single invocation, after
// applying precedence: flag > env > active profile > built-in default.
type Resolved struct {
	ProfileName string
	Profile     Profile // effective profile with overrides applied
	Token       string  // explicit bearer token (flag/env); bypasses signing
}

// Resolve computes the effective settings from the loaded config, command-line
// overrides, and the environment.
func Resolve(c *Config, ov Overrides, env EnvLookup) (Resolved, error) {
	if env == nil {
		env = OSEnv
	}

	name := firstNonEmpty(ov.Profile, envOr(env, EnvProfile), c.ActiveProfile)
	if name == "" {
		if _, ok := c.Profiles["default"]; ok {
			name = "default"
		} else if len(c.Profiles) == 1 {
			name = c.Names()[0]
		}
	}

	base := c.Profiles[name] // zero Profile if absent (env may supply everything)

	eff := base
	eff.ID = firstNonEmpty(ov.ID, envOr(env, EnvID), base.ID)
	eff.Environment = firstNonEmpty(ov.Environment, envOr(env, EnvServer), envOr(env, EnvEnvironment), base.Environment, EnvSandbox)
	eff.MerchantAccountID = firstNonEmpty(ov.MerchantAccountID, envOr(env, EnvMerchantAccountID), base.MerchantAccountID)
	eff.AuthMethod = firstNonEmpty(envOr(env, EnvAuthMethod), base.AuthMethod, AuthKey)
	eff.AuthHost = firstNonEmpty(envOr(env, EnvAuthHost), base.AuthHost)
	eff.Email = firstNonEmpty(envOr(env, EnvEmail), base.Email)

	if eff.Environment != EnvSandbox && eff.Environment != EnvProduction {
		return Resolved{}, fmt.Errorf("invalid environment %q: must be %q or %q", eff.Environment, EnvSandbox, EnvProduction)
	}
	if eff.AuthMethod != AuthKey && eff.AuthMethod != AuthLogin {
		return Resolved{}, fmt.Errorf("invalid auth_method %q: must be %q or %q", eff.AuthMethod, AuthKey, AuthLogin)
	}

	return Resolved{
		ProfileName: name,
		Profile:     eff,
		Token:       firstNonEmpty(ov.Token, envOr(env, EnvToken)),
	}, nil
}

func envOr(env EnvLookup, key string) string {
	if v, ok := env(key); ok {
		return v
	}
	return ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
