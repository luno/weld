package ops

type Backends interface {
	Foo() Foo
	Bar() Bar
	Baz() Baz
	Qux() Qux
}

// This is the dependency tree for the providers below:
//
//       baz
//        |
//    +---+---+
//    |       |
//   bar     qux
//    |
//   foo
//
// We expect to construct the state in this  order: foo, qux, bar, baz.
// foo and qux can be provided in any order because they're leaves in the tree.

type Foo struct {
}

func NewFoo() (Foo, error) {
	return Foo{}, nil
}

type Bar struct {
	foo Foo
}

func NewBar(foo Foo) Bar {
	return Bar{foo: foo}
}

type Baz struct {
	bar Bar
	qux Qux
}

func NewBaz(bar Bar, qux Qux) Baz {
	return Baz{bar: bar, qux: qux}
}

type Qux struct {
}

func NewQux() (Qux, error) {
	return Qux{}, nil
}
