package internal

import (
	"bytes"
	_ "embed"
	"go/format"
	"go/types"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

//go:embed templates/weld.tmpl
var weldTplText string

//go:embed templates/bcks.tmpl
var bcksTplText string

//go:embed templates/testing.tmpl
var testingTplText string

var (
	weldTpl    = template.Must(template.New("").Parse(weldTplText))
	bcksTpl    = template.Must(template.New("").Parse(bcksTplText))
	testingTpl = template.Must(template.New("").Parse(testingTplText))
)

// execWeldTpl returns the generated source of the template data.
func execWeldTpl(data *TplData) ([]byte, error) {
	var buf bytes.Buffer
	err := weldTpl.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	imports.LocalPrefix = "bitx"
	src, err := imports.Process("weld_gen.go", buf.Bytes(), nil)
	if err != nil {
		return nil, err
	}

	src, err = format.Source(src)
	if err != nil {
		return nil, err
	}

	return src, nil
}

func execTestingTpl(data *TplData) ([]byte, error) {
	var buf bytes.Buffer
	err := testingTpl.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	imports.LocalPrefix = "bitx"
	src, err := imports.Process("testing_gen.go", buf.Bytes(), nil)
	if err != nil {
		return nil, err
	}

	src, err = format.Source(src)
	if err != nil {
		return nil, err
	}

	return src, nil
}

func maybeExecBackendsTpl(tplData *TplData, bcks Backends, genBcks bool) ([]byte, error) {
	if !genBcks {
		return nil, nil
	}

	// Remove transitive deps
	var deps []TplDep
	for _, dep := range tplData.Deps {
		var found bool
		for _, bdep := range bcks.Deps {
			if dep.Getter == bdep.Getter {
				found = true
				break
			}
		}
		if found {
			deps = append(deps, dep)
		}
	}

	clone := *tplData
	clone.Deps = deps

	var buf bytes.Buffer
	err := bcksTpl.Execute(&buf, clone)
	if err != nil {
		return nil, err
	}

	imports.LocalPrefix = "bitx"
	src, err := imports.Process("backends_gen.go", buf.Bytes(), nil)
	if err != nil {
		return nil, err
	}

	src, err = format.Source(src)
	if err != nil {
		return nil, err
	}

	return src, nil
}

// makeTplData returns the template data for backends and selected nodes.
func makeTplData(in, out *packages.Package, tags string, selected NodeSelection, specBcks Backends) (*TplData, error) {
	pkgCache := NewPkgCache(in, out)
	pkgCache.Add(specBcks.Package)

	unionDeps := union(specBcks, selected.TransitiveBackends)
	err := sortInDependencyOrder(unionDeps, selected.SelectedNodes, selected.UnselectedTypes)
	if err != nil {
		return nil, errors.Wrap(err, "error sorting in dependency order")
	}

	// TODO(neil): The deps are now sorted mostly alphabetically, but also in
	// dependency order. It would be nice if the Backends interface was just
	// sorted alphabetically.

	// We do a first pass to figure out what the var name for each dependency is.
	// This is needed to correctly construct transitive dependencies.
	varMap := make(map[string]string)
	for _, dep := range unionDeps {
		varMap[dep.Type.String()] = "b." + getter2Var(dep.Getter)
	}
	for _, param := range selected.UnselectedTypes {
		v, err := type2Param(param)
		if err != nil {
			return nil, err
		}
		varMap[param.String()] = v
	}

	var deps []TplDep
	// NOTE: union dedupes on the getter name. That ensures we don't have
	// multiple getters in the Backends interface with the same name. But when
	// we create backendsImpl, we additionally need to dedupe on the var name.
	// For example, Foo and GetFoo are different getters, but the fields they
	// access would both be called foo. That would lead to a duplicate field in
	// the backendsImpl struct.
	//
	// Additionally, if the fields they access have the same name but different
	// types, we include both but append a sequence number to the field name to
	// avoid conflicts.
	uniqVars := make(map[string]string)
	for _, dep := range unionDeps {
		d, err := makeTplDep(pkgCache, selected.SelectedNodes, dep.Getter, dep.Type, varMap)
		if err != nil {
			return nil, err
		}

		orig := d.Var
		n := 1
		for {
			prev := uniqVars[d.Var]
			if prev == "" {
				uniqVars[d.Var] = d.Type
				break
			}
			if prev == d.Type {
				d.IsDuplicate = true
				break
			}
			d.Var = orig + strconv.Itoa(n)
			n++
		}

		deps = append(deps, *d)
	}

	tb, err := makeTplBcks(pkgCache, selected.TransitiveBackends)
	if err != nil {
		return nil, err
	}

	bcksTypeRef, err := makeTypeRef(pkgCache, specBcks.Type)
	if err != nil {
		return nil, err
	}

	params, err := makeParams(pkgCache, selected.UnselectedTypes)
	if err != nil {
		return nil, err
	}

	return &TplData{
		Package:      out.Name,
		Tags:         tags,
		BackendsName: specBcks.Name,
		BackendsType: bcksTypeRef,
		Imports:      pkgCache.Pkgs,
		Params:       params,
		Deps:         deps,
		TransBcks:    tb,
	}, nil
}

// makeTplBcks returns the backends type references and template imports of the backends.
func makeTplBcks(pkgCache *PkgCache, bcks []Backends) ([]string, error) {
	var res []string
	for _, b := range bcks {
		if err := addTypeImports(pkgCache, b.Type); err != nil {
			return nil, err
		}

		ref, err := makeTypeRef(pkgCache, b.Type)
		if err != nil {
			return nil, err
		}

		res = append(res, ref)
	}

	return res, nil
}

// makeTplDep returns the template dependency and template imports of the dependency.
func makeTplDep(pkgCache *PkgCache, nodes []Node, getter string, dep types.Type, varMap map[string]string) (*TplDep, error) {
	for _, node := range nodes {
		if node.Type == NodeTypeBind {
			if !types.Identical(node.BindInterface, dep) {
				continue
			}

			if err := addTypeImports(pkgCache, node.BindInterface); err != nil {
				return nil, err
			}

			implDep, err := makeTplDep(pkgCache, nodes, getter, node.BindImpl, varMap)
			if err != nil {
				return nil, err
			}

			typeRef, err := makeTypeRef(pkgCache, node.BindInterface)
			if err != nil {
				return nil, err
			}

			return &TplDep{
				Type:     typeRef,
				Var:      getter2Var(getter),
				Getter:   getter,
				Provider: implDep.Provider,
			}, nil

		} else if node.Type == NodeTypeFunc {
			if !types.Identical(node.FuncResult, dep) {
				continue
			}

			if err := addObjImports(pkgCache, node.FuncObj); err != nil {
				return nil, err
			}

			typeRef, err := makeTypeRef(pkgCache, node.FuncResult)
			if err != nil {
				return nil, errors.Wrap(err, "",
					j.MKV{"node": node.FuncResult})
			}

			providerFuncName := node.FuncObj.Name()

			params, err := getParams(node.FuncSig, varMap)
			if err != nil {
				return nil, errors.Wrap(err, "error getting params")
			}

			return &TplDep{
				Type:   typeRef,
				Var:    getter2Var(getter),
				Getter: getter,
				Provider: TplProvider{
					FuncPkg:    pkgCache.Name(node.FuncObj.Pkg()),
					FuncName:   providerFuncName,
					ReturnsErr: returnsErr(node.FuncSig),
					Params:     params,
					ErrWrapMsg: makeWrapMsg(node.FuncObj.Pkg().Path(), providerFuncName),
				},
			}, nil
		} else {
			return nil, errors.New("unsupported node type")
		}
	}

	return nil, errors.New("dep no found", j.MKV{"getter": getter, "type": dep})
}

// makeTypeRef return the type reference; including package alias or without package if input package.
func makeTypeRef(pkgCache *PkgCache, typ types.Type) (string, error) {
	pkgs, err := getTypePkgs(typ)
	if err != nil {
		return "", err
	}

	res := typ.String()

	for _, pkg := range pkgs {
		if pkg.Path() == pkgCache.In {
			res = strings.ReplaceAll(res, pkg.Path()+".", "")
			continue
		}

		name := pkgCache.Name(pkg)
		if name == "" {
			return "", errors.New("import not found for type", j.MKV{"pkg": pkg})
		}
		res = strings.ReplaceAll(res, pkg.Path(), name)
	}

	return res, nil
}

func makeParams(pkgCache *PkgCache, unselected []types.Type) ([]TplParam, error) {
	var params []TplParam
	for _, p := range unselected {
		v, err := type2Param(p)
		if err != nil {
			return nil, err
		}
		t, err := makeTypeRef(pkgCache, p)
		if err != nil {
			return nil, err
		}
		params = append(params, TplParam{
			Name: v, Type: t,
		})
	}
	return params, nil
}

// takesBcks returns true if the function signature takes a Backends as parameter.
func takesBcks(sig *types.Signature) bool {
	for _, p := range tupleSlice(sig.Params()) {
		n, ok := p.Type().(*types.Named)
		if !ok {
			continue
		}
		if _, ok := n.Underlying().(*types.Interface); !ok {
			continue
		}
		if strings.HasSuffix(n.String(), "Backends") {
			return true
		}
	}
	return false
}

// returnsErr returns true if the function signature returns an error.
func returnsErr(sig *types.Signature) bool {
	for _, r := range tupleSlice(sig.Results()) {
		if r.Type() == types.Universe.Lookup("error").Type() {
			return true
		}
	}
	return false
}

// makeWrapMsg returns an error wrap message for a func call.
func makeWrapMsg(pkg string, funcName string) string {
	s := strings.Join(smartAlias(pkg), " ")
	if s != "" {
		s += " "
	}
	s += splitCamel(funcName)
	return s
}

// splitCamel returns the camel case name as space separated.
func splitCamel(name string) string {
	var res []rune
	chars := []rune(name)
	for _, char := range chars {
		lower := unicode.ToLower(char)
		if len(res) > 0 && char != lower {
			res = append(res, ' ')
		}
		res = append(res, lower)
	}
	return string(res)
}

// getter2Var returns the variable name for a getter function.
// It strips any "Get" prefix and lowercases the first letter.
func getter2Var(getter string) string {
	getter = strings.TrimPrefix(getter, "Get") // Warning: This might cause variable to clash...
	chars := []rune(getter)
	chars[0] = unicode.ToLower(chars[0])
	return string(chars)
}

func type2Param(t types.Type) (string, error) {
	if point, ok := t.(*types.Pointer); ok {
		return type2Param(point.Elem())
	}

	named, ok := t.(*types.Named)
	if !ok {
		return "", errors.New("cannot construct param var with unnamed parameter type", j.KV("param_type", t.String()))
	}
	return strings.ToLower(named.Obj().Name()), nil
}

// smartAlias returns a "unique" import alias for a package as slice of path sections.
// Note this is very luno specific.
func smartAlias(pkgPath string) []string {
	more := map[string]bool{
		"client":  true,
		"ops":     true,
		"server":  true,
		"db":      true,
		"grpc":    true,
		"logical": true,
		"dev":     true,
		"state":   true,
	}

	pkgParts := strings.Split(pkgPath, "/")

	var res []string
	for i := len(pkgParts) - 1; i >= 0; i-- {
		part := strings.ReplaceAll(pkgParts[i], "-", "")
		if strings.HasPrefix(part, "v") {
			_, err := strconv.Atoi(part[1:])
			if err != nil {
				// NoReturnErr: We're checking if this is a number.
			} else {
				// This is a versioned import. Include the version as well as
				// the next part.
				res = append([]string{part}, res...)
				continue
			}
		}
		res = append([]string{part}, res...)
		if part == "bitx" || !more[part] {
			break
		}
	}
	return res
}

// addObjImports adds the imports related to the object to the cache.
func addObjImports(pkgCache *PkgCache, obj types.Object) error {
	pkgs, err := getObjPkgs(obj)
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		pkgCache.Add(pkg)
	}

	return nil
}

