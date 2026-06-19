package main

import (
	"fmt"
	"go/types"
	"reflect"
	"strings"
)

type flagMeta struct {
	Name  string
	Usage string
	Kind  string // commands.FlagKind constant name, e.g. "KindString"
}

type genOp struct {
	Group, Name string
	Short, Long string
	PathParams  []string
	HasBody     bool
	BodyType    string
	IsList      bool
	Optionals   []flagMeta
	Body        string // Go statements for the Run closure
}

// buildGenOp classifies a method signature and produces the closure body and
// command metadata.
func buildGenOp(op specOp, res resourceInfo, m *types.Func) (*genOp, error) {
	sig := m.Type().(*types.Signature)
	core := coreParams(sig)
	selector := "c." + strings.Join(res.goPath, ".") + "." + m.Name()

	g := &genOp{
		Group:  op.Group,
		Name:   op.Name,
		Short:  short(op),
		Long:   strings.TrimSpace(op.Description),
		IsList: op.Name == "list",
	}

	if len(core) == 1 {
		if named, ok := requestStruct(core[0].Type()); ok {
			return buildRequestStructOp(g, selector, named, sig)
		}
	}
	return buildFlattenedOp(g, selector, core, sig)
}

// buildFlattenedOp handles methods with individual positional/optional params.
func buildFlattenedOp(g *genOp, selector string, core []*types.Var, sig *types.Signature) (*genOp, error) {
	var pre []string
	callArgs := []string{"ctx"}
	argIdx := 0

	for _, p := range core {
		pt := p.Type()
		name := p.Name()
		switch {
		case isBasicString(pt):
			callArgs = append(callArgs, fmt.Sprintf("in.Args[%d]", argIdx))
			g.PathParams = append(g.PathParams, kebab(name))
			argIdx++
		case isBodyParam(pt):
			decl, isPtr, typeName, _ := bodyParam(pt)
			g.HasBody = true
			g.BodyType = typeName
			pre = append(pre,
				"var body "+decl,
				"if len(in.Body) > 0 {",
				"if err := json.Unmarshal(in.Body, &body); err != nil {",
				"return nil, err",
				"}",
				"}",
			)
			callArgs = append(callArgs, ptrPrefix(isPtr)+"body")
		case isStringSlice(pt):
			fn := kebab(name)
			pre = append(pre, fmt.Sprintf("%s := commands.StringSlice(in.Flags, %q)", name, fn))
			callArgs = append(callArgs, name)
			g.Optionals = append(g.Optionals, flagMeta{Name: fn, Usage: usageFor(fn), Kind: "KindStringSlice"})
		default:
			stmt, expr, fm, ok := scalarArg(name, pt)
			if !ok {
				return nil, fmt.Errorf("unsupported param %s of type %s", name, pt.String())
			}
			if stmt != "" {
				pre = append(pre, stmt)
			}
			callArgs = append(callArgs, expr)
			if fm.Name != "merchant-account-id" {
				g.Optionals = append(g.Optionals, fm)
			}
		}
	}

	call := selector + "(" + strings.Join(callArgs, ", ") + ")"
	g.Body = strings.Join(pre, "\n") + "\n" + returnBlock(call, sig)
	return g, nil
}

