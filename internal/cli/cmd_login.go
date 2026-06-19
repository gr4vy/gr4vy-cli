package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gr4vy/gr4vy-cli/internal/app"
	"github.com/gr4vy/gr4vy-cli/internal/auth"
	"github.com/gr4vy/gr4vy-cli/internal/clierr"
)

func newLoginCmd() *cobra.Command {
	var email string
	var passwordStdin bool
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in with email and password",
		Long: "Authenticate against the Gr4vy session endpoint and store the resulting " +
			"tokens for the active profile. The CLI refreshes the access token automatically.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := app.Resolve(cmd)
			if err != nil {
				return err
			}

			if email == "" {
				email = s.Resolved.Profile.Email
			}
			if email == "" && isInteractive() {
				if email, err = promptLine("Email: "); err != nil {
					return err
				}
			}
			if email == "" {
				return clierr.Usage(fmt.Errorf("email is required (use --email or set it on the profile)"))
			}

			password := os.Getenv(auth.EnvPassword)
			if password == "" {
				switch {
				case passwordStdin:
					if password, err = promptLine(""); err != nil {
						return err
					}
				case isInteractive():
					if password, err = promptSecret("Password: "); err != nil {
						return err
					}
				}
			}
			if password == "" {
				return clierr.Usage(fmt.Errorf("password is required (use --password-stdin, %s, or run interactively)", auth.EnvPassword))
			}

			if err := auth.Login(cmd.Context(), s.Resolved, s.Store, email, password, nil); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "logged in as %s (profile %q)\n", email, s.Resolved.ProfileName)
			return nil
		},
	}
	cmd.Flags().StringVar(&email, "email", "", "login email")
	cmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "read the password from stdin")
	return cmd
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Log out and clear stored session tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := app.Resolve(cmd)
			if err != nil {
				return err
			}
			if err := auth.Logout(cmd.Context(), s.Resolved, s.Store, nil); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "logged out (profile %q)\n", s.Resolved.ProfileName)
			return nil
		},
	}
}
