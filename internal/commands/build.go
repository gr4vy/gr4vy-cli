package commands

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gr4vy/gr4vy-cli/internal/app"
)

// Build attaches every registered API command to root, creating intermediate
// resource (parent) commands as needed.
func Build(root *cobra.Command) {
	parents := map[string]*cobra.Command{"": root}

	// Stable order so the tree (and golden output) is deterministic.
	ops := append([]*Operation(nil), All()...)
	sort.Slice(ops, func(i, j int) bool {
		if ops[i].Group != ops[j].Group {
			return ops[i].Group < ops[j].Group
		}
		return ops[i].Name < ops[j].Name
	})

	for _, op := range ops {
		parent := ensureParents(parents, op.Group)
		parent.AddCommand(leafCommand(op))
	}
}

// ensureParents returns the command for a dotted group path, creating each
// missing segment.
func ensureParents(parents map[string]*cobra.Command, group string) *cobra.Command {
	if group == "" {
		return parents[""]
	}
	segs := strings.Split(group, ".")
	path := ""
	parent := parents[""]
	for _, seg := range segs {
		if path == "" {
			path = seg
		} else {
			path += "." + seg
		}
		if c, ok := parents[path]; ok {
			parent = c
			continue
		}
		c := &cobra.Command{
			Use:   seg,
			Short: "Manage " + strings.ReplaceAll(path, ".", " "),
		}
		parent.AddCommand(c)
		parents[path] = c
		parent = c
	}
	return parent
}

// leafCommand builds the cobra command for a single operation.
func leafCommand(op *Operation) *cobra.Command {
	use := op.Name
	for _, p := range op.PathParams {
		use += " <" + strings.ReplaceAll(p, "_", "-") + ">"
	}
	cmd := &cobra.Command{
		Use:   use,
		Short: op.Short,
		Long:  op.Long,
		Args:  cobra.ExactArgs(len(op.PathParams)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOperation(cmd, op, args)
		},
	}
	if op.HasBody {
		usage := "request body as JSON: inline, @file, or - for stdin"
		if op.BodyType != "" {
			usage += " (" + op.BodyType + ")"
		}
		cmd.Flags().String("data", "", usage)
	}
	for _, f := range op.Optionals {
		switch f.Kind {
		case KindInt:
			cmd.Flags().Int64(f.Name, 0, f.Usage)
		case KindBool:
			cmd.Flags().Bool(f.Name, false, f.Usage)
		case KindFloat:
			cmd.Flags().Float64(f.Name, 0, f.Usage)
		case KindStringSlice:
			cmd.Flags().StringSlice(f.Name, nil, f.Usage)
		default:
			cmd.Flags().String(f.Name, "", f.Usage)
		}
	}
	return cmd
}

func runOperation(cmd *cobra.Command, op *Operation, args []string) error {
	s, err := app.Resolve(cmd)
	if err != nil {
		return err
	}
	client, err := s.Client()
	if err != nil {
		return err
	}
	body, err := loadBody(cmd, op.HasBody)
	if err != nil {
		return err
	}
	res, err := op.Run(cmd.Context(), client, Inputs{Args: args, Body: body, Flags: cmd.Flags()})
	if err != nil {
		return err
	}
	if res == nil {
		fmt.Fprintln(cmd.ErrOrStderr(), "OK")
		return nil
	}
	return s.Render(cmd.OutOrStdout(), res)
}

// loadBody resolves the --data flag into raw JSON bytes. An unset --data yields
// a nil body, so the generated closures omit the request body entirely.
func loadBody(cmd *cobra.Command, hasBody bool) ([]byte, error) {
	if !hasBody {
		return nil, nil
	}
	v, _ := cmd.Flags().GetString("data")
	switch {
	case v == "":
		// No --data: omit the body entirely. Optional-body operations send
		// nothing; required-body ones get a clear server-side error. Send an
		// explicit empty object with --data '{}' if that's intended.
		return nil, nil
	case v == "-":
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("read body from stdin: %w", err)
		}
		return data, nil
	case strings.HasPrefix(v, "@"):
		data, err := os.ReadFile(v[1:])
		if err != nil {
			return nil, fmt.Errorf("read body file: %w", err)
		}
		return data, nil
	default:
		return []byte(v), nil
	}
}
