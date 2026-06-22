// Package clierr defines stable process exit codes and maps command errors —
// including gr4vy-go API errors — onto them. It lives in its own package so
// both the runtime and generated commands can classify errors without import
// cycles.
package clierr

import "errors"

// Stable process exit codes. Documented; scripts may depend on them.
const (
	ExitOK        = 0
	ExitGeneric   = 1
	ExitUsage     = 2 // invalid flags/args
	ExitConfig    = 3 // configuration/auth problem, before any request
	ExitAPIClient = 4 // API responded 4xx (other than 429)
	ExitAPIServer = 5 // API responded 5xx
	ExitRateLimit = 6 // API responded 429
)

// ConfigError marks a failure resolving configuration, profiles, credentials,
// or auth — before any API request was attempted.
type ConfigError struct{ Err error }

func (e *ConfigError) Error() string { return e.Err.Error() }
func (e *ConfigError) Unwrap() error { return e.Err }

// Config wraps err as a ConfigError (no-op for nil).
func Config(err error) error {
	if err == nil {
		return nil
	}
	return &ConfigError{Err: err}
}

// UsageError marks invalid user input.
type UsageError struct{ Err error }

func (e *UsageError) Error() string { return e.Err.Error() }
func (e *UsageError) Unwrap() error { return e.Err }

// Usage wraps err as a UsageError (no-op for nil).
func Usage(err error) error {
	if err == nil {
		return nil
	}
	return &UsageError{Err: err}
}

// ExitCodeFor maps an error to a process exit code.
func ExitCodeFor(err error) int {
	if err == nil {
		return ExitOK
	}
	var cfgErr *ConfigError
	if errors.As(err, &cfgErr) {
		return ExitConfig
	}
	var usageErr *UsageError
	if errors.As(err, &usageErr) {
		return ExitUsage
	}
	return apiExitCode(err)
}