// addTypeImports adds the imports related to the types to the cache.
func addTypeImports(pkgCache *PkgCache, typ types.Type) error {
	pkgs, err := getTypePkgs(typ)
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		pkgCache.Add(pkg)
	}

	return nil
}

// getObjPkgs returns the packages related to the types.
func getObjPkgs(objs ...types.Object) ([]*types.Package, error) {
	var res []*types.Package
	for _, obj := range objs {
		var (
			pl  []*types.Package
			err error
		)
		switch o := obj.(type) {
		case *types.Builtin:
			continue
		case *types.Const:
			pl, err = getTypePkgs(o.Type())
		case *types.Var:
			pl, err = getTypePkgs(o.Type())
		case *types.Func:
			pl, err = getTypePkgs(o.Type())
			pl = append(pl, o.Pkg())
		default:
			return nil, errors.New("detect imports unsupported object", j.MKV{"obj": o})
		}

		if err != nil {
			return nil, err
		}

		res = append(res, pl...)
	}

	return res, nil
}

// tupleTypes returns the types of the tuple values.
func tupleTypes(tuple *types.Tuple) []types.Type {
	var res []types.Type
	for i := 0; i < tuple.Len(); i++ {
		v := tuple.At(i)
		res = append(res, v.Type())
	}
	return res
}

