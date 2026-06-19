// Package commands holds the registry of generated API commands and the
// generic builder that turns them into the cobra command tree. The generated
// package (commands/generated) registers one Operation per SDK operation, each
// carrying metadata (for the command tree) and a typed closure (for the call).
package commands

import (
	"context"

	gr4vygo "github.com/gr4vy/gr4vy-go"
	"github.com/spf13/pflag"
)

// FlagKind is the scalar type of a generated flag.
type FlagKind int

const (
	KindString FlagKind = iota
	KindInt
	KindBool
	KindFloat
	KindStringSlice
)

// Flag describes a generated command-line flag.
type Flag struct {
	Name  string
	Usage string
	Kind  FlagKind
}

// Operation is a single generated API command.
type Operation struct {
	Group      string   // dotted kebab resource path, e.g. "transactions.refunds"
	Name       string   // verb, e.g. "create"
	Short      string   // one-line help
	Long       string   // full help
	PathParams []string // positional arg names, in path order
	HasBody    bool     // accepts a request body (--data)
	BodyType   string   // body component type name (for help)
	IsList     bool     // renders a paginated collection
	Optionals  []Flag   // optional scalar flags (idempotency-key, x-forwarded-for, query params)
	Run        RunFunc  // typed call into the SDK
}

// Inputs are the resolved inputs passed to an operation's Run closure.
type Inputs struct {
	Args  []string
	Body  []byte
	Flags *pflag.FlagSet
}

// RunFunc executes the typed SDK call and returns the value to render (or nil).
type RunFunc func(ctx context.Context, c *gr4vygo.Gr4vy, in Inputs) (any, error)

var registry []*Operation

// Register adds an operation to the registry. Called from generated init().
func Register(op *Operation) { registry = append(registry, op) }

// All returns every registered operation.
func All() []*Operation { return registry }

// --- helpers used by generated closures to read optional flags ---

// OptString returns a pointer to the flag value if it was set, else nil.
func OptString(fs *pflag.FlagSet, name string) *string {
	if fs == nil || !fs.Changed(name) {
		return nil
	}
	v, err := fs.GetString(name)
	if err != nil {
		return nil
	}
	return &v
}

// OptInt64 returns a pointer to the int flag value if set, else nil.
func OptInt64(fs *pflag.FlagSet, name string) *int64 {
	if fs == nil || !fs.Changed(name) {
		return nil
	}
	v, err := fs.GetInt64(name)
	if err != nil {
		return nil
	}
	return &v
}

// OptBool returns a pointer to the bool flag value if set, else nil.
func OptBool(fs *pflag.FlagSet, name string) *bool {
	if fs == nil || !fs.Changed(name) {
		return nil
	}
	v, err := fs.GetBool(name)
	if err != nil {
		return nil
	}
	return &v
}

// OptFloat64 returns a pointer to the float flag value if set, else nil.
func OptFloat64(fs *pflag.FlagSet, name string) *float64 {
	if fs == nil || !fs.Changed(name) {
		return nil
	}
	v, err := fs.GetFloat64(name)
	if err != nil {
		return nil
	}
	return &v
}

// StringSlice returns the string-slice flag value (nil if unset).
func StringSlice(fs *pflag.FlagSet, name string) []string {
	if fs == nil || !fs.Changed(name) {
		return nil
	}
	v, err := fs.GetStringSlice(name)
	if err != nil {
		return nil
	}
	return v
}
