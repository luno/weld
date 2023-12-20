package internal

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"golang.org/x/tools/go/packages"
)

func Generate(ctx context.Context, args Args) (*Result, error) {
	logf(args, "Generating state for %s\n", args.InDir)
	logf(args, "Loading ast...")
	t0 := time.Now()

	pkg, err := load(ctx, args.InDir, args.Env, ".")
	if err != nil {
		return nil, err
	} else if len(pkg.Errors) > 0 {
		var res Result
		for _, e := range pkg.Errors {
			res.Errors = append(res.Errors, e)
		}
		return &res, nil
	}

	outPkg := pkg
	if args.InDir != args.OutDir {
		outPkg, err = load(ctx, args.OutDir, args.Env, ".")
		if err != nil {
			return nil, err
		}
	}

	logf(args, " done (%v)\n", time.Since(t0).Truncate(time.Millisecond))
	setPos(pkg) // Globals oh no :(
	logf(args, "Welding dependencies...")
	t0 = time.Now()

	graphExpr, bckExpr, err := findSpecParams(pkg)
	if err != nil {
		return nil, err
	}

	root, err := makeGraph(pkg, graphExpr)
	if err != nil {
		return nil, err
	}

	specBcks, genBcks, err := makeSpecBackends(pkg, bckExpr)
	if err != nil {
		return nil, err
	}

	selected, err := selectNodes(root, specBcks)
	if err != nil {
		return nil, err
	}

	tplData, err := makeTplData(pkg, outPkg, args.Tags, selected, specBcks)
	if err != nil {
		return nil, err
	}

	weldOut, err := execWeldTpl(tplData)
	if err != nil {
		return nil, err
	}

	bcksOut, err := maybeExecBackendsTpl(tplData, specBcks, genBcks)
	if err != nil {
		return nil, err
	}

	logf(args, " done (%v)\n", time.Since(t0).Truncate(time.Microsecond))

	if len(selected.UnselectedTypes) > 0 {
		logf(args, "%d unresolved dependency added to MakeBackends function\n", len(selected.UnselectedTypes))
		for _, t := range selected.UnselectedTypes {
			logf(args, "  - %s\n", t.String())
		}
	}

	return &Result{
		Root:           root,
		SpecBackends:   specBcks,
		SelectedNodes:  selected.SelectedNodes,
		TransBackends:  selected.TransitiveBackends,
		TplData:        tplData,
		WeldOutput:     weldOut,
		BackendsOutput: bcksOut,
	}, nil
}

func union(b Backends, bl []Backends) []BackendsDep {
	var res []BackendsDep
	uniq := make(map[string]types.Type)
	add := func(bl ...Backends) {
		for _, b := range bl {
			for _, dep := range b.Deps {
				if _, ok := uniq[dep.Getter]; ok {
					// TODO(corver): Check if types are identical.
					continue
				}
				uniq[dep.Getter] = dep.Type
				res = append(res, dep)
			}
		}
	}

	add(b)
	add(bl...)

	return res
}

type NodeSelection struct {
	// SelectedNodes is all the functions that need to be called in order
	// to create the weld state
	SelectedNodes []Node
	// TransitiveBackends is the other backends that needed to be satisfied by required nodes
	TransitiveBackends []Backends
	// UnselectedTypes is a list of parameters that could not be satisfied by dependencies
	UnselectedTypes []types.Type
}

// selectNodes returns the nodes required by the backends as well as
// any transitive backends found.
func selectNodes(root *Node, bcks Backends) (NodeSelection, error) {
	var ret NodeSelection
	selector := NewSelector(bcks)

	for !selector.Empty() {
		dep := selector.Pop()

		selectResult, ok, err := selectNode(root, dep)
		if err != nil {
			return NodeSelection{}, err
		} else if !ok {
			ret.UnselectedTypes = append(ret.UnselectedTypes, dep)
			continue
		}

		// If the selected node binds an implementation, add it.
		if selectResult.BindImpl != nil {
			selector.AddDep(selectResult.BindImpl)
		}

		// If selected node is a function, see if it has transitive backends or dep params, add them.
		if selectResult.Node.Type == NodeTypeFunc {
			sig := selectResult.Node.FuncSig
			if sig.Params().Len() == 1 && sig.Variadic() {
				// Just add variadic-only dependencies as is, assume it's functional options.
			} else {
				for _, p := range tupleSlice(sig.Params()) {
					if isBackends(p.Type()) {
						b, err := newBackends(p.Type(), selectResult.Node.FuncObj)
						if err != nil {
							return NodeSelection{}, err
						}
						selector.AddBackends(b, true)
						continue
					}
					selector.AddDep(p.Type())
				}
			}
		}

		ret.SelectedNodes = append(ret.SelectedNodes, *selectResult.Node)
	}
	ret.TransitiveBackends = selector.GetBackends()
	sort.Slice(ret.UnselectedTypes, func(i, j int) bool {
		tI, tJ := ret.UnselectedTypes[i], ret.UnselectedTypes[j]
		return len(tI.String()) < len(tJ.String())
	})
	return ret, nil
}

