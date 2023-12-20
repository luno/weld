//go:build weld

package state

import (
	"context"

	"github.com/luno/weld"
)

//go:generate weld -tags=!dev
// Note that github.com/luno/weld/internal/gen_test.go generates this specific state_gen.go as well.

var _ = weld.NewSpec(
	weld.NewSet(MakeFizz, MakeFoo),
	weld.GenUnion(new(Backends)),
)

type Foo struct {
	Cocktail int
}

type Bar interface {
	MakeMeACocktail() int
}

func MakeFoo(b Bar) Foo {
	return Foo{Cocktail: b.MakeMeACocktail()}
}

type Fizz struct{}

func MakeFizz(context.Context) (Fizz, error) {
	return Fizz{}, nil
}

type Backends interface {
	GetFizz() Fizz
	GetFoo() Foo
}
