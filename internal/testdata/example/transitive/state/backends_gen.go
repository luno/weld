package state

// Code generated by weld. DO NOT EDIT.

import (
	transitive_ops "example/transitive/ops"
)

type Backends interface {
	Foo() transitive_ops.Foo
	Qux() transitive_ops.Qux
	Bar() transitive_ops.Bar
	Baz() transitive_ops.Baz
}
