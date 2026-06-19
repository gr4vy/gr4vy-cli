// Package app is the per-invocation runtime shared by hand-written and
// generated commands. It resolves configuration from the persistent flags,
// builds an authenticated gr4vy-go client, and renders results. It depends only
// on lower-level packages so both the cli and generated command packages can
// import it without cycles.
package app

import (
	"io"
	"os"
	"time"

	gr4vygo "github.com/gr4vy/gr4vy-go"
	"github.com/spf13/cobra"

	"github.com/gr4vy/gr4vy-cli/internal/auth"
	"github.com/gr4vy/gr4vy-cli/internal/clierr"
	"github.com/gr4vy/gr4vy-cli/internal/config"
	"github.com/gr4vy/gr4vy-cli/internal/output"
	"github.com/gr4vy/gr4vy-cli/internal/secret"
)

// EnvOutput overrides the default output format.
const EnvOutput = "GR4VY_OUTPUT"

// Settings is the effective configuration for a single command invocation.
type Settings struct {
	Resolved   config.Resolved
	Config     *config.Config
	ConfigPath string
	Store      secret.Store
	Output     output.Format
	Compact    bool
	Timeout    time.Duration
	Debug      bool
}

// ConfigPath resolves the config file path from the --config flag, the
// GR4VY_CONFIG env var, or the XDG default.
func ConfigPath(cmd *cobra.Command) (string, error) {
	if p, _ := cmd.Flags().GetString("config"); p != "" {
		return p, nil
	}
	return config.DefaultPath()
}

// LoadConfig loads the config file selected for this invocation, returning the
// config and its path. Profile-management commands use this directly.
func LoadConfig(cmd *cobra.Command) (*config.Config, string, error) {
	path, err := ConfigPath(cmd)
	if err != nil {
		return nil, "", clierr.Config(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, "", clierr.Config(err)
	}
	return cfg, path, nil
}

// Resolve builds the Settings for an API command from the persistent flags,
// config, and environment.
func Resolve(cmd *cobra.Command) (*Settings, error) {
	cfg, path, err := LoadConfig(cmd)
	if err != nil {
		return nil, err
	}
	fl := cmd.Flags()
	getStr := func(name string) string { v, _ := fl.GetString(name); return v }

	ov := config.Overrides{
		Profile:           getStr("profile"),
		ID:                getStr("id"),
		Environment:       getStr("server"),
		MerchantAccountID: getStr("merchant-account-id"),
		Token:             getStr("token"),
	}
	r, err := config.Resolve(cfg, ov, config.OSEnv)
	if err != nil {
		return nil, clierr.Config(err)
	}

	store, err := secret.Open("")
	if err != nil {
		return nil, clierr.Config(err)
	}

	fmtStr := getStr("output")
	if fmtStr == "" {
		fmtStr = os.Getenv(EnvOutput)
	}
	of, err := output.Parse(fmtStr, isTTY(os.Stdout))
	if err != nil {
		return nil, clierr.Usage(err)
	}

	timeout, _ := fl.GetDuration("timeout")
	debug, _ := fl.GetBool("debug")
	compact, _ := fl.GetBool("compact")

	return &Settings{
		Resolved:   r,
		Config:     cfg,
		ConfigPath: path,
		Store:      store,
		Output:     of,
		Compact:    compact,
		Timeout:    timeout,
		Debug:      debug,
	}, nil
}

// Client builds an authenticated gr4vy-go client for these settings.
func (s *Settings) Client() (*gr4vygo.Gr4vy, error) {
	prov, err := auth.BuildProvider(s.Resolved, s.Store, config.OSEnv)
	if err != nil {
		return nil, clierr.Config(err)
	}
	client, err := auth.NewClient(s.Resolved, prov, s.Timeout)
	if err != nil {
		return nil, clierr.Config(err)
	}
	return client, nil
}

// Render writes a command result in the configured output format.
func (s *Settings) Render(w io.Writer, v any) error {
	return output.Render(w, v, s.Output, output.Options{Compact: s.Compact})
}

func isTTY(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
