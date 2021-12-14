//go:build !dev
// +build !dev

package state

// Code generated by weld. DO NOT EDIT.

import (
	transitive_ops "example/transitive/ops"

	"github.com/luno/jettison/errors"
)

func MakeBackends() (Backends, error) {
	var (
		b   backendsImpl
		err error
	)

	b.foo, err = transitive_ops.NewFoo()
	if err != nil {
		return nil, errors.Wrap(err, "transitive ops new foo")
	}

	b.qux, err = transitive_ops.NewQux()
	if err != nil {
		return nil, errors.Wrap(err, "transitive ops new qux")
	}

	b.bar = transitive_ops.NewBar(b.foo)

	b.baz = transitive_ops.NewBaz(b.bar, b.qux)

	return &b, nil
}

type backendsImpl struct {
	foo transitive_ops.Foo
	qux transitive_ops.Qux
	bar transitive_ops.Bar
	baz transitive_ops.Baz
}

func (b *backendsImpl) Foo() transitive_ops.Foo {
	return b.foo
}

func (b *backendsImpl) Qux() transitive_ops.Qux {
	return b.qux
}

func (b *backendsImpl) Bar() transitive_ops.Bar {
	return b.bar
}

func (b *backendsImpl) Baz() transitive_ops.Baz {
	return b.baz
}
