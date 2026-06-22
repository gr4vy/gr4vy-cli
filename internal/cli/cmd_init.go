package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/gr4vy/gr4vy-cli/internal/app"
	"github.com/gr4vy/gr4vy-cli/internal/clierr"
	"github.com/gr4vy/gr4vy-cli/internal/config"
	"github.com/gr4vy/gr4vy-cli/internal/secret"
)

func newInitCmd() *cobra.Command {
	pf := &profileFlags{}
	noInput := false
	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Create your first profile interactively",
		Long:  "Bootstrap a configuration profile. Prompts interactively on a terminal; use flags (and --no-input) for scripted setup.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := "default"
			if len(args) == 1 {
				name = args[0]
			}
			return runAddProfile(cmd, name, pf, !noInput && isInteractive())
		},
	}
	pf.register(cmd)
	cmd.Flags().BoolVar(&noInput, "no-input", false, "never prompt; fail if a required field is missing")
	return cmd
}

// legacyConfig is the shape of the old ~/.gr4vyrc.json file.
type legacyConfig struct {
	Gr4vyID     string `json:"gr4vyId"`
	Environment string `json:"environment"`
	PrivateKey  string `json:"privateKey"` // inlined PEM contents
}

func newConfigImportCmd() *cobra.Command {
	var from, name string
	var setActive, deleteLegacy bool
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import the legacy ~/.gr4vyrc.json into a profile",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if name == "" {
				return clierr.Usage(fmt.Errorf("--name must not be empty"))
			}
			if from == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					return clierr.Config(err)
				}
				from = filepath.Join(home, ".gr4vyrc.json")
			}
			data, err := os.ReadFile(from)
			if err != nil {
				return clierr.Config(fmt.Errorf("read legacy config %s: %w", from, err))
			}
			var lc legacyConfig
			if err := json.Unmarshal(data, &lc); err != nil {
				return clierr.Config(fmt.Errorf("parse legacy config: %w", err))
			}
			if lc.Gr4vyID == "" {
				return clierr.Config(fmt.Errorf("legacy config has no gr4vyId"))
			}

			cfg, path, err := app.LoadConfig(cmd)
			if err != nil {
				return err
			}
			store, err := secret.Open("")
			if err != nil {
				return clierr.Config(err)
			}

			env := lc.Environment
			if env != config.EnvProduction {
				env = config.EnvSandbox
			}
			p := config.Profile{
				ID:                lc.Gr4vyID,
				Environment:       env,
				MerchantAccountID: "default",
				AuthMethod:        config.AuthKey,
				KeyRef:            config.KeyRefStore,
			}
			if lc.PrivateKey != "" {
				if err := secret.ValidatePrivateKeyPEM(lc.PrivateKey); err != nil {
					return clierr.Config(fmt.Errorf("legacy private key invalid: %w", err))
				}
				if err := store.Set(name, secret.KindKey, secret.NormalizePEM(lc.PrivateKey)); err != nil {
					return clierr.Config(err)
				}
			}
			cfg.Set(name, p)
			if setActive || cfg.ActiveProfile == "" {
				cfg.ActiveProfile = name
			}
			if err := config.Save(path, cfg); err != nil {
				return clierr.Config(err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "imported legacy config into profile %q (%s)\n", name, path)
			fmt.Fprintf(cmd.OutOrStdout(), "private key stored in the %s backend\n", store.Backend())
			if deleteLegacy {
				if err := os.Remove(from); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not delete %s: %v\n", from, err)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "deleted legacy file %s\n", from)
				}
			} else if lc.PrivateKey != "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "note: %s still contains a plaintext private key; remove it with --delete-legacy or manually\n", from)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "path to the legacy config (default ~/.gr4vyrc.json)")
	cmd.Flags().StringVar(&name, "name", "default", "profile name to create")
	cmd.Flags().BoolVar(&setActive, "set-active", false, "make the imported profile active")
	cmd.Flags().BoolVar(&deleteLegacy, "delete-legacy", false, "delete the legacy file after import")
	return cmd
}