func selectNode(node *Node, dep types.Type) (SelectResult, bool, error) {
	switch node.Type {
	case NodeTypeSet:
		for _, child := range node.Children {
			sr, ok, err := selectNode(child, dep)
			if err != nil {
				return SelectResult{}, false, err
			} else if !ok {
				continue
			}
			return sr, true, nil
		}
		return SelectResult{}, false, nil
	case NodeTypeBind:
		if !types.Identical(node.BindInterface, dep) {
			return SelectResult{}, false, nil
		}
		return SelectResult{Node: node, BindImpl: node.BindImpl}, true, nil
	case NodeTypeFunc:
		if !types.Identical(node.FuncResult, dep) {
			return SelectResult{}, false, nil
		}
		return SelectResult{Node: node}, true, nil
	default:
		return SelectResult{}, false, errors.New("bug: unsupported node type")
	}
}

// makeSpecBackends returns the spec backends from backends expression.
func makeSpecBackends(pkg *packages.Package, expr ast.Expr) (Backends, bool, error) {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return Backends{}, false, errWithPos(expr, "weld.NewSpec second arg not a call expression")
	}

	se, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return Backends{}, false, errWithPos(expr, "weld.NewSpec second arg not a weld call expression")
	}

	get := func(expr ast.Expr, method string) (Backends, error) {
		call, ok := expr.(*ast.CallExpr)
		if !ok {
			return Backends{}, errWithPos(call, method+" only supports `new(pkg.Backends)` format")
		}
		if ident, ok := call.Fun.(*ast.Ident); !ok || ident.Name != "new" {
			return Backends{}, errWithPos(call, method+" only supports `new(pkg.Backends)` format")
		}

		b, err := makeBackends(pkg, call.Args[0])
		if err != nil {
			return Backends{}, err
		}

		return b, nil
	}

	if isImportedFuncCall(pkg.TypesInfo, se, "github.com/luno/weld", "Existing") {
		b, err := get(call.Args[0], "weld.Existing")
		return b, false, err
	}

	if !isImportedFuncCall(pkg.TypesInfo, se, "github.com/luno/weld", "GenUnion") {
		return Backends{}, false, errWithPos(expr, "unsupported weld.Backends function")
	}

	var bl []Backends
	for _, arg := range call.Args {
		b, err := get(arg, "weld.GenUnion")
		if err != nil {
			return Backends{}, false, err
		}

		bl = append(bl, b)
	}

	deps := union(Backends{}, bl)

	typ := types.NewNamed(
		types.NewTypeName(0, pkg.Types, "Backends", nil),
		new(types.Interface),
		nil)
	return Backends{
		Name:    "Backends",
		Type:    typ,
		Package: pkg.Types,
		Deps:    deps,
	}, true, nil
}

func makeBackends(pkg *packages.Package, expr ast.Expr) (Backends, error) {
	var ident *ast.Ident
	switch a := expr.(type) {
	case *ast.SelectorExpr:
		ident = a.Sel
	case *ast.Ident:
		ident = a
	default:
		return Backends{}, errWithPos(expr, "unsupported backends expression")
	}

	obj, ok := pkg.TypesInfo.Uses[ident]
	if !ok {
		return Backends{}, errWithPos(expr, "weld spec backends object not found")
	}

	b, err := newBackends(obj.Type(), obj)
	if err != nil {
		return Backends{}, err
	}

	return b, nil
}

func isBackends(typ types.Type) bool {
	return strings.HasSuffix(typ.String(), "Backends")
}

