package state

type OneBackends interface {
	GetFoo() FooSomething
}

type TwoBackends interface {
	Foo() FooSomething
}

type FooSomething struct {
	Bar string
}

func FooProvider() FooSomething {
	return FooSomething{"42"}
}
