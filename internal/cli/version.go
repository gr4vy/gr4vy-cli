package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "gr4vy %s\n", build.Version)
			fmt.Fprintf(cmd.OutOrStdout(), "  commit:  %s\n", build.Commit)
			fmt.Fprintf(cmd.OutOrStdout(), "  built:   %s\n", build.Date)
			fmt.Fprintf(cmd.OutOrStdout(), "  go:      %s\n", runtime.Version())
			fmt.Fprintf(cmd.OutOrStdout(), "  platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
			return nil
		},
	}
}
