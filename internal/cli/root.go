// Package cli wires up the gr4vy command tree: the root command, its
// persistent flags, configuration/profile resolution, and the hand-written
// commands. Generated API commands are registered from internal/commands.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gr4vy/gr4vy-cli/internal/clierr"
	"github.com/gr4vy/gr4vy-cli/internal/commands"
	_ "github.com/gr4vy/gr4vy-cli/internal/commands/generated" // registers generated API commands
)

// BuildInfo carries version metadata stamped into the binary at build time.
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

// build holds the BuildInfo for the current process. It is set once by
// Execute and read by the version command.
var build BuildInfo

// Execute builds the root command and runs it, translating any error into a
// process exit code.
func Execute(bi BuildInfo) {
	build = bi
	root := NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error: "+clierr.FormatError(err))
		os.Exit(clierr.ExitCodeFor(err))
	}
}

// NewRootCmd constructs the root cobra command and attaches the persistent
// flags shared by every subcommand.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "gr4vy",
		Short:         "The Gr4vy CLI",
		Long:          "gr4vy is the command-line interface for the Gr4vy payment orchestration platform.",
		Version:       build.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	pf := cmd.PersistentFlags()
	pf.String("profile", "", "configuration profile to use (env: GR4VY_PROFILE)")
	pf.String("config", "", "path to the config file (env: GR4VY_CONFIG)")
	pf.String("id", "", "Gr4vy instance id used for the API host (env: GR4VY_ID)")
	pf.String("server", "", "server environment: sandbox|production (env: GR4VY_SERVER)")
	pf.String("merchant-account-id", "", "merchant account id (env: GR4VY_MERCHANT_ACCOUNT_ID)")
	pf.String("token", "", "pre-generated bearer token; skips JWT signing (env: GR4VY_TOKEN)")
	pf.StringP("output", "o", "", "output format: json|yaml|table (env: GR4VY_OUTPUT)")
	pf.Bool("compact", false, "compact single-line JSON output")
	pf.Duration("timeout", 0, "per-request timeout, e.g. 30s")
	pf.Bool("debug", false, "print debug information to stderr")

	cmd.AddCommand(
		newVersionCmd(),
		newInitCmd(),
		newConfigCmd(),
		newLoginCmd(),
		newLogoutCmd(),
		newTokenCmd(),
		newEmbedCmd(),
	)

	// Attach the generated API command tree (buyers, transactions, ...).
	commands.Build(cmd)

	return cmd
}