// scalarArg emits the prelude (if any), the call argument expression, and flag
// metadata for an optional scalar/pointer parameter.
func scalarArg(name string, t types.Type) (prelude, expr string, fm flagMeta, ok bool) {
	fn := kebab(name)
	elem, isPtr := derefPtr(t)
	if !isPtr {
		return "", "", flagMeta{}, false
	}
	if b, isBasic := elem.(*types.Basic); isBasic {
		switch b.Kind() {
		case types.String:
			return "", fmt.Sprintf("commands.OptString(in.Flags, %q)", fn), flagMeta{fn, usageFor(fn), "KindString"}, true
		case types.Int64:
			return "", fmt.Sprintf("commands.OptInt64(in.Flags, %q)", fn), flagMeta{fn, usageFor(fn), "KindInt64"}, true
		case types.Int:
			return "", fmt.Sprintf("commands.OptInt(in.Flags, %q)", fn), flagMeta{fn, usageFor(fn), "KindInt"}, true
		case types.Bool:
			return "", fmt.Sprintf("commands.OptBool(in.Flags, %q)", fn), flagMeta{fn, usageFor(fn), "KindBool"}, true
		case types.Float64, types.Float32:
			return "", fmt.Sprintf("commands.OptFloat64(in.Flags, %q)", fn), flagMeta{fn, usageFor(fn), "KindFloat"}, true
		}
	}
	if n, isNamed := elem.(*types.Named); isNamed {
		if b, isBasic := n.Underlying().(*types.Basic); isBasic && b.Kind() == types.String {
			qual := pkgQualifier(n)
			prelude = fmt.Sprintf("var %s *%s.%s\nif v := commands.OptString(in.Flags, %q); v != nil { e := %s.%s(*v); %s = &e }",
				name, qual, n.Obj().Name(), fn, qual, n.Obj().Name(), name)
			return prelude, name, flagMeta{fn, usageFor(fn), "KindString"}, true
		}
	}
	return "", "", flagMeta{}, false
}

func ptrPrefix(isPtr bool) string {
	if isPtr {
		return "&"
	}
	return ""
}

// buildRequestStructOp handles methods that take a single operations.*Request.
func buildRequestStructOp(g *genOp, selector string, named *types.Named, sig *types.Signature) (*genOp, error) {
	reqType := named.Obj().Name()
	st, ok := named.Underlying().(*types.Struct)
	if !ok {
		return nil, fmt.Errorf("%s is not a struct", reqType)
	}
	pre := []string{"req := operations." + reqType + "{}"}
	argIdx := 0

	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		if !f.Exported() {
			continue
		}
		tag := reflect.StructTag(st.Tag(i))
		field := f.Name()
		switch {
		case tag.Get("pathParam") != "":
			name := tagName(tag.Get("pathParam"))
			pre = append(pre, fmt.Sprintf("req.%s = in.Args[%d]", field, argIdx))
			g.PathParams = append(g.PathParams, snakeToKebab(name))
			argIdx++
		case tag.Get("request") != "":
			decl, isPtr, typeName, _ := bodyParam(f.Type())
			g.HasBody = true
			g.BodyType = typeName
			if isPtr {
				// Only allocate the pointer body when one was actually
				// provided, so optional bodies stay nil (omitted).
				pre = append(pre,
					"if len(in.Body) > 0 {",
					fmt.Sprintf("req.%s = new(%s)", field, strings.TrimPrefix(decl, "*")),
					fmt.Sprintf("if err := json.Unmarshal(in.Body, req.%s); err != nil {", field),
					"return nil, err", "}", "}",
				)
			} else {
				pre = append(pre,
					"if len(in.Body) > 0 {",
					fmt.Sprintf("if err := json.Unmarshal(in.Body, &req.%s); err != nil {", field),
					"return nil, err", "}", "}",
				)
			}
		case tag.Get("header") != "":
			hname := tagName(tag.Get("header"))
			switch {
			case hname == "x-gr4vy-merchant-account-id":
				pre = append(pre, fmt.Sprintf("req.%s = commands.OptString(in.Flags, %q)", field, "merchant-account-id"))
			case isStringSlice(f.Type()):
				pre = append(pre, fmt.Sprintf("req.%s = commands.StringSlice(in.Flags, %q)", field, hname))
				g.Optionals = append(g.Optionals, flagMeta{Name: hname, Usage: usageFor(hname), Kind: "KindStringSlice"})
			default:
				pre = append(pre, fmt.Sprintf("req.%s = commands.OptString(in.Flags, %q)", field, hname))
				g.Optionals = append(g.Optionals, flagMeta{Name: hname, Usage: usageFor(hname), Kind: "KindString"})
			}
		case tag.Get("queryParam") != "":
			qname := snakeToKebab(tagName(tag.Get("queryParam")))
			stmt, fm, ok := queryField(field, qname, f.Type())
			if !ok {
				continue // complex query param: not exposed as a flag
			}
			pre = append(pre, stmt)
			g.Optionals = append(g.Optionals, fm)
		}
	}

	call := selector + "(ctx, req)"
	g.Body = strings.Join(pre, "\n") + "\n" + returnBlock(call, sig)
	return g, nil
}

