package providers

import "example/sort_with_bind"

type Foo struct {
}

func NewFoo() sort_with_bind.Foo {
	return new(Foo)
}

func (f *Foo) Foo() {
}

type Bar struct {
	f sort_with_bind.Foo
}

func NewBar(f sort_with_bind.Foo) *Bar {
	return &Bar{f}
}

func (b *Bar) Bar() {
}

type baz struct {
	b sort_with_bind.Bar
}

func NewBaz(b sort_with_bind.Bar) sort_with_bind.Baz {
	return &baz{b}
}

func (b *baz) Baz() {
}