// getTypePkgs returns the packages related to the types.
func getTypePkgs(tl ...types.Type) ([]*types.Package, error) {
	var res []*types.Package
	for _, typ := range tl {
		var (
			pl  []*types.Package
			err error
		)
		switch t := typ.(type) {
		case *types.Basic:
			continue
		case *types.Struct:
			continue
		case *types.Slice:
			pl, err = getTypePkgs(t.Elem())
		case *types.Array:
			pl, err = getTypePkgs(t.Elem())
		case *types.Pointer:
			pl, err = getTypePkgs(t.Elem())
		case *types.Chan:
			pl, err = getTypePkgs(t.Elem())
		case *types.Named:
			pkg := t.Obj().Pkg()
			if pkg != nil {
				pl = []*types.Package{pkg}
			}
		case *types.Map:
			pl, err = getTypePkgs(t.Elem(), t.Key())
		case *types.Signature:
			tl := tupleTypes(t.Params())
			tl = append(tl, tupleTypes(t.Results())...)
			pl, err = getTypePkgs(tl...)
		default:
			return nil, errors.New("cannot detect import for type", j.MKV{"type": t})
		}

		if err != nil {
			return nil, err
		}

		res = append(res, pl...)
	}

	return res, nil
}

// getParams returns the string representations of the parameters that should be
// passed to a function provider. For example, given the weld output:
//
//	b.dep, err := MakeDep(&b)
//	if err != nil {
//	  return nil, err
//	}
//
// this function returns []string{"&b"}.
func getParams(typ *types.Signature, varMap map[string]string) (params []string, err error) {
	for i := 0; i < typ.Params().Len(); i++ {
		p := typ.Params().At(i)
		if isBackends(p.Type()) {
			params = append(params, "&b")
			continue
		}

		isLast := typ.Params().Len()-1 == i
		if isLast && typ.Variadic() {
			// TODO(neil): Variadic parameters are often used for options.
			// We skip those for now.
			continue
		}

		v := varMap[p.Type().String()]
		if v == "" {
			return nil, errors.New("param type not found in var map", j.MKV{
				"param_type": p.Type().String(),
				"sig":        typ.String(),
			})
		}
		params = append(params, v)
	}
	return params, nil
}
