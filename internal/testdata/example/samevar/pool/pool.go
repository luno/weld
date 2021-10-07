package pool

type FooPool struct {
}

func NewFooPool() *FooPool {
	return new(FooPool)
}

type BarPool struct {
}

func NewBarPool() *BarPool {
	return new(BarPool)
}
