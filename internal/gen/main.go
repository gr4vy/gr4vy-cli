// Command gen reads the gr4vy-go SDK's Go types and emits one typed,
// cobra-backed command per SDK operation into internal/commands/generated. It
// is run via `go generate ./...` / `go run ./internal/gen` and is never
// compiled into the shipped binary.
//
// The SDK is the single source of truth: the command tree comes from the
// resource struct fields, the verbs from the method names, the flags/types from
// the method signatures (so `go build` verifies the wiring), and the help text
// from the methods' doc comments. There is intentionally no dependency on the
// OpenAPI spec — a typed CLI can only expose what gr4vy-go ships anyway.
package main

import (
	"fmt"
	"go/ast"
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

	outPath = "internal/commands/generated/zz_generated_commands.go"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}
}

func run() error {
	pkgs, err := loadPackages()
	if err != nil {
		return err
	}
	root := pkgs[rootPkg]
	resources := buildResourceMap(root)
	docs := extractDocs(root)

	// Deterministic resource order for stable output.
	groups := make([]string, 0, len(resources))
	for g := range resources {
		groups = append(groups, g)
	}
	sort.Strings(groups)

	var gens []*genOp
	var skipped []string
	total := 0
	for _, group := range groups {
		res := resources[group]
		for _, m := range operationMethods(res.named) {
			total++
			verb := kebab(m.Name())
			g, err := buildGenOp(group, verb, res, m, docs)
			if err != nil {
				skipped = append(skipped, fmt.Sprintf("%s.%s (%v)", group, verb, err))
				continue
			}
			gens = append(gens, g)
		}
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

	fmt.Fprintf(os.Stderr, "generated %d/%d operations into %s\n", len(gens), total, outPath)
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
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedDeps | packages.NeedImports | packages.NeedSyntax,
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
	if len(errs) > 0 {
		// Fail fast: generating from packages that didn't type-check cleanly
		// would silently produce an incomplete/incorrect command surface.
		return nil, fmt.Errorf("gr4vy-go packages did not load cleanly: %s", strings.Join(errs, "; "))
	}
	if out[rootPkg] == nil {
		return nil, fmt.Errorf("could not load %s", rootPkg)
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

// operationMethods returns the resource's exported operation methods — those
// whose first parameter is a context.Context — in deterministic order.
func operationMethods(named *types.Named) []*types.Func {
	var out []*types.Func
	ms := types.NewMethodSet(types.NewPointer(named))
	for i := 0; i < ms.Len(); i++ {
		fn, ok := ms.At(i).Obj().(*types.Func)
		if !ok || !fn.Exported() {
			continue
		}
		sig, ok := fn.Type().(*types.Signature)
		if !ok || sig.Params().Len() == 0 || !isContext(sig.Params().At(0).Type()) {
			continue
		}
		out = append(out, fn)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

// extractDocs maps "<ReceiverType>.<Method>" to the method's doc comment, read
// from the root package's syntax trees.
func extractDocs(root *packages.Package) map[string]string {
	docs := map[string]string{}
	for _, file := range root.Syntax {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv == nil || fd.Doc == nil || len(fd.Recv.List) == 0 {
				continue
			}
			recv := recvTypeName(fd.Recv.List[0].Type)
			if recv == "" {
				continue
			}
			docs[recv+"."+fd.Name.Name] = strings.TrimSpace(fd.Doc.Text())
		}
	}
	return docs
}

func recvTypeName(e ast.Expr) string {
	if star, ok := e.(*ast.StarExpr); ok {
		e = star.X
	}
	if id, ok := e.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}
