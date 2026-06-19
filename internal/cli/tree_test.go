package cli

import (
	"flag"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/gr4vy/gr4vy-cli/internal/commands"
)

var update = flag.Bool("update", false, "update golden files")

// walkLeaves returns one line per runnable command: its full path followed by
// its own (non-inherited) flags, sorted.
func walkLeaves(root *cobra.Command) []string {
	var lines []string
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		if c.Runnable() && !c.HasAvailableSubCommands() {
			var flags []string
			c.LocalNonPersistentFlags().VisitAll(func(f *pflag.Flag) {
				flags = append(flags, "--"+f.Name)
			})
			sort.Strings(flags)
			line := c.CommandPath()
			if len(flags) > 0 {
				line += "  " + strings.Join(flags, " ")
			}
			lines = append(lines, line)
		}
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(root)
	sort.Strings(lines)
	return lines
}

// TestCommandTreeSnapshot guards the full command surface against unexpected
// drift. When the spec legitimately changes, regenerate and run with -update.
func TestCommandTreeSnapshot(t *testing.T) {
	root := NewRootCmd()
	got := strings.Join(walkLeaves(root), "\n") + "\n"

	golden := filepath.Join("testdata", "command_tree.golden")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden (run with -update first): %v", err)
	}
	if got != string(want) {
		t.Errorf("command tree drifted from golden.\nRegenerate (go generate ./...) and, if intended, run: go test ./internal/cli -run TestCommandTreeSnapshot -update\n\n--- got ---\n%s", got)
	}
}

// TestGeneratedLeavesMatchRegistry asserts every registered operation produced
// exactly one runnable leaf command.
func TestGeneratedLeavesMatchRegistry(t *testing.T) {
	root := NewRootCmd()
	leaves := walkLeaves(root)

	// Hand-written runnable leaves (version, login, token, embed, config *, ...)
	// plus generated ones. Generated count must equal the registry size.
	var generatedLeaves int
	registered := map[string]bool{}
	for _, op := range commands.All() {
		path := "gr4vy " + strings.ReplaceAll(op.Group, ".", " ") + " " + op.Name
		registered[path] = true
	}
	for _, line := range leaves {
		path := line
		if i := strings.Index(line, "  "); i >= 0 {
			path = line[:i]
		}
		if registered[path] {
			generatedLeaves++
		}
	}
	if generatedLeaves != len(commands.All()) {
		t.Errorf("matched %d generated leaves in the tree, want %d", generatedLeaves, len(commands.All()))
	}
}
