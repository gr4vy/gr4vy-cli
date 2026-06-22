package cli

import (
	"encoding/json"
	"fmt"

	gr4vygo "github.com/gr4vy/gr4vy-go"
	"github.com/spf13/cobra"

	"github.com/gr4vy/gr4vy-cli/internal/app"
	"github.com/gr4vy/gr4vy-cli/internal/auth"
	"github.com/gr4vy/gr4vy-cli/internal/clierr"
	"github.com/gr4vy/gr4vy-cli/internal/config"
)

func newTokenCmd() *cobra.Command {
	var (
		scopes     []string
		expiresIn  string
		listScopes bool
	)
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Generate a server-to-server API access token (JWT)",
		Long: "Generate a signed bearer token for the Gr4vy API using the profile's " +
			"private key. Scopes default to the profile's default_scopes (or *.read and " +
			"*.write).",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if listScopes {
				for _, s := range auth.ScopeStrings() {
					fmt.Fprintln(cmd.OutOrStdout(), s)
				}
				return nil
			}

			s, err := app.Resolve(cmd)
			if err != nil {
				return err
			}
			pem, err := auth.ResolveKeyPEM(s.Resolved, s.Store, config.OSEnv)
			if err != nil {
				return clierr.Config(err)
			}

			rawScopes := scopes
			if len(rawScopes) == 0 {
				rawScopes = s.Resolved.Profile.DefaultScopes
			}
			if len(rawScopes) == 0 {
				// Enforce the documented fallback (see Long help) rather than
				// leaving scopes empty and depending on how gr4vy-go treats it.
				rawScopes = []string{"*.read", "*.write"}
			}
			scopeList, err := auth.ParseScopes(rawScopes)
			if err != nil {
				return clierr.Usage(err)
			}

			ttl := expiresIn
			if ttl == "" {
				ttl = s.Resolved.Profile.TokenTTL
			}
			if ttl == "" {
				ttl = "1h"
			}
			seconds, err := parseTTLSeconds(ttl)
			if err != nil {
				return clierr.Usage(err)
			}

			token, err := gr4vygo.GetToken(pem, scopeList, seconds)
			if err != nil {
				return clierr.Config(err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), token)

			if debug, _ := cmd.Flags().GetBool("debug"); debug {
				printDecoded(cmd, token)
			}
			return nil
		},
	}
	cmd.Flags().StringSliceVarP(&scopes, "scope", "s", nil, "scope to include (repeatable); see --list-scopes")
	cmd.Flags().StringVarP(&expiresIn, "expires-in", "e", "", "token lifetime, e.g. 1h, 30m, 10d, 3600 (default 1h)")
	cmd.Flags().BoolVar(&listScopes, "list-scopes", false, "list all valid scopes and exit")
	_ = cmd.RegisterFlagCompletionFunc("scope", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return auth.ScopeStrings(), cobra.ShellCompDirectiveNoFileComp
	})
	return cmd
}

// printDecoded writes a JWT's decoded header and claims to stderr.
func printDecoded(cmd *cobra.Command, token string) {
	dec, err := decodeJWT(token)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), "debug: could not decode token:", err)
		return
	}
	out, _ := json.MarshalIndent(dec, "", "  ")
	fmt.Fprintln(cmd.ErrOrStderr(), string(out))
}
