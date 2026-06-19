package cli

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/gr4vy/gr4vy-cli/internal/app"
	"github.com/gr4vy/gr4vy-cli/internal/auth"
	"github.com/gr4vy/gr4vy-cli/internal/clierr"
	"github.com/gr4vy/gr4vy-cli/internal/config"
	"github.com/gr4vy/gr4vy-cli/internal/secret"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Aliases: []string{"profile"},
		Short:   "Manage configuration profiles",
	}
	cmd.AddCommand(
		newConfigAddCmd(),
		newConfigListCmd(),
		newConfigShowCmd(),
		newConfigKeyCmd(),
		newConfigUseCmd(),
		newConfigRemoveCmd(),
		newConfigPathCmd(),
		newConfigImportCmd(),
	)
	return cmd
}

func newConfigKeyCmd() *cobra.Command {
	var b64 bool
	cmd := &cobra.Command{
		Use:   "key",
		Short: "Print the active profile's private key (PEM)",
		Long: "Output the private key the selected profile would use, resolved from the " +
			"secret store, a key file, or the environment. Select a profile with --profile. " +
			"Handle the output with care — it is your secret signing key.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := app.Resolve(cmd)
			if err != nil {
				return err
			}
			if s.Resolved.Profile.AuthMethod != config.AuthKey {
				return clierr.Usage(fmt.Errorf("profile %q uses %s auth and has no private key",
					s.Resolved.ProfileName, s.Resolved.Profile.AuthMethod))
			}
			pem, err := auth.ResolveKeyPEM(s.Resolved, s.Store, config.OSEnv)
			if err != nil {
				return clierr.Config(err)
			}
			if b64 {
				fmt.Fprintln(cmd.OutOrStdout(), base64.StdEncoding.EncodeToString([]byte(pem)))
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), strings.TrimRight(pem, "\n"))
			return nil
		},
	}
	cmd.Flags().BoolVar(&b64, "base64", false, "output the key base64-encoded (single line)")
	return cmd
}

// profileFlags holds the editable fields of a profile, shared by `config add`
// and `init`.
type profileFlags struct {
	id          string
	environment string
	mid         string
	authMethod  string
	keyFile     string
	keyStdin    bool
	keyPath     string
	keyEnv      string
	email       string
	authHost    string
	scopes      []string
	tokenTTL    string
	setActive   bool
}

func (pf *profileFlags) register(cmd *cobra.Command) {
	f := cmd.Flags()
	f.StringVar(&pf.id, "id", "", "Gr4vy instance id")
	f.StringVar(&pf.environment, "environment", "", "sandbox or production")
	f.StringVar(&pf.mid, "merchant-account-id", "", "default merchant account id")
	f.StringVar(&pf.authMethod, "auth-method", "", "key or login (default key)")
	f.StringVar(&pf.keyFile, "key-file", "", "path to a PEM private key to import into the secret store")
	f.BoolVar(&pf.keyStdin, "key-stdin", false, "read the PEM private key from stdin")
	f.StringVar(&pf.keyPath, "key-path", "", "reference a PEM file in place (not copied into the store)")
	f.StringVar(&pf.keyEnv, "key-env", "", "name of an env var holding the PEM private key")
	f.StringVar(&pf.email, "email", "", "login email (for auth-method=login)")
	f.StringVar(&pf.authHost, "auth-host", "", "override the auth host for login")
	f.StringSliceVar(&pf.scopes, "default-scope", nil, "default token scope (repeatable)")
	f.StringVar(&pf.tokenTTL, "token-ttl", "", "default token lifetime, e.g. 1h")
	f.BoolVar(&pf.setActive, "set-active", false, "make this the active profile")
}

func newConfigAddCmd() *cobra.Command {
	pf := &profileFlags{}
	noInput := false
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add or update a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddProfile(cmd, args[0], pf, !noInput && isInteractive())
		},
	}
	pf.register(cmd)
	cmd.Flags().BoolVar(&noInput, "no-input", false, "never prompt; fail if a required field is missing")
	return cmd
}

