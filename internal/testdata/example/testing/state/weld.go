//go:build weld

package state

import "github.com/luno/weld"

//go:generate weld -tags=!dev -testing
// Note that github.com/luno/weld/internal/gen_test.go generates this specific state_gen.go as well.

var _ = weld.NewSpec(
	weld.NewSet(MakeFoo),
	weld.GenUnion(new(Backends)),
)

type Backends interface {
	GetFoo() Foo
}

type Foo struct{}

func MakeFoo() Foo {
	return Foo{}
}
