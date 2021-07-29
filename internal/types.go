package internal

import (
	"go/types"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

type Args struct {
	WorkDir string
	Env     []string
	Verbose bool
	Tags    string
}

type Result struct {
	Root           *Node
	SpecBackends   Backends
	TransBackends  []Backends
	SelectedNodes  []Node
	Errors         []error
	TplData        *TplData
	WeldOutput     []byte
	BackendsOutput []byte
}

//go:generate stringer -type=NodeType -trimprefix=NodeType

type NodeType int

const (
	NodeTypeUnknown NodeType = iota
	NodeTypeSet              // Type=weld.ProviderSet, value=weld.NewSet
	NodeTypeFunc             // Type=func return type, value=func literal
	NodeTypeBind             // Type=weld.Binding,     value=weld.Bind
)

// Node represents provider in a provider set graph.
type Node struct {
	// Type of the node.
	Type NodeType

	// Parent of this node in the provider set graph (nil if root).
	Parent *Node

	// Children of this node in the provider set graph.
	Children []*Node

	// Deps is descendents' provided types.
	Deps []types.Type

	// HasDups indicates if this node has duplicate providers.
	HasDups bool

	// VarPkg is the fully qualified package if the provider set is assigned to a variable.
	VarPkg string

	// VarName is the variable name if the provider set is assigned to a variable.
	VarName string

	FuncObj    types.Object
	FuncSig    *types.Signature
	FuncResult types.Type

	BindInterface types.Type
	BindImpl      types.Type
}

func (n *Node) AddChild(child *Node) {
	n.Children = append(n.Children, child)
	n.Deps = append(n.Deps, child.Deps...)
	child.Parent = n

	uniq := make(TypeMap)
	for _, dep := range n.Deps {
		if uniq.Put(dep) {
			n.HasDups = true
			break
		}
	}
}

// Backends represents a Backends interface.
type Backends struct {
	// Name of the backends interface, usually Backends.
	Name string
	// Type is the actual Backends interface type itself.
	Type types.Type

	// Package of the backends.
	Package *types.Package

	// Deps is the list of dependencies provided by the backends.
	Deps []BackendsDep
}

type BackendsDep struct {
	Getter string
	Type   types.Type
}

type SelectResult struct {
	Node     *Node
	BindImpl types.Type
}

type Selector struct {
	added []Backends
	deps  []types.Type
	uniq  TypeMap
}

func (s *Selector) Empty() bool {
	return len(s.deps) == 0
}

func (s *Selector) Pop() types.Type {
	dep := s.deps[0]
	s.deps = s.deps[1:]
	return dep
}

func (s *Selector) GetBackends() []Backends {
	return s.added
}

func (s *Selector) AddBackends(b Backends, add bool) {
	var deps []types.Type
	for _, dep := range b.Deps {
		deps = append(deps, dep.Type)
	}

	sort.Slice(deps, func(i, j int) bool {
		return deps[i].String() < deps[j].String()
	})

	for _, dep := range deps {
		s.AddDep(dep)
	}

	if add {
		s.added = append(s.added, b)
	}
}

func (s *Selector) AddDep(dep types.Type) {
	if s.uniq.Put(dep) {
		return
	}
	s.deps = append(s.deps, dep)
}

func NewSelector(bcks Backends) *Selector {
	s := &Selector{
		uniq: make(TypeMap),
	}

	s.AddBackends(bcks, false)

	return s
}

type TplData struct {
	// Package is the generated source package name.
	Package string

	// Tags is the build tags string
	Tags string

	BackendsType string // Backends or ops.Backends
	BackendsName string // Backends

	// Imports are the imported packages by package path.
	Imports   map[string]TplImport
	Deps      []TplDep
	TransBcks []string //  _ email_logical.Backends = (*backendsImpl)(nil)
}

type TplDep struct {
	Type     string //  *email_db.EmailDB
	Var      string //  emailDB
	Getter   string //  EmailDB
	Provider TplProvider
}

func (d TplDep) FormatVar() string {
	return "b." + d.Var
}

type TplProvider struct {
	FuncPkg    string //  email_db
	FuncName   string //  Connect
	ReturnsErr bool   //  s.email, err = email_db.Connect()
	TakesBcks  bool   //  email_db.Connect(&b)
	ErrWrapMsg string //  return nil, errors.Wrap(err, "email db connect")
}

func (p TplProvider) FormatFunc() string {
	if p.FuncPkg == "" {
		return p.FuncName
	}
	return p.FuncPkg + "." + p.FuncName
}

func (p TplProvider) FormatArgs() string {
	if p.TakesBcks {
		return "&b"
	}
	return ""
}

type TplImport struct {
	Name    string
	PkgPath string
	Aliased bool
}

// PkgCache manages the packages used in code generation providing
// imports including aliases and type references.
type PkgCache struct {
	// Local package path being code generated.
	Local string

	// Pkgs is a map of imported packages by path
	Pkgs map[string]TplImport

	// Names is map of packages paths by import alias/name.
	Names map[string]string
}

func (c *PkgCache) Name(pkg *types.Package) string {
	return c.Pkgs[pkg.Path()].Name
}

func (c *PkgCache) Add(pkg *types.Package) {
	if pkg == nil {
		return
	}

	pkgPath := pkg.Path()

	if c.Local == pkgPath {
		// No need to import local package.
		return
	} else if _, ok := c.Pkgs[pkgPath]; ok {
		// Already added this package.
		return
	}

	//  maybeAdd returns true if it could add the package,
	// otherwise it returns false since the name clashed.
	maybeAdd := func(pkgPath, name string) bool {
		if _, ok := c.Names[name]; ok {
			return false
		}

		c.Pkgs[pkgPath] = TplImport{
			PkgPath: pkgPath,
			Name:    name,
			Aliased: name != filepath.Base(pkgPath),
		}
		c.Names[name] = pkgPath
		return true
	}

	// First try with our "smart" alias
	alias := smartAlias(pkgPath)

	for {
		name := strings.Join(alias, "_")
		if ok := maybeAdd(pkgPath, name); ok {
			break
		}

		// Else try with an additional folder.
		folders := strings.Split(pkgPath, "/")
		from := len(folders) - len(alias) - 1
		alias = folders[from:]
	}
}

func NewPkgCache(local *packages.Package) *PkgCache {
	return &PkgCache{
		Local: local.PkgPath,
		Pkgs:  make(map[string]TplImport),
		Names: map[string]string{local.Name: local.PkgPath},
	}
}
