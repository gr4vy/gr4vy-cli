// Command gen reads the committed OpenAPI spec and the gr4vy-go SDK's Go types,
// and emits one typed cobra-backed command per SDK operation into
// internal/commands/generated. It is run via `go generate ./...` / `go run
// ./internal/gen` and is never compiled into the shipped binary.
//
// Source of truth:
//   - internal/spec/openapi.json — command tree (x-speakeasy-group / name),
//     help text, and which operations exist.
//   - the gr4vy-go package types — the exact method signatures the generated
//     code calls (so `go build` verifies the wiring).
package main

import (
	"fmt"
	"go/types"
	"os"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

const (
	rootPkg = "github.com/gr4vy/gr4vy-go"
	compPkg = rootPkg + "/models/components"
	opsPkg  = rootPkg + "/models/operations"

	specPath = "internal/spec/openapi.json"
	outPath  = "internal/commands/generated/zz_generated_commands.go"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}
}

func run() error {
	ops, err := parseSpec(specPath)
	if err != nil {
		return err
	}
	pkgs, err := loadPackages()
	if err != nil {
		return err
	}
	resources := buildResourceMap(pkgs[rootPkg])

	var gens []*genOp
	var skipped []string
	for _, op := range ops {
		res, ok := resources[op.Group]
		if !ok {
			skipped = append(skipped, fmt.Sprintf("%s.%s (no resource for group)", op.Group, op.Name))
			continue
		}
		m := findMethod(res.named, op.Name)
		if m == nil {
			skipped = append(skipped, fmt.Sprintf("%s.%s (no method)", op.Group, op.Name))
			continue
		}
		g, err := buildGenOp(op, res, m)
		if err != nil {
			skipped = append(skipped, fmt.Sprintf("%s.%s (%v)", op.Group, op.Name, err))
			continue
		}
		gens = append(gens, g)
	}

	sort.Slice(gens, func(i, j int) bool {
		if gens[i].Group != gens[j].Group {
			return gens[i].Group < gens[j].Group
		}
		return gens[i].Name < gens[j].Name
	})

	if err := emit(outPath, gens); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "generated %d/%d operations into %s\n", len(gens), len(ops), outPath)
	if len(skipped) > 0 {
		fmt.Fprintf(os.Stderr, "skipped %d operations:\n", len(skipped))
		for _, s := range skipped {
			fmt.Fprintln(os.Stderr, "  -", s)
		}
	}
	return nil
}

// --- gr4vy-go type loading & resource discovery ---

func loadPackages() (map[string]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedImports,
	}
	loaded, err := packages.Load(cfg, rootPkg, compPkg, opsPkg)
	if err != nil {
		return nil, fmt.Errorf("load gr4vy-go packages: %w", err)
	}
	out := map[string]*packages.Package{}
	var errs []string
	packages.Visit(loaded, nil, func(p *packages.Package) {
		out[p.PkgPath] = p
		for _, e := range p.Errors {
			errs = append(errs, e.Error())
		}
	})
	if out[rootPkg] == nil {
		return nil, fmt.Errorf("could not load %s (errors: %s)", rootPkg, strings.Join(errs, "; "))
	}
	return out, nil
}

type resourceInfo struct {
	goPath []string // e.g. ["Transactions","Refunds"]
	named  *types.Named
}

// buildResourceMap walks the Gr4vy struct, mapping each dotted-kebab resource
// path to its Go selector path and type.
func buildResourceMap(root *packages.Package) map[string]resourceInfo {
	out := map[string]resourceInfo{}
	gr, _ := root.Types.Scope().Lookup("Gr4vy").Type().(*types.Named)
	if gr == nil {
		return out
	}
	var recurse func(named *types.Named, goPath, kebab []string)
	recurse = func(named *types.Named, goPath, kebab []string) {
		st, ok := named.Underlying().(*types.Struct)
		if !ok {
			return
		}
		for i := 0; i < st.NumFields(); i++ {
			f := st.Field(i)
			if !f.Exported() {
				continue
			}
			ptr, ok := f.Type().(*types.Pointer)
			if !ok {
				continue
			}
			sub, ok := ptr.Elem().(*types.Named)
			if !ok || sub.Obj().Pkg() == nil || sub.Obj().Pkg().Path() != rootPkg {
				continue
			}
			newGo := append(append([]string{}, goPath...), f.Name())
			newKebab := append(append([]string{}, kebab...), pascalToKebab(f.Name()))
			key := strings.Join(newKebab, ".")
			if _, seen := out[key]; seen {
				continue
			}
			out[key] = resourceInfo{goPath: newGo, named: sub}
			recurse(sub, newGo, newKebab)
		}
	}
	recurse(gr, nil, nil)
	return out
}

func findMethod(named *types.Named, name string) *types.Func {
	want := normalizeName(name)
	ms := types.NewMethodSet(types.NewPointer(named))
	for i := 0; i < ms.Len(); i++ {
		fn, ok := ms.At(i).Obj().(*types.Func)
		if !ok || !fn.Exported() {
			continue
		}
		if normalizeName(fn.Name()) == want {
			return fn
		}
	}
	return nil
}

func normalizeName(s string) string {
	return strings.ToLower(strings.NewReplacer("_", "", "-", "").Replace(s))
}
