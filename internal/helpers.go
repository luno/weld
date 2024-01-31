package internal

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

// load typechecks the given packages and all transitive dependencies.
//
// wd is the working directory and env is the set of environment
// variables to use when loading the specified packages. If env is nil
// or empty, it is interpreted as an empty set of variables. In case of
// duplicate environment variables, the last one in the list takes precedence.
func load(ctx context.Context, wd string, env []string, pkgs ...string) ([]*packages.Package, error) {
	mode := packages.NeedSyntax | packages.NeedImports | packages.NeedTypes |
		packages.NeedTypesInfo | packages.NeedDeps | packages.NeedName |
		packages.NeedModule
	cfg := &packages.Config{
		Context:    ctx,
		Mode:       mode,
		Dir:        wd,
		Env:        env,
		BuildFlags: []string{"-tags=weld"},
	}

	loadedPackages, err := packages.Load(cfg, pkgs...)
	if err != nil {
		return nil, err
	} else if len(loadedPackages) != len(pkgs) {
		return nil, errors.New("unexpected number of packages loaded", j.MKV{
			"want": len(pkgs),
			"got":  len(loadedPackages),
		})
	}

	return loadedPackages, nil
}

// varDecl finds the declaration that defines the given variable.
//
// Note this was copied from: github.com/google/wire@v0.4.0/internal/wire/parse.go
func varDecl(pkg *packages.Package, obj *types.Var) (*ast.ValueSpec, bool) {
	pos := obj.Pos()
	for _, f := range pkg.Syntax {
		tokenFile := pkg.Fset.File(f.Pos())
		if base := tokenFile.Base(); base <= int(pos) && int(pos) < base+tokenFile.Size() {
			path, _ := astutil.PathEnclosingInterval(f, pos, pos)
			for _, node := range path {
				if spec, ok := node.(*ast.ValueSpec); ok {
					return spec, true
				}
			}
		}
	}
	return nil, false
}

// getImportedFunc returns the package, function name and true of the selector expression.
// It returns false if the selector expression is not an imported function call.
// Note pkg is the full import path not the aliased or short form.
//
// Note this was copied from: bitx/tools/dev/lunovet/vetutil/vetutil.go
func getImportedFunc(info *types.Info, se *ast.SelectorExpr) (pkg string, fn string, ok bool) {
	pkgID, ok := se.X.(*ast.Ident)
	if !ok {
		return "", "", false
	}

	object := info.Uses[pkgID]
	pkgName, ok := object.(*types.PkgName)
	if !ok {
		return "", "", false
	}

	pkg = pkgName.Imported().Path()
	fn = se.Sel.Name
	return pkg, fn, true
}

// isImportedFuncCall returns true if the selector expression is a call to pkg.fn().
// Note pkg should be the full import path not the aliased or short form.
//
// Note this was copied from: bitx/tools/dev/lunovet/vetutil/vetutil.go
func isImportedFuncCall(info *types.Info, se *ast.SelectorExpr, pkg, fn string) bool {
	p, f, ok := getImportedFunc(info, se)
	return ok && pkg == p && fn == f
}

// getWeldFuncType returns the node type for a weld package
// function call.
func getWeldFuncType(info *types.Info, call *ast.CallExpr) (NodeType, bool) {
	se, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return NodeTypeUnknown, false
	}

	p, f, ok := getImportedFunc(info, se)
	if !ok {
		return NodeTypeUnknown, false
	}
	if p != "github.com/luno/weld" {
		return NodeTypeUnknown, false
	}

	switch f {
	case "NewSet":
		return NodeTypeSet, true
	case "Bind":
		return NodeTypeBind, true
	default:
		return NodeTypeUnknown, false
	}
}

// TypeMap provides a map of types.
type TypeMap map[string]types.Type

// Put adds the typ to the map and returns true if it already existed.
func (m TypeMap) Put(typ types.Type) bool {
	if _, ok := m[typ.String()]; ok {
		return true
	}
	m[typ.String()] = typ
	return false
}

func logf(args Args, format string, a ...interface{}) {
	if !args.Verbose {
		return
	}
	fmt.Printf(format, a...)
}

type haspos interface {
	Pos() token.Pos
}

// pos is a global position formatting function. It should only be used for
// improving error messages since globals are bad practice.
var pos = func(p haspos) string { return "not available yet" }

// setPos set a pos function using the
// provided package's fset to format positions.
func setPos(pkg *packages.Package) {
	pos = func(p haspos) string {
		return pkg.Fset.Position(p.Pos()).String()
	}
}

// errWithPos returns a error with a position prefix. It uses a global
// pos function to resolve the position string. Sorry :(
func errWithPos(p haspos, msg string, opts ...errors.Option) error {
	msg = pos(p) + ": " + msg
	return errors.New(msg, opts...)
}

func tupleSlice(t *types.Tuple) (res []*types.Var) {
	for i := 0; i < t.Len(); i++ {
		res = append(res, t.At(i))
	}
	return res
}