func newBackends(typ types.Type, parent haspos) (Backends, error) {
	if !isBackends(typ) {
		return Backends{}, errWithPos(parent, "weld spec argument not named .*Backends")
	}
	named, ok := typ.(*types.Named)
	if !ok {
		return Backends{}, errWithPos(parent, "weld spec Backends argument not an named interface")
	}

	iface, ok := typ.Underlying().(*types.Interface)
	if !ok {
		return Backends{}, errWithPos(parent, "weld spec Backends argument not an named interface")
	}

	var deps []BackendsDep
	for i := 0; i < iface.NumMethods(); i++ {
		meth := iface.Method(i)
		sig := meth.Type().(*types.Signature)
		if sig.Params().Len() > 0 {
			return Backends{}, errWithPos(meth, "unsupported weld spec Backends with parameterized method", j.MKV{"meth": meth})
		}
		result, err := getSigResult(sig)
		if err != nil {
			return Backends{}, errors.Wrap(err, pos(meth))
		}

		deps = append(deps, BackendsDep{Getter: meth.Name(), Type: result})
	}

	return Backends{
		Name:    named.Obj().Name(),
		Package: named.Obj().Pkg(),
		Type:    typ,
		Deps:    deps,
	}, nil
}

// makeGraph returns a graph of nodes representing all available providers by parsing
// the provider set tree.
func makeGraph(pkg *packages.Package, expr ast.Expr) (*Node, error) {
	switch e := expr.(type) {
	case *ast.CallExpr:
		typ, ok := getWeldFuncType(pkg.TypesInfo, e)
		if !ok {
			return nil, errWithPos(e, "unsupported non-weld func call")
		}
		if typ == NodeTypeBind {
			// Make bind type and return
			iface, ok := pkg.TypesInfo.Types[e.Args[0]]
			if !ok {
				return nil, errWithPos(e, "missing type info for bind interface")
			}
			impl, ok := pkg.TypesInfo.Types[e.Args[1]]
			if !ok {
				return nil, errWithPos(e, "missing type info for bind impl")
			}
			ifacePtr, ok := iface.Type.(*types.Pointer)
			if !ok {
				return nil, errWithPos(e, "weld.Bind requires pointers")
			}
			implPtr, ok := impl.Type.(*types.Pointer)
			if !ok {
				return nil, errWithPos(e, "weld.Bind requires pointers")
			}

			return &Node{
				Type:          typ,
				Deps:          []types.Type{ifacePtr.Elem()},
				BindInterface: ifacePtr.Elem(),
				BindImpl:      implPtr.Elem(),
			}, nil

		} else if typ == NodeTypeSet {
			n := Node{Type: typ}
			for _, arg := range e.Args {
				child, err := makeGraph(pkg, arg)
				if err != nil {
					return nil, err
				}
				n.AddChild(child)
			}

			return &n, nil

		} else {
			return nil, errWithPos(e, "unsupported weld function", j.MKV{"type": typ})
		}
	case *ast.SelectorExpr:
		// Provider set variable or func
		obj, ok := pkg.TypesInfo.Uses[e.Sel]
		if !ok {
			return nil, errWithPos(e, "selector object not found")
		}
		pkg, ok := pkg.Imports[obj.Pkg().Path()]
		if !ok {
			return nil, errWithPos(obj, "imported package not found", j.MKV{"pkg": obj.Pkg().Path()})
		}

		return makeObjectNode(pkg, obj)

	case *ast.Ident:
		obj, ok := pkg.TypesInfo.Uses[e]
		if !ok {
			return nil, errWithPos(e, "ident object not found")
		}
		return makeObjectNode(pkg, obj)
	default:
		return nil, errWithPos(e, "unsupported node expression type", j.MKV{"type": reflect.TypeOf(expr)})
	}
}

func makeObjectNode(pkg *packages.Package, obj types.Object) (*Node, error) {
	switch o := obj.(type) {
	case *types.Var:
		val, ok := varDecl(pkg, o)
		if !ok {
			return nil, errWithPos(o, "ident object not found")
		}
		if obj.Type().String() != "github.com/luno/weld.ProviderSet" {
			return nil, errWithPos(o, "unsupported provider object type, expect github.com/luno/weld.ProviderSet")
		}

		n := Node{
			Type:    NodeTypeSet,
			VarName: obj.Name(),
			VarPkg:  obj.Pkg().Path(),
		}
		for _, arg := range val.Values[0].(*ast.CallExpr).Args {
			child, err := makeGraph(pkg, arg)
			if err != nil {
				return nil, err
			}
			n.AddChild(child)
		}

		return &n, nil

	case *types.Func:

		sig := o.Type().(*types.Signature)
		res, err := getSigResult(sig)
		if err != nil {
			return nil, errors.Wrap(err, pos(o))
		}

		n := Node{
			Type:       NodeTypeFunc,
			FuncResult: res,
			FuncSig:    sig,
			FuncObj:    obj,
			Deps:       []types.Type{res},
		}

		return &n, nil
	default:
		return nil, errWithPos(o, "unsupported object type", j.MKV{"type": reflect.TypeOf(o)})
	}
}

