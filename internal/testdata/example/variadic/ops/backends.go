package ops

import "testing"

type Backends interface {
	Foo() Foo
}

type options struct{}

type Option func(o *options)

type Foo struct {
}

func NewFooForTesting(t *testing.T, opts ...Option) (Foo, error) {
	return Foo{}, nil
}