// runAddProfile builds a profile from flags (prompting when interactive),
// stores any provided key, and writes it to the config.
func runAddProfile(cmd *cobra.Command, name string, pf *profileFlags, interactive bool) error {
	cfg, path, err := app.LoadConfig(cmd)
	if err != nil {
		return err
	}
	store, err := secret.Open("")
	if err != nil {
		return clierr.Config(err)
	}

	p := config.Profile{}
	if existing, ok := cfg.Lookup(name); ok {
		p = existing
	}

	if err := fillProfile(&p, pf, interactive); err != nil {
		return err
	}

	// Resolve and store the private key for key auth.
	if p.AuthMethod == config.AuthKey {
		if err := storeKeyForProfile(name, &p, pf, store, interactive); err != nil {
			return err
		}
	}

	cfg.Set(name, p)
	if pf.setActive || cfg.ActiveProfile == "" || len(cfg.Profiles) == 1 {
		cfg.ActiveProfile = name
	}
	if err := config.Save(path, cfg); err != nil {
		return clierr.Config(err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "saved profile %q (%s)\n", name, path)
	return nil
}

func fillProfile(p *config.Profile, pf *profileFlags, interactive bool) error {
	if pf.id != "" {
		p.ID = pf.id
	}
	if pf.environment != "" {
		p.Environment = pf.environment
	}
	if pf.mid != "" {
		p.MerchantAccountID = pf.mid
	}
	if pf.authMethod != "" {
		p.AuthMethod = pf.authMethod
	}
	if pf.email != "" {
		p.Email = pf.email
	}
	if pf.authHost != "" {
		p.AuthHost = pf.authHost
	}
	if len(pf.scopes) > 0 {
		p.DefaultScopes = pf.scopes
	}
	if pf.tokenTTL != "" {
		p.TokenTTL = pf.tokenTTL
	}

	if interactive {
		var err error
		if p.ID, err = promptLineDefault("Instance id", p.ID); err != nil {
			return err
		}
		if p.Environment == "" {
			p.Environment = config.EnvSandbox
		}
		if p.Environment, err = promptLineDefault("Environment (sandbox/production)", p.Environment); err != nil {
			return err
		}
		if p.AuthMethod == "" {
			p.AuthMethod = config.AuthKey
		}
		if p.AuthMethod, err = promptLineDefault("Auth method (key/login)", p.AuthMethod); err != nil {
			return err
		}
		if p.MerchantAccountID, err = promptLineDefault("Default merchant account id", firstNonEmptyStr(p.MerchantAccountID, "default")); err != nil {
			return err
		}
	}

	// Defaults.
	if p.Environment == "" {
		p.Environment = config.EnvSandbox
	}
	if p.AuthMethod == "" {
		p.AuthMethod = config.AuthKey
	}
	if p.ID == "" {
		return clierr.Usage(fmt.Errorf("instance id is required (use --id)"))
	}
	if p.Environment != config.EnvSandbox && p.Environment != config.EnvProduction {
		return clierr.Usage(fmt.Errorf("environment must be sandbox or production"))
	}
	if p.AuthMethod == config.AuthLogin && p.Email == "" && interactive {
		v, err := promptLineDefault("Login email", p.Email)
		if err != nil {
			return err
		}
		p.Email = v
	}
	return nil
}

// storeKeyForProfile resolves the private key from the flags (or prompts for a
// path), validates it, and records the appropriate key reference on the profile.
func storeKeyForProfile(name string, p *config.Profile, pf *profileFlags, store secret.Store, interactive bool) error {
	switch {
	case pf.keyEnv != "":
		p.KeyRef = config.KeyRefEnv
		p.KeyEnv = pf.keyEnv
		return nil
	case pf.keyPath != "":
		pem, err := os.ReadFile(pf.keyPath)
		if err != nil {
			return clierr.Config(fmt.Errorf("read key file: %w", err))
		}
		if err := secret.ValidatePrivateKeyPEM(string(pem)); err != nil {
			return clierr.Config(err)
		}
		p.KeyRef = config.KeyRefFile
		p.KeyPath = pf.keyPath
		return nil
	}

	var pem string
	switch {
	case pf.keyStdin:
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return clierr.Config(err)
		}
		pem = string(data)
	case pf.keyFile != "":
		data, err := os.ReadFile(pf.keyFile)
		if err != nil {
			return clierr.Config(fmt.Errorf("read key file: %w", err))
		}
		pem = string(data)
	case interactive:
		in, err := promptLine("Path to PEM file, or a base64-encoded key: ")
		if err != nil {
			return err
		}
		if in != "" {
			if fi, statErr := os.Stat(in); statErr == nil && !fi.IsDir() {
				data, rerr := os.ReadFile(in)
				if rerr != nil {
					return clierr.Config(fmt.Errorf("read key file: %w", rerr))
				}
				pem = string(data)
			} else {
				pem = in // treat the input as inline key material (base64 or PEM)
			}
		}
	}

	if pem == "" {
		// Allow a keyless profile to be created; the key can be added later.
		if p.KeyRef == "" {
			p.KeyRef = config.KeyRefStore
		}
		return nil
	}
	if err := secret.ValidatePrivateKeyPEM(pem); err != nil {
		return clierr.Config(err)
	}
	if err := store.Set(name, secret.KindKey, secret.NormalizePEM(pem)); err != nil {
		return clierr.Config(err)
	}
	p.KeyRef = config.KeyRefStore
	if store.Backend() == secret.BackendFile {
		fmt.Fprintln(os.Stderr, "note: no OS keychain available; private key stored in a 0600 file under ~/.config/gr4vy/secrets")
	}
	return nil
}

func newConfigListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured profiles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, _, err := app.LoadConfig(cmd)
			if err != nil {
				return err
			}
			if len(cfg.Profiles) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no profiles configured; run `gr4vy init`")
				return nil
			}
			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "ACTIVE\tNAME\tID\tENVIRONMENT\tAUTH")
			for _, name := range cfg.Names() {
				p := cfg.Profiles[name]
				active := ""
				if name == cfg.ActiveProfile {
					active = "*"
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", active, name, p.ID, p.Environment, p.AuthMethod)
			}
			return tw.Flush()
		},
	}
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [name]",
		Short: "Show a profile (secrets redacted)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := app.LoadConfig(cmd)
			if err != nil {
				return err
			}
			name := cfg.ActiveProfile
			if len(args) == 1 {
				name = args[0]
			}
			p, ok := cfg.Lookup(name)
			if !ok {
				return clierr.Usage(fmt.Errorf("no such profile %q", name))
			}
			kv := map[string]any{
				"name":                name,
				"id":                  p.ID,
				"environment":         p.Environment,
				"merchant_account_id": p.MerchantAccountID,
				"auth_method":         p.AuthMethod,
				"key_ref":             p.KeyRef,
				"key_path":            p.KeyPath,
				"key_env":             p.KeyEnv,
				"email":               p.Email,
				"auth_host":           p.AuthHost,
				"default_scopes":      p.DefaultScopes,
				"token_ttl":           p.TokenTTL,
			}
			keys := make([]string, 0, len(kv))
			for k := range kv {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			for _, k := range keys {
				fmt.Fprintf(tw, "%s\t%v\n", k, kv[k])
			}
			return tw.Flush()
		},
	}
}

func newConfigUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <name>",
		Short: "Set the active profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := app.LoadConfig(cmd)
			if err != nil {
				return err
			}
			if _, ok := cfg.Lookup(args[0]); !ok {
				return clierr.Usage(fmt.Errorf("no such profile %q", args[0]))
			}
			cfg.ActiveProfile = args[0]
			if err := config.Save(path, cfg); err != nil {
				return clierr.Config(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "active profile set to %q\n", args[0])
			return nil
		},
	}
}

func newConfigRemoveCmd() *cobra.Command {
	var keepSecrets, yes bool
	cmd := &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm", "delete"},
		Short:   "Remove a profile and its secrets",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := app.LoadConfig(cmd)
			if err != nil {
				return err
			}
			name := args[0]
			if _, ok := cfg.Lookup(name); !ok {
				return clierr.Usage(fmt.Errorf("no such profile %q", name))
			}
			if !yes && isInteractive() && !confirm(fmt.Sprintf("Remove profile %q?", name), false) {
				return nil
			}
			cfg.Remove(name)
			if err := config.Save(path, cfg); err != nil {
				return clierr.Config(err)
			}
			if !keepSecrets {
				if store, serr := secret.Open(""); serr == nil {
					_ = store.Delete(name, secret.KindKey)
					_ = store.Delete(name, secret.KindLogin)
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed profile %q\n", name)
			return nil
		},
	}
	cmd.Flags().BoolVar(&keepSecrets, "keep-secrets", false, "do not delete stored secrets")
	cmd.Flags().BoolVar(&yes, "yes", false, "do not prompt for confirmation")
	return cmd
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the resolved config path and secret backend",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := app.ConfigPath(cmd)
			if err != nil {
				return clierr.Config(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "config: %s\n", path)
			if store, serr := secret.Open(""); serr == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "secret backend: %s\n", store.Backend())
			}
			return nil
		},
	}
}

func firstNonEmptyStr(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