func getSigResult(sig *types.Signature) (types.Type, error) {
	if sig.Results().Len() < 1 || sig.Results().Len() > 2 {
		return nil, errors.New("unsupported number of function results (not 1 or 2)")
	} else if sig.Results().Len() == 2 && sig.Results().At(1).Type().String() != "error" {
		return nil, errors.New("unsupported non-error 2nd function result")
	}
	return sig.Results().At(0).Type(), nil
}

func findSpecParams(pkg *packages.Package) (ast.Expr, ast.Expr, error) {
	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			if gd.Tok != token.VAR {
				continue
			}
			for _, spec := range gd.Specs {
				val, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				call, ok := val.Values[0].(*ast.CallExpr)
				if !ok {
					continue
				}
				se, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					continue
				}

				if isImportedFuncCall(pkg.TypesInfo, se, "github.com/luno/weld", "NewSpec") {
					return call.Args[0], call.Args[1], nil
				}
			}
		}
	}

	return nil, nil, errors.New("top level weld spec not found. Expect `var _ = weld.NewSpec(...)`")
}

// sortInDependencyOrder orders the dependencies such that transitive
// dependencies are only used after they've been created. For example, given
//
//	b.foo = MakeFoo(b.bar)
//
// we expect b.bar to be set before setting b.foo.
// Additionally, dependencies are sorted alphabetically so that the ordering
// is deterministic.
//
// TODO(neil): Optimise this. It's a little messy, and makes two copies of deps.
func sortInDependencyOrder(deps []BackendsDep, nodes []Node, paramTypes []types.Type) error {
	// Sort alphabetically first. We'll rearrange if there are dependencies, but
	// this is the ordering we start with.
	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Getter < deps[j].Getter
	})

	nodeDepsMap := getNodeDepsMap(nodes)

	q := make([]BackendsDep, 0, len(deps))
	for _, d := range deps {
		q = append(q, d)
	}

	var (
		result    = make([]BackendsDep, 0, len(q))
		resultMap = make(map[string]bool)
		i         = 0
		target    = len(deps)
	)
	for _, t := range paramTypes {
		resultMap[t.String()] = true
	}
	for len(q) > 0 {
		i++
		if i > target {
			var unresolved []string
			for _, d := range q {
				unresolved = append(unresolved, d.Getter)
			}
			return errors.New("unresolved dependency or dependency cycle", j.MKV{
				"unresolved": fmt.Sprintf("%v", unresolved),
				"resolution": "Make sure that at least one Backends provides each transitive dependency, and that there aren't any transitive dependency cycles",
			})
		}
		n := q[0]
		q = q[1:]
		allFound := true
		for k := range nodeDepsMap[n.Type.String()] {
			if !resultMap[k] {
				allFound = false
				break
			}
		}
		if allFound {
			result = append(result, n)
			resultMap[n.Type.String()] = true
			i = 0
			target = len(q)
		} else {
			q = append(q, n)
		}
	}

	for i := range result {
		deps[i] = result[i]
	}

	return nil
}

func getNodeDepsMap(nodes []Node) map[string]map[string]bool {
	// Build up a map of bind replacements. When we look up dependencies in
	// sortInDependencyOrder, we look up by the type the Backends provides,
	// not the type the providers provide. This map will help us map from the
	// latter to the former in the loop further down.
	bindMap := make(map[string]string)
	for _, n := range nodes {
		if n.Type == NodeTypeBind {
			bindMap[n.BindImpl.String()] = n.BindInterface.String()
		}
	}

	nodeMap := make(map[string]map[string]bool)
	for _, n := range nodes {
		if n.Type != NodeTypeFunc {
			continue
		}
		for i, p := range tupleSlice(n.FuncSig.Params()) {
			t := p.Type()
			if isBackends(t) {
				continue
			}

			// If the function is variadic and this is the last parameter we
			// can ignore it. We assume that variadic parameters are not
			// dependencies, but rather functional options.
			if i == n.FuncSig.Params().Len()-1 && n.FuncSig.Variadic() {
				continue
			}

			key := n.FuncResult.String()
			if k, ok := bindMap[key]; ok {
				key = k
			}

			if nodeMap[key] == nil {
				nodeMap[key] = make(map[string]bool)
			}
			nodeMap[key][t.String()] = true
		}
	}
	return nodeMap
}
