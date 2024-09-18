[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=luno_weld&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=luno_weld)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=luno_weld&metric=coverage)](https://sonarcloud.io/summary/new_code?id=luno_weld)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=luno_weld&metric=bugs)](https://sonarcloud.io/summary/new_code?id=luno_weld)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=luno_weld&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=luno_weld)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=luno_weld&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=luno_weld)
[![Duplicated Lines (%)](https://sonarcloud.io/api/project_badges/measure?project=luno_weld&metric=duplicated_lines_density)](https://sonarcloud.io/summary/new_code?id=luno_weld)
[![Go Report Card](https://goreportcard.com/badge/github.com/luno/weld)](https://goreportcard.com/report/github.com/luno/weld)
[![GoDoc](https://godoc.org/github.com/luno/weld?status.png)](https://godoc.org/github.com/luno/weld)

# Weld

Weld is a golang package that contains directives for "state and backends" dependency injection using compile time code generation.

Weld is heavily based on [github.com/google/wire](https://github.com/google/wire), borrowing its syntax and concepts, but tailoring it for backends-based dependency injection pattern. 

Unlike wire, weld:
- Supports multiple providers for the same type by selecting the first provider found in the set using depth-first search.
- Supports transitive "backends-type" cyclic dependencies by adding these interfaces to the generated implementation. 
- Is much less dynamic & has fewer features: it takes a provider set as input and a backends as output and generates a `Make` function that returns an implementation of that backends interface.

For convenience it can also generate an aggregate Backends interface from the union of a slice of backends since golang doesn't support embedding interfaces with the same name or overlapping methods.

## Relation to wire syntax:
- Supported functions and types:
  - `wire.Bind`
  - `wire.Binding`
  - `wire.NewSet`
  - `wire.ProviderSet`
- Unsupported functions and types: 
  - `wire.Build` 
  - `wire.FieldsOf`
  - `wire.InterfaceValue`
  - `wire.Struct`
  - `wire.Value`

## Example
- See the [internal/testdata/example](./internal/testdata/example) project for how this is used.
