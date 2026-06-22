// Package output renders command results in JSON, YAML, or a human-readable
// table. Values are normalized through JSON first, so typed gr4vy-go responses
// and generic maps render identically.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// Format is an output encoding.
type Format string

const (
	JSON  Format = "json"
	YAML  Format = "yaml"
	Table Format = "table"
)

// Parse resolves a format string, defaulting to table on a TTY and json
// otherwise. An unknown value is an error.
func Parse(s string, isTTY bool) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "":
		if isTTY {
			return Table, nil
		}
		return JSON, nil
	case "json":
		return JSON, nil
	case "yaml", "yml":
		return YAML, nil
	case "table":
		return Table, nil
	default:
		return "", fmt.Errorf("unknown output format %q: use json, yaml, or table", s)
	}
}

// Options tune rendering.
type Options struct {
	Compact bool // single-line JSON (jq-friendly)
}

// Render writes v to w in the given format.
func Render(w io.Writer, v any, format Format, opts Options) error {
	norm, err := normalize(v)
	if err != nil {
		return err
	}
	switch format {
	case JSON:
		return renderJSON(w, norm, opts.Compact)
	case YAML:
		return renderYAML(w, norm)
	case Table:
		return renderTable(w, norm)
	default:
		return renderJSON(w, norm, opts.Compact)
	}
}

// normalize round-trips a value through JSON to produce generic maps/slices.
func normalize(v any) (any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || string(data) == "null" {
		return nil, nil
	}
	var out any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func renderJSON(w io.Writer, v any, compact bool) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if !compact {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(v)
}

func renderYAML(w io.Writer, v any) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	defer enc.Close()
	return enc.Encode(v)
}

// renderTable prints a list as rows or a single object as key/value pairs.
func renderTable(w io.Writer, v any) error {
	switch t := v.(type) {
	case nil:
		return nil
	case map[string]any:
		// Paginated list responses wrap rows under "items".
		if items, ok := t["items"].([]any); ok {
			return renderRows(w, items)
		}
		return renderKV(w, t)
	case []any:
		return renderRows(w, t)
	default:
		_, err := fmt.Fprintln(w, scalar(t))
		return err
	}
}

// preferredCols are surfaced first (in this order) when present.
var preferredCols = []string{"id", "type", "status", "method", "amount", "currency", "name", "display_name", "created_at", "updated_at"}

func renderRows(w io.Writer, items []any) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "(no results)")
		return err
	}
	cols := columnsFor(items)
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, strings.ToUpper(strings.Join(cols, "\t")))
	for _, it := range items {
		row, _ := it.(map[string]any)
		cells := make([]string, len(cols))
		for i, c := range cols {
			cells[i] = scalar(row[c])
		}
		fmt.Fprintln(tw, strings.Join(cells, "\t"))
	}
	return tw.Flush()
}

// columnsFor picks columns: preferred keys first, then remaining scalar keys.
func columnsFor(items []any) []string {
	present := map[string]bool{}
	for _, it := range items {
		row, ok := it.(map[string]any)
		if !ok {
			continue
		}
		for k, val := range row {
			if isScalar(val) {
				present[k] = true
			}
		}
	}
	var cols []string
	for _, c := range preferredCols {
		if present[c] {
			cols = append(cols, c)
			delete(present, c)
		}
	}
	rest := make([]string, 0, len(present))
	for k := range present {
		rest = append(rest, k)
	}
	sort.Strings(rest)
	return append(cols, rest...)
}

func renderKV(w io.Writer, m map[string]any) error {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	for _, k := range keys {
		fmt.Fprintf(tw, "%s\t%s\n", k, scalar(m[k]))
	}
	return tw.Flush()
}

func isScalar(v any) bool {
	switch v.(type) {
	case map[string]any, []any:
		return false
	default:
		return true
	}
}

// scalar renders a value as a single cell, collapsing nested structures.
func scalar(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case map[string]any:
		return "{…}"
	case []any:
		return fmt.Sprintf("[%d]", len(t))
	case float64:
		// JSON numbers decode as float64; render integers without a decimal.
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%v", t)
	case string:
		return t
	default:
		return fmt.Sprintf("%v", t)
	}
}
