package cli

import (
	"fmt"
	"strconv"
	"strings"

	gr4vygo "github.com/gr4vy/gr4vy-go"
	"github.com/spf13/cobra"

	"github.com/gr4vy/gr4vy-cli/internal/app"
	"github.com/gr4vy/gr4vy-cli/internal/auth"
	"github.com/gr4vy/gr4vy-cli/internal/clierr"
	"github.com/gr4vy/gr4vy-cli/internal/config"
)

func newEmbedCmd() *cobra.Command {
	var (
		params         []string
		withCheckout   bool
		merchantAcctID string
	)
	cmd := &cobra.Command{
		Use:   "embed <amount> <currency> [key=value ...]",
		Short: "Generate an Embed token for the checkout form",
		Long: "Generate a JWT for Gr4vy Embed, pinning the given amount, currency, and any " +
			"extra key=value parameters. With --checkout-session, a checkout session is " +
			"created and its id is baked into the token.",
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			amount, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return clierr.Usage(fmt.Errorf("amount must be an integer in the smallest currency unit: %w", err))
			}
			embedParams := map[string]any{
				"amount":   amount,
				"currency": args[1],
			}
			kv, err := parseKVPairs(append(args[2:], params...))
			if err != nil {
				return clierr.Usage(err)
			}
			for k, v := range kv {
				embedParams[k] = v
			}

			s, err := app.Resolve(cmd)
			if err != nil {
				return err
			}
			pem, err := auth.ResolveKeyPEM(s.Resolved, s.Store, config.OSEnv)
			if err != nil {
				return clierr.Config(err)
			}

			var token string
			if withCheckout {
				client, err := s.Client()
				if err != nil {
					return err
				}
				var mid *string
				if merchantAcctID != "" {
					mid = &merchantAcctID
				}
				token, err = gr4vygo.GetEmbedTokenWithCheckoutSession(cmd.Context(), client, pem, embedParams, nil, mid)
				if err != nil {
					return err
				}
			} else {
				token, err = gr4vygo.GetEmbedToken(pem, embedParams, "")
				if err != nil {
					return clierr.Config(err)
				}
			}

			fmt.Fprintln(cmd.OutOrStdout(), token)
			if debug, _ := cmd.Flags().GetBool("debug"); debug {
				printDecoded(cmd, token)
			}
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&params, "param", nil, "extra embed parameter as key=value (repeatable)")
	cmd.Flags().BoolVar(&withCheckout, "checkout-session", false, "create a checkout session and pin its id into the token")
	cmd.Flags().StringVar(&merchantAcctID, "mid", "", "merchant account id for the checkout session (with --checkout-session)")
	return cmd
}

// parseKVPairs parses "key=value" strings into a map with string values.
func parseKVPairs(pairs []string) (map[string]any, error) {
	out := map[string]any{}
	for _, p := range pairs {
		k, v, ok := strings.Cut(p, "=")
		if !ok || k == "" {
			return nil, fmt.Errorf("invalid key=value pair: %q", p)
		}
		out[k] = v
	}
	return out, nil
}
