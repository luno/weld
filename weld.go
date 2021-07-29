// Package weld contains directives for Luno style state and backends dependency injection
// using compile time code generation.
//
// Weld is heavily based on wire (github.com/google/wire), borrowing its syntax and concepts but tailoring
// it for Luno's specific state and backends based dependency injection pattern. Unlike wire, weld supports
// multiple providers for the same type by selecting the first provider found in the set using depth first search.
// Unlike wire, weld also supports transitive "backends-type" cyclic dependencies by adding these interfaces to
// the generated implementation.
//
// Unlike wire, weld is much less dynamic with fewer features, it takes a provider
// set as input and a backends as output and generates a Make function that returns
// an implementation of that backends interface.
//
// For convenience it can also generate an aggregate Backends interface from the union of a slice of
// backends since golang doesn't support embedding interfaces with the same name or overlapping methods.
//
// Relation to wire syntax:
//  Supported functions and types: wire.NewSet, wire.ProviderSet, wire.Bind, wire.Binding.
//  Unsupported functions and types: wire.Build, wire.Struct, wire.FieldsOf, wire.Value, wire.InterfaceValue.
//
// See the internal/testdata/example project for how this is used.
package weld

// NewSpec creates a weld specification that can generate a Make function
// for a backends interface from the provider set. The backends interface is
// a Luno style dependency provider interface with only pure getters.
//
// Note that NewSpec is not a wire concept, but it could be seen as replacing the wire.Build function.
//
// Given a spec, the weld command will generate a weld_gen.go file.
//
// Example:
//  // +build weld
//
//  package state
//
//  import (
//    ...
//  )
//
//  //go:generate weld
//
//  var _ = weld.NewSpec(
//     weld.NewSet(provider.WeldProd),
//     weld.GenUnion(new(exchange_ops.Backends), new(matcher_ops.Backends))
//
// Or for an alternative dev environment state:
//  var _ = weld.NewSpec(
//     weld.NewSet(provider.WeldDev),
//     weld.Existing(new(state.Backends))
func NewSpec(set ProviderSet, backends Backends) string {
	return "implementation not generated, run weld"
}

// ProviderSet is a marker type that collects a group of providers.
//
// This type is heavily based on github.com/google/wire.ProviderSet, see it for more details.
type ProviderSet struct{}

// NewSet creates a new provider set that includes the providers in its
// arguments. Each argument is a function value, a provider set or a call to
// Bind.
//
// This type is heavily based on github.com/google/wire.NewSet, see it for more details.
func NewSet(...interface{}) ProviderSet {
	return ProviderSet{}
}

// Binding maps an interface to a concrete type.
//
// This type is heavily based on github.com/google/wire.Binding, see it for more details.
type Binding struct{}

// Bind declares that a concrete type should be used to satisfy a dependency on
// the type of iface. iface must be a pointer to an interface type, to must be a
// pointer to a concrete type.
//
// This type is heavily based on github.com/google/wire.Bind, see it for more details.
//
// Example:
//
//  type Client interface {
//    List()
//  }
//
//  func New() *client {...}
//
//  type client struct{}
//
//  func (*client) List() {}
//
//  var Provider = weld.NewSet(New, weld.Bind(new(Client), new(*client)))
func Bind(iface, to interface{}) Binding {
	return Binding{}
}

// Backends is a marker type of the resulting generated backends interface;
// basically the output of weld.
//
// Note that Backends is not a wire concept.
type Backends struct{}

// GenUnion specifies that a new backends interface must be generated
// that is the union of the all the provided backends interfaces.
//
// Note that GenUnion is not a wire concept.
//
// Example:
//  var _ = weld.NewSpec(
//      providers.WeldProd,
//      weld.GenUnion(new(exchange_ops.Backends), new(matcher_ops.Backends))
func GenUnion(backends ...interface{}) Backends {
	return Backends{}
}

// Existing specifies that a single existing backends interface be used
// as output of weld.
//
// Note that Existing is not a wire concept.
//
// Example:
//  var _ = weld.NewSpec(
//      providers.WeldDev,
//      weld.Existing(new(state.Backends))
func Existing(backends interface{}) Backends {
	return Backends{}
}