// queryField emits the assignment and flag metadata for a query parameter, or
// reports ok=false for an unsupported (complex) type.
func queryField(field, flag string, t types.Type) (string, flagMeta, bool) {
	if isStringSlice(t) {
		return fmt.Sprintf("req.%s = commands.StringSlice(in.Flags, %q)", field, flag),
			flagMeta{Name: flag, Usage: usageFor(flag), Kind: "KindStringSlice"}, true
	}
	elem, isPtr := derefPtr(t)
	if !isPtr {
		return "", flagMeta{}, false
	}
	if b, ok := elem.(*types.Basic); ok {
		switch b.Kind() {
		case types.String:
			return fmt.Sprintf("req.%s = commands.OptString(in.Flags, %q)", field, flag),
				flagMeta{Name: flag, Usage: usageFor(flag), Kind: "KindString"}, true
		case types.Int64:
			return fmt.Sprintf("req.%s = commands.OptInt64(in.Flags, %q)", field, flag),
				flagMeta{Name: flag, Usage: usageFor(flag), Kind: "KindInt64"}, true
		case types.Int:
			return fmt.Sprintf("req.%s = commands.OptInt(in.Flags, %q)", field, flag),
				flagMeta{Name: flag, Usage: usageFor(flag), Kind: "KindInt"}, true
		case types.Bool:
			return fmt.Sprintf("req.%s = commands.OptBool(in.Flags, %q)", field, flag),
				flagMeta{Name: flag, Usage: usageFor(flag), Kind: "KindBool"}, true
		case types.Float64, types.Float32:
			return fmt.Sprintf("req.%s = commands.OptFloat64(in.Flags, %q)", field, flag),
				flagMeta{Name: flag, Usage: usageFor(flag), Kind: "KindFloat"}, true
		}
	}
	// Named enum with a string underlying type.
	if n, ok := elem.(*types.Named); ok {
		if b, ok := n.Underlying().(*types.Basic); ok && b.Kind() == types.String {
			qual := pkgQualifier(n)
			stmt := fmt.Sprintf("if v := commands.OptString(in.Flags, %q); v != nil { e := %s.%s(*v); req.%s = &e }",
				flag, qual, n.Obj().Name(), field)
			return stmt, flagMeta{Name: flag, Usage: usageFor(flag), Kind: "KindString"}, true
		}
	}
	return "", flagMeta{}, false
}

// returnBlock emits the call plus result handling based on the method's results.
func returnBlock(call string, sig *types.Signature) string {
	results := sig.Results()
	var valueResult types.Type
	for i := 0; i < results.Len(); i++ {
		if !isError(results.At(i).Type()) {
			valueResult = results.At(i).Type()
			break
		}
	}
	if valueResult == nil {
		return "return nil, " + call
	}
	return "resp, err := " + call + "\n" +
		"if err != nil {\nreturn nil, err\n}\n" +
		"return " + resultExpr(valueResult) + ", nil"
}

// resultExpr renders the payload: an operations response wrapper exposes its
// body via .Result; otherwise the value is rendered directly.
func resultExpr(t types.Type) string {
	elem, _ := derefPtr(t)
	if n, ok := elem.(*types.Named); ok && pkgPath(n) == opsPkg && hasField(n, "Result") {
		return "resp.Result"
	}
	return "resp"
}

// --- type helpers ---

func coreParams(sig *types.Signature) []*types.Var {
	params := sig.Params()
	n := params.Len()
	var core []*types.Var
	for i := 0; i < n; i++ {
		p := params.At(i)
		if i == n-1 && sig.Variadic() {
			continue
		}
		if isContext(p.Type()) {
			continue
		}
		core = append(core, p)
	}
	return core
}

