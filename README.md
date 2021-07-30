# Weld

Weld is a golang package that contains directives for Luno style "state and backends" dependency injection
using compile time code generation.

Weld is heavily based on [wire](https://github.com/google/wire) (github.com/google/wire), borrowing its syntax and concepts but tailoring
it for Luno's specific state and backends based dependency injection pattern. Unlike wire, weld supports
multiple providers for the same type by selecting the first provider found in the set using depth first search.
Unlike wire, weld also supports transitive "backends-type" cyclic dependencies by adding these interfaces to
the generated implementation.

Unlike wire, weld is much less dynamic with fewer features, it takes a provider
set as input and a backends as output and generates a Make function that returns
an implementation of that backends interface.

For convenience it can also generate an aggregate Backends interface from the union of a slice of
backends since golang doesn't support embedding interfaces with the same name or overlapping methods.

Relation to wire syntax:
- Supported functions and types: wire.NewSet, wire.ProviderSet, wire.Bind, wire.Binding.
- Unsupported functions and types: wire.Build, wire.Struct, wire.FieldsOf, wire.Value, wire.InterfaceValue.

See the [internal/testdata/example](./internal/testdata/example) project for how this is used.

For more details, see the [weld godoc](https://pkg.go.dev/github.com/luno/weld).
