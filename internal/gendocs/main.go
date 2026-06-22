// Command gendocs renders a Markdown reference for the whole gr4vy command tree
// into ./docs, using cobra's doc generator. Run via `go generate ./...` /
// `go run ./internal/gendocs`; it is never compiled into the shipped binary.
package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/gr4vy/gr4vy-cli/internal/cli"
)

const outDir = "docs"

func main() {
	root := cli.NewRootCmd()
	root.DisableAutoGenTag = true // deterministic output (no generation date)
	disableAutoGenTag(root)

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		log.Fatalf("gendocs: %v", err)
	}
	if err := doc.GenMarkdownTree(root, outDir); err != nil {
		log.Fatalf("gendocs: %v", err)
	}
}

// disableAutoGenTag turns off cobra's "Auto generated … on <date>" footer for
// every command so regeneration is deterministic (no date churn in diffs).
func disableAutoGenTag(c *cobra.Command) {
	c.DisableAutoGenTag = true
	for _, sub := range c.Commands() {
		disableAutoGenTag(sub)
	}
}