func derefPtr(t types.Type) (types.Type, bool) {
	if p, ok := t.(*types.Pointer); ok {
		return p.Elem(), true
	}
	return t, false
}

func isContext(t types.Type) bool { return types.TypeString(t, nil) == "context.Context" }

func isError(t types.Type) bool { return types.TypeString(t, nil) == "error" }

func isBasicString(t types.Type) bool {
	b, ok := t.(*types.Basic)
	return ok && b.Kind() == types.String
}

func isPtrString(t types.Type) bool {
	elem, isPtr := derefPtr(t)
	return isPtr && isBasicString(elem)
}

func isStringSlice(t types.Type) bool {
	s, ok := t.(*types.Slice)
	return ok && isBasicString(s.Elem())
}

func isBodyParam(t types.Type) bool {
	_, _, _, ok := bodyParam(t)
	return ok
}

// bodyParam reports whether t is a request body: a component type, an
// operations body union (not a *Request), or a freeform map. It returns the Go
// type expression to declare, whether the parameter is a pointer, and a short
// type name for help text.
func bodyParam(t types.Type) (declType string, isPtr bool, typeName string, ok bool) {
	elem, ptr := derefPtr(t)
	switch e := elem.(type) {
	case *types.Named:
		pp := pkgPath(e)
		name := e.Obj().Name()
		if pp == compPkg || (pp == opsPkg && !strings.HasSuffix(name, "Request")) {
			return types.TypeString(elem, qualifier), ptr, name, true
		}
	case *types.Map:
		return types.TypeString(elem, qualifier), ptr, "", true
	}
	return "", false, "", false
}

// qualifier renders package-qualified type names with the short aliases used in
// the generated file.
func qualifier(p *types.Package) string {
	if p == nil {
		return ""
	}
	switch p.Path() {
	case compPkg:
		return "components"
	case opsPkg:
		return "operations"
	case rootPkg:
		return "gr4vygo"
	}
	return p.Name()
}

func requestStruct(t types.Type) (*types.Named, bool) {
	n, ok := t.(*types.Named)
	if ok && pkgPath(n) == opsPkg && strings.HasSuffix(n.Obj().Name(), "Request") {
		return n, true
	}
	return nil, false
}

func pkgPath(n *types.Named) string {
	if n.Obj().Pkg() == nil {
		return ""
	}
	return n.Obj().Pkg().Path()
}

func pkgQualifier(n *types.Named) string {
	switch pkgPath(n) {
	case opsPkg:
		return "operations"
	case compPkg:
		return "components"
	default:
		return "components"
	}
}

func hasField(n *types.Named, name string) bool {
	st, ok := n.Underlying().(*types.Struct)
	if !ok {
		return false
	}
	for i := 0; i < st.NumFields(); i++ {
		if st.Field(i).Name() == name && st.Field(i).Exported() {
			return true
		}
	}
	return false
}

// tagName extracts the name= value from a Speakeasy struct tag value such as
// "style=form,explode=true,name=cursor".
func tagName(v string) string {
	for _, part := range strings.Split(v, ",") {
		if strings.HasPrefix(part, "name=") {
			return strings.TrimPrefix(part, "name=")
		}
	}
	return ""
}

func short(op specOp) string {
	if op.Summary != "" {
		return op.Summary
	}
	return capitalize(op.Name) + " " + strings.ReplaceAll(op.Group, ".", " ")
}

var usageOverrides = map[string]string{
	"idempotency-key": "unique key to make the request idempotent",
	"x-forwarded-for": "originating client IP address",
	"prefer":          "preferred response resource type",
	"cursor":          "pagination cursor",
	"limit":           "maximum number of items to return",
	"search":          "free-text search filter",
}

func usageFor(flag string) string {
	if u, ok := usageOverrides[flag]; ok {
		return u
	}
	return flag + " parameter"
}
